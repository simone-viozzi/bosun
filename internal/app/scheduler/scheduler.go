// Package scheduler implements the core cron-based job scheduling engine.
//
// The Scheduler manages job registration, cron-based execution, overlap policies,
// circuit-breaking, and state persistence. It uses robfig/cron/v3 for time-based
// scheduling and delegates actual job execution to the ports.JobExecutor interface.
//
// Key responsibilities:
//   - Parse cron expressions and register jobs with the cron library
//   - Apply overlap policies (queue, skip) via cron job wrappers
//   - Execute jobs through JobExecutor with global semaphore control
//   - Track job status (idle, running, completed, failed)
//   - Implement circuit-breaker: auto-disable after 3 consecutive failures
//   - Persist execution state via JobStateStore
//   - Emit lifecycle events via EventEmitter
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/semaphore"

	"github.com/simone-viozzi/bosun/internal/app/concurrency"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// circuitBreakerThreshold is the number of consecutive failures before a job
// is automatically disabled. Once circuit-broken, the job is NOT re-enabled
// by config refresh — manual intervention is required.
const circuitBreakerThreshold = 3

// entry tracks a registered cron job.
type entry struct {
	job     jobs.Job
	entryID cron.EntryID
	// circuitBroken is true if the job has been auto-disabled.
	circuitBroken bool
}

// Options configures scheduler behavior.
type Options struct {
	// Parallelism sets the global semaphore weight (max concurrent jobs).
	// Default: 1 (serial execution).
	Parallelism int64

	// RefreshInterval is the interval between config refresh cycles.
	// Zero or negative disables the refresh loop.
	RefreshInterval time.Duration

	// DiscoverFn is called periodically to discover the current set of jobs.
	// If nil, config refresh is disabled.
	DiscoverFn DiscoverFunc
}

// Scheduler manages cron-based job scheduling with execution, overlap handling,
// circuit breaking, and state persistence.
type Scheduler struct {
	cron       *cron.Cron
	executor   ports.JobExecutor
	events     ports.EventEmitter
	stateStore ports.JobStateStore
	stackLocks *concurrency.StackLockManager
	globalSem  *semaphore.Weighted
	logger     *slog.Logger
	opts       Options

	// mu protects entries and statusMap.
	mu        sync.RWMutex
	entries   map[string]*entry         // jobName -> entry
	statusMap map[string]jobs.JobStatus // jobName -> current status

	// wg tracks running jobs for graceful shutdown.
	wg sync.WaitGroup
}

// New creates a Scheduler with the given dependencies.
func New(
	executor ports.JobExecutor,
	events ports.EventEmitter,
	stateStore ports.JobStateStore,
	opts Options,
	logger *slog.Logger,
) *Scheduler {
	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = 1
	}

	return &Scheduler{
		cron:       cron.New(cron.WithSeconds()),
		executor:   executor,
		events:     events,
		stateStore: stateStore,
		stackLocks: concurrency.NewStackLockManager(),
		globalSem:  semaphore.NewWeighted(parallelism),
		logger:     logger,
		opts:       opts,
		entries:    make(map[string]*entry),
		statusMap:  make(map[string]jobs.JobStatus),
	}
}

// AddJob registers a job with the cron scheduler.
// It parses the cron expression, applies the overlap policy wrapper,
// and emits a JobScheduled event.
func (s *Scheduler) AddJob(ctx context.Context, job jobs.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine effective overlap policy.
	policy := job.OverlapPolicy
	if policy == "" {
		policy = jobs.DefaultOverlapPolicy
	}

	// Build the cron job function.
	jobFunc := s.makeJobFunc(job)

	// Apply overlap policy wrapper.
	var wrappedJob cron.Job
	switch policy {
	case jobs.OverlapPolicySkip:
		wrappedJob = s.skipIfRunning(job.Name)(jobFunc)
	case jobs.OverlapPolicyQueue:
		wrappedJob = cron.DelayIfStillRunning(cron.DefaultLogger)(jobFunc)
	default:
		// Unknown policy — treat as queue (safe default).
		wrappedJob = cron.DelayIfStillRunning(cron.DefaultLogger)(jobFunc)
	}

	// Parse and register with cron.
	entryID, err := s.cron.AddJob(job.Schedule, wrappedJob)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q for job %q: %w", job.Schedule, job.Name, err)
	}

	// Track entry.
	s.entries[job.Name] = &entry{
		job:     job,
		entryID: entryID,
	}

	// Initialize status.
	cronEntry := s.cron.Entry(entryID)
	s.statusMap[job.Name] = jobs.JobStatus{
		JobName:       job.Name,
		Status:        jobs.ScheduleStatusIdle,
		Schedule:      job.Schedule,
		OverlapPolicy: policy,
		NextRunTime:   cronEntry.Next,
	}

	// Emit event.
	s.events.EmitJobScheduled(ctx, job.Name, job.Schedule, cronEntry.Next)

	return nil
}

// RemoveJob unregisters a job from the scheduler.
// Any currently running execution will complete normally.
func (s *Scheduler) RemoveJob(ctx context.Context, jobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[jobName]
	if !ok {
		return fmt.Errorf("job %q not found", jobName)
	}

	s.cron.Remove(e.entryID)
	delete(s.entries, jobName)
	delete(s.statusMap, jobName)

	s.events.EmitJobRemoved(ctx, jobName)

	return nil
}

// executeJob runs a single job execution. It acquires the global semaphore,
// calls the executor, updates status and state, and implements the circuit breaker.
func (s *Scheduler) executeJob(ctx context.Context, job jobs.Job) {
	runID := uuid.New().String()

	// Acquire global semaphore.
	if err := s.globalSem.Acquire(ctx, 1); err != nil {
		s.logger.WarnContext(ctx, "failed to acquire semaphore",
			slog.String("job", job.Name),
			slog.String("error", err.Error()),
		)
		return
	}
	defer s.globalSem.Release(1)

	// Track for graceful shutdown.
	s.wg.Add(1)
	defer s.wg.Done()

	// Acquire per-stack locks (sorted order, deadlock-free).
	if len(job.TargetStacks) > 0 {
		if err := s.stackLocks.LockAll(ctx, job.TargetStacks); err != nil {
			s.logger.WarnContext(ctx, "failed to acquire stack locks",
				slog.String("job", job.Name),
				slog.String("error", err.Error()),
			)
			return
		}
		defer s.stackLocks.UnlockAll(job.TargetStacks)
	}

	// Update status to running.
	s.setRunning(job.Name, runID)

	// Emit started event.
	s.events.EmitJobStarted(ctx, job.Name, runID)

	start := time.Now()

	// Execute the job.
	result, err := s.executor.Execute(ctx, job, ports.ExecuteOptions{})

	duration := time.Since(start)
	now := time.Now()

	if err != nil {
		// Execution failed.
		s.events.EmitJobFailed(ctx, job.Name, runID, err, duration)
		s.setFailed(job.Name, runID, err.Error(), &now)
		s.handleFailure(ctx, job.Name, err.Error(), &now)
	} else if result.Run.Status == jobs.RunStatusFailed {
		// Non-zero exit code or step failure.
		errMsg := result.Run.Error
		if errMsg == "" {
			errMsg = "worker failed"
		}
		s.events.EmitJobFailed(ctx, job.Name, runID, fmt.Errorf("%s", errMsg), duration)
		s.setFailed(job.Name, runID, errMsg, &now)
		s.handleFailure(ctx, job.Name, errMsg, &now)
	} else {
		// Success.
		s.events.EmitJobCompleted(ctx, job.Name, runID, duration)
		s.setCompleted(job.Name, &now)
		s.handleSuccess(ctx, job.Name, &now)
	}
}

// makeJobFunc returns a cron.FuncJob that executes the given job.
func (s *Scheduler) makeJobFunc(job jobs.Job) cron.FuncJob {
	return func() {
		s.executeJob(context.Background(), job)
	}
}

// skipIfRunning returns a cron.JobWrapper that skips the job if it's still
// running from a previous invocation, AND emits a JobSkipped event (T033).
// This replaces the stock cron.SkipIfStillRunning so we get observability.
func (s *Scheduler) skipIfRunning(jobName string) cron.JobWrapper {
	return func(j cron.Job) cron.Job {
		var running int32 // 0 = idle, 1 = running
		return cron.FuncJob(func() {
			if !atomic.CompareAndSwapInt32(&running, 0, 1) {
				s.events.EmitJobSkipped(context.Background(), jobName, "overlap-policy=skip: previous run still active")
				s.logger.InfoContext(context.Background(), "job skipped: still running",
					slog.String("job", jobName),
				)
				return
			}
			defer atomic.StoreInt32(&running, 0)
			j.Run()
		})
	}
}

// setRunning updates the status map to reflect a running job.
func (s *Scheduler) setRunning(jobName, runID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status, ok := s.statusMap[jobName]; ok {
		status.Status = jobs.ScheduleStatusRunning
		status.CurrentRunID = runID
		s.statusMap[jobName] = status
	}
}

// setCompleted updates the status map to reflect a completed job.
func (s *Scheduler) setCompleted(jobName string, completedAt *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status, ok := s.statusMap[jobName]; ok {
		status.Status = jobs.ScheduleStatusCompleted
		status.CurrentRunID = ""
		status.LastRunTime = completedAt
		status.LastResult = "success"
		s.statusMap[jobName] = status
	}
}

// setFailed updates the status map to reflect a failed job.
func (s *Scheduler) setFailed(jobName, _, errMsg string, failedAt *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status, ok := s.statusMap[jobName]; ok {
		status.Status = jobs.ScheduleStatusFailed
		status.CurrentRunID = ""
		status.LastRunTime = failedAt
		status.LastResult = "error: " + errMsg
		s.statusMap[jobName] = status
	}
}

// handleSuccess resets the circuit-breaker counter and persists state.
func (s *Scheduler) handleSuccess(ctx context.Context, jobName string, completedAt *time.Time) {
	s.mu.Lock()
	e, ok := s.entries[jobName]
	if ok {
		e.circuitBroken = false
	}
	s.mu.Unlock()

	// Persist state.
	_ = s.stateStore.SaveJobState(ctx, ports.JobState{
		JobName:             jobName,
		LastRunAt:           completedAt,
		LastResult:          "success",
		ConsecutiveFailures: 0,
	})
}

// handleFailure increments the circuit-breaker counter, persists state,
// and auto-disables the job if the threshold is reached.
func (s *Scheduler) handleFailure(ctx context.Context, jobName, errMsg string, failedAt *time.Time) {
	// Load current state to get consecutive failure count.
	state, err := s.stateStore.LoadJobState(ctx, jobName)
	if err != nil {
		// First failure or state not found — start fresh.
		state = ports.JobState{
			JobName: jobName,
		}
	}

	state.ConsecutiveFailures++
	state.LastRunAt = failedAt
	state.LastResult = "error: " + errMsg

	_ = s.stateStore.SaveJobState(ctx, state)

	// Circuit-breaker check.
	if state.ConsecutiveFailures >= circuitBreakerThreshold {
		s.mu.Lock()
		if e, ok := s.entries[jobName]; ok {
			e.circuitBroken = true
			s.cron.Remove(e.entryID)
		}
		s.mu.Unlock()

		s.events.EmitJobCircuitBroken(ctx, jobName, state.ConsecutiveFailures)
		s.logger.ErrorContext(ctx, "job circuit-broken: auto-disabled after consecutive failures",
			slog.String("job", jobName),
			slog.Int("consecutive_failures", state.ConsecutiveFailures),
		)
	}
}

// Start begins the cron scheduler. It blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) error {
	s.cron.Start()
	s.logger.InfoContext(ctx, "scheduler started")

	// T062: Start config refresh loop if configured.
	if s.opts.DiscoverFn != nil && s.opts.RefreshInterval > 0 {
		go s.startRefreshLoop(ctx, s.opts.RefreshInterval, s.opts.DiscoverFn)
	}

	// Block until context is cancelled.
	<-ctx.Done()

	return s.Stop(ctx)
}

// Stop gracefully shuts down the scheduler.
// It stops the cron scheduler and waits for all running jobs to complete.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.logger.InfoContext(ctx, "scheduler stopping")

	// Stop scheduling new jobs.
	cronCtx := s.cron.Stop()

	// Wait for cron's internal processing to finish.
	<-cronCtx.Done()

	// Wait for all running jobs to complete.
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.InfoContext(ctx, "scheduler stopped: all jobs completed")
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "scheduler stopped: timeout waiting for running jobs")
	}

	return nil
}

// ListJobs returns current status of all scheduled jobs.
// Thread-safe for concurrent access.
func (s *Scheduler) ListJobs() []jobs.JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]jobs.JobStatus, 0, len(s.statusMap))
	for name, status := range s.statusMap {
		// Update NextRunTime from cron entry.
		if e, ok := s.entries[name]; ok {
			cronEntry := s.cron.Entry(e.entryID)
			status.NextRunTime = cronEntry.Next
		}
		result = append(result, status)
	}

	return result
}

// IsCircuitBroken returns whether a job has been auto-disabled by the circuit breaker.
func (s *Scheduler) IsCircuitBroken(jobName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.entries[jobName]; ok {
		return e.circuitBroken
	}
	return false
}

// StartCron starts the underlying cron scheduler without blocking.
// This is useful for tests that need to trigger cron-based execution
// without the blocking Start() method.
func (s *Scheduler) StartCron() {
	s.cron.Start()
}

// StopCron stops the underlying cron scheduler without waiting for running jobs.
// This is the non-blocking counterpart to Stop().
func (s *Scheduler) StopCron() {
	s.cron.Stop()
}
