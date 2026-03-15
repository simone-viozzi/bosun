# Quickstart: Implementing M4 Scheduling Engine

**Date**: 2026-02-15
**For**: Developers implementing M4 features
**Prerequisites**: Go 1.24+, Docker, familiarity with Bosun M3 architecture

## Overview

This guide walks through implementing the M4 scheduling engine step-by-step. Follow the task order in [tasks.md](tasks.md) for proper dependency management.

## Quick Links

- **Spec**: [spec.md](spec.md) - User stories and requirements
- **Plan**: [plan.md](plan.md) - Architecture and technical approach
- **Research**: [research.md](research.md) - Concurrency model and design decisions
- **Data Model**: [data-model.md](data-model.md) - Types, interfaces, and data flow
- **Tasks**: [tasks.md](tasks.md) - Detailed task breakdown

## Phase 1: Foundation (Week 1)

### Day 1: App Bootstrap (#168)

**Goal**: Create service factory that wires all dependencies.

**Files to create**:
- `internal/app/app.go`

**Steps**:
1. Define `Services` struct with all port fields:
   ```go
   type Services struct {
       Planner      ports.JobPlanner
       Discoverer   ports.JobDiscoverer
       Executor     ports.JobExecutor
       Scheduler    *scheduler.Scheduler  // nil if not daemon mode
       EventEmitter ports.EventEmitter
       StackLocks   *concurrency.StackLockManager
       GlobalSem    *semaphore.Weighted
       DockerClient *docker.Client
   }
   ```

2. Implement `Bootstrap(ctx, opts)` function:
   ```go
   func Bootstrap(ctx context.Context, opts BootstrapOptions) (*Services, error) {
       // 1. Create Docker client
       dockerClient, err := docker.NewClientFromEnv()
       if err != nil {
           return nil, fmt.Errorf("failed to create Docker client: %w", err)
       }

       // 2. Instantiate adapters
       labelSource := dockerlabels.New(dockerClient)
       composeCtrl := compose.New(dockerClient)
       workerRunner := worker.New(dockerClient)
       eventEmitter := events.NewLogEmitter(slog.Default())
       stateStore := state.NewInMemoryStateStore() // M4: in-memory; #177 swaps in BoltStateStore

       // 3. Create app services
       planner := planner.New(labelSource)
       discoverer := planner  // same instance
       stackLocks := concurrency.NewStackLockManager()
       globalSem := semaphore.NewWeighted(int64(opts.Parallelism))

       executor := executor.New(planner, composeCtrl, workerRunner, dockerClient,
                                stackLocks, globalSem, eventEmitter)

       // 4. Create scheduler if daemon mode
       var sched *scheduler.Scheduler
       if opts.DaemonMode {
           sched = scheduler.New(executor, discoverer, eventEmitter, stateStore, scheduler.Options{
               RefreshInterval: opts.RefreshInterval,
           })
       }

       return &Services{
           Planner:      planner,
           Discoverer:   discoverer,
           Executor:     executor,
           Scheduler:    sched,
           EventEmitter: eventEmitter,
           StateStore:   stateStore,
           StackLocks:   stackLocks,
           GlobalSem:    globalSem,
           DockerClient: dockerClient,
       }, nil
   }
   ```

3. Update `cmd/job_run.go` to use `Bootstrap()`:
   ```go
   services, err := app.Bootstrap(ctx, app.BootstrapOptions{
       DaemonMode: false,
       Parallelism: 1,
   })
   if err != nil {
       return err
   }

   return services.Executor.Execute(ctx, jobName, opts)
   ```

4. **Verify**: `go build ./cmd/bosun/` succeeds, `bosun job run` still works.

---

### Day 2: Domain Types & Config Labels (#169)

**Goal**: Add overlap policy and scheduling fields to domain model.

**Files to create/modify**:
- `internal/domain/jobs/overlap.go` (NEW)
- `internal/domain/jobs/status.go` (NEW)
- `internal/domain/jobs/types.go` (EXTEND)
- `internal/config/schema/job_labels.go` (EXTEND)

**Steps**:
1. Create overlap policy type:
   ```go
   // internal/domain/jobs/overlap.go
   type OverlapPolicy string

   const (
       OverlapPolicyQueue OverlapPolicy = "queue"
       OverlapPolicySkip  OverlapPolicy = "skip"
       OverlapPolicyCancelRestart OverlapPolicy = "cancel-and-restart"
   )

   const DefaultOverlapPolicy = OverlapPolicyQueue

   func ValidateOverlapPolicy(p OverlapPolicy) error {
       switch p {
       case OverlapPolicyQueue, OverlapPolicySkip:
           return nil
       case OverlapPolicyCancelRestart:
           return fmt.Errorf("cancel-and-restart not implemented (tracked in #176)")
       default:
           return fmt.Errorf("invalid overlap policy: %s", p)
       }
   }
   ```

2. Extend Job struct:
   ```go
   // internal/domain/jobs/types.go
   type Job struct {
       Name             string
       Schedule         string        // NEW: cron expression
       OverlapPolicy    OverlapPolicy // NEW
       Enabled          bool          // NEW
       // ... existing fields ...
   }
   ```

3. Add labels to schema:
   ```go
   // internal/config/schema/job_labels.go
   type JobLabelsV1 struct {
       // ... existing fields ...
       Schedule      string `bosun:"key=bosun.job.schedule,scope=container,type=string,required=true,doc='Cron schedule'"`
       OverlapPolicy string `bosun:"key=bosun.job.overlap-policy,scope=container,type=string,default=queue,enum=queue|skip,doc='Overlap behavior'"`
       Enabled       *bool  `bosun:"key=bosun.job.enabled,scope=container,type=bool,default=true,doc='Job enabled'"`
   }
   ```

4. Create status types:
   ```go
   // internal/domain/jobs/status.go
   type JobStatus struct {
       JobName       string
       Status        RunStatus
       Schedule      string
       OverlapPolicy OverlapPolicy
       LastRunTime   *time.Time
       LastResult    string
       NextRunTime   time.Time
       CurrentRunID  string
   }

   type RunStatus string
   const (
       RunStatusIdle      RunStatus = "idle"
       RunStatusRunning   RunStatus = "running"
       RunStatusCompleted RunStatus = "completed"
       RunStatusFailed    RunStatus = "failed"
   )
   ```

5. **Verify**: `go test ./internal/domain/jobs/...` passes, labels parse correctly.

---

### Day 3: Concurrency Primitives (#171)

**Goal**: Implement per-stack mutex and global semaphore.

**Files to create**:
- `internal/app/concurrency/stack_lock.go`

**Steps**:
1. Implement StackLockManager:
   ```go
   type StackLockManager struct {
       locks sync.Map // map[string]*sync.Mutex
   }

   func NewStackLockManager() *StackLockManager {
       return &StackLockManager{}
   }

   func (m *StackLockManager) Lock(ctx context.Context, stackName string) error {
       // Get or create mutex for stack
       val, _ := m.locks.LoadOrStore(stackName, &sync.Mutex{})
       mu := val.(*sync.Mutex)

       // Try to acquire with context timeout
       acquired := make(chan struct{})
       go func() {
           mu.Lock()
           close(acquired)
       }()

       select {
       case <-acquired:
           return nil
       case <-ctx.Done():
           return ctx.Err()
       }
   }

   func (m *StackLockManager) Unlock(stackName string) {
       if val, ok := m.locks.Load(stackName); ok {
           mu := val.(*sync.Mutex)
           mu.Unlock()
       }
   }

   func (m *StackLockManager) IsLocked(stackName string) bool {
       // Try non-blocking lock
       val, ok := m.locks.Load(stackName)
       if !ok {
           return false
       }
       mu := val.(*sync.Mutex)
       if mu.TryLock() {
           mu.Unlock()
           return false
       }
       return true
   }

   // LockAll acquires locks for multiple stacks in sorted alphabetical order.
   // Sorted acquisition prevents deadlocks when overlapping stack sets are targeted.
   func (m *StackLockManager) LockAll(ctx context.Context, stacks []string) error {
       sorted := make([]string, len(stacks))
       copy(sorted, stacks)
       sort.Strings(sorted)

       acquired := make([]string, 0, len(sorted))
       for _, stack := range sorted {
           if err := m.Lock(ctx, stack); err != nil {
               // Release already-acquired locks on failure
               m.UnlockAll(acquired)
               return err
           }
           acquired = append(acquired, stack)
       }
       return nil
   }

   // UnlockAll releases locks for multiple stacks.
   func (m *StackLockManager) UnlockAll(stacks []string) {
       for _, stack := range stacks {
           m.Unlock(stack)
       }
   }
   ```

2. Add tests:
   ```go
   func TestStackLock_ConcurrentAccess(t *testing.T) {
       mgr := NewStackLockManager()
       var counter int
       var wg sync.WaitGroup

       // Two goroutines trying to increment counter
       for i := 0; i < 2; i++ {
           wg.Add(1)
           go func() {
               defer wg.Done()
               mgr.Lock(context.Background(), "test-stack")
               defer mgr.Unlock("test-stack")

               // Critical section
               tmp := counter
               time.Sleep(10 * time.Millisecond)
               counter = tmp + 1
           }()
       }

       wg.Wait()
       assert.Equal(t, 2, counter) // Should be 2, not 1 (no race)
   }
   ```

3. **Verify**: `go test -race ./internal/app/concurrency/...` passes.

---

### Day 4: Event System (#170)

**Goal**: Port interface for events + log adapter.

**Files to create**:
- `internal/ports/executor.go` (EXTEND)
- `internal/adapters/events/log_emitter.go`

**Steps**:
1. Define interface:
   ```go
   // internal/ports/executor.go (add to file)
   type EventEmitter interface {
       EmitJobScheduled(ctx context.Context, jobName string, schedule string, nextRun time.Time)
       EmitJobStarted(ctx context.Context, jobName string, runID string)
       EmitJobCompleted(ctx context.Context, jobName string, runID string, duration time.Duration)
       EmitJobFailed(ctx context.Context, jobName string, runID string, err error, duration time.Duration)
       EmitJobSkipped(ctx context.Context, jobName string, reason string)
       EmitJobAdded(ctx context.Context, jobName string)
       EmitJobRemoved(ctx context.Context, jobName string)
       EmitJobChanged(ctx context.Context, jobName string, oldSchedule, newSchedule string)
   }
   ```

2. Implement log adapter:
   ```go
   type LogEmitter struct {
       logger *slog.Logger
   }

   func NewLogEmitter(logger *slog.Logger) *LogEmitter {
       return &LogEmitter{logger: logger}
   }

   func (e *LogEmitter) EmitJobStarted(ctx context.Context, jobName string, runID string) {
       e.logger.InfoContext(ctx, "Job started",
           "event", "job.started",
           "job_name", jobName,
           "run_id", runID,
       )
   }
   // ... implement other methods ...
   ```

3. **Verify**: `go test ./internal/adapters/events/...` passes.

---

### Checkpoint: Phase 1 Complete

**Before proceeding**: Ensure all 4 issues (#168, #169, #170, #171) are closed and tests pass.

```bash
go test ./internal/app/...
go test ./internal/domain/jobs/...
go test ./internal/adapters/events/...
go test -race ./...
```

---

## Phase 2: Integration (Week 2)

### Day 5-6: Wire Concurrency into Executor (#172)

**Goal**: Wrap job execution with stack locks and semaphore.

**Files to modify**:
- `internal/app/executor/executor.go`

**Steps**:
1. Extend Executor struct:
   ```go
   type Executor struct {
       // Existing fields
       planner   ports.JobPlanner
       compose   ports.ComposeController
       worker    ports.WorkerRunner
       docker    *docker.Client

       // NEW
       stackLocks *concurrency.StackLockManager
       globalSem  *semaphore.Weighted
       events     ports.EventEmitter
   }
   ```

2. Update constructor:
   ```go
   func New(planner ports.JobPlanner, compose ports.ComposeController,
            worker ports.WorkerRunner, docker *docker.Client,
            stackLocks *concurrency.StackLockManager, globalSem *semaphore.Weighted,
            events ports.EventEmitter) *Executor {
       return &Executor{
           planner:    planner,
           compose:    compose,
           worker:     worker,
           docker:     docker,
           stackLocks: stackLocks,
           globalSem:  globalSem,
           events:     events,
       }
   }
   ```

3. Wrap Execute() method:
   ```go
   func (e *Executor) Execute(ctx context.Context, jobName string, opts ExecuteOptions) (*jobs.ExecutionResult, error) {
       runID := uuid.New().String()
       startTime := time.Now()

       // Acquire global semaphore
       if err := e.globalSem.Acquire(ctx, 1); err != nil {
           return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
       }
       defer e.globalSem.Release(1)

       // Acquire per-stack lock(s) — sorted order prevents deadlocks
       stacks := /* determine target stacks from job config */
       if err := e.stackLocks.LockAll(ctx, stacks); err != nil {
           return nil, fmt.Errorf("failed to acquire stack locks: %w", err)
       }
       defer e.stackLocks.UnlockAll(stacks)

       // Emit started event
       e.events.EmitJobStarted(ctx, jobName, runID)

       // Execute job (existing logic)
       result, err := e.executeInternal(ctx, jobName, opts)

       duration := time.Since(startTime)
       if err != nil {
           e.events.EmitJobFailed(ctx, jobName, runID, err, duration)
           return nil, err
       }

       e.events.EmitJobCompleted(ctx, jobName, runID, duration)
       return result, nil
   }
   ```

4. **Verify**: `go test ./internal/app/executor/...` passes, race detector clean.

---

### Day 7-10: Scheduler Engine (#173)

**Goal**: Implement cron-based scheduler with config refresh.

**Files to create**:
- `internal/app/scheduler/scheduler.go`
- `internal/app/scheduler/refresh.go`

**Steps** (see [tasks.md](tasks.md) T6 for details):

1. Create Scheduler struct with cron wrapper
2. Implement AddJob with overlap policy wrappers
3. Implement RemoveJob
4. Implement Start/Stop lifecycle
5. Implement config refresh loop
6. Implement ListJobs for status queries
7. Write unit tests with mock executor

**Key code snippets**:

```go
// scheduledJobEntry tracks per-job state within the scheduler
type scheduledJobEntry struct {
    job                 jobs.Job
    entryID             cron.EntryID
    consecutiveFailures int  // reset to 0 on success; auto-disable at 3
}

// AddJob with overlap policy
func (s *Scheduler) AddJob(ctx context.Context, job jobs.Job) error {
    // Validate
    if err := jobs.ValidateOverlapPolicy(job.OverlapPolicy); err != nil {
        return err
    }

    // Create job wrapper
    jobFunc := func() {
        s.executeJob(ctx, job)
    }

    // Apply overlap policy wrapper
    var chain cron.Chain
    switch job.OverlapPolicy {
    case jobs.OverlapPolicyQueue:
        chain = cron.NewChain(cron.DelayIfStillRunning(s.logger))
    case jobs.OverlapPolicySkip:
        chain = cron.NewChain(cron.SkipIfStillRunning(s.logger))
    }

    // Register with cron
    entryID, err := s.cron.AddJob(job.Schedule, chain.Then(jobFunc))
    if err != nil {
        return fmt.Errorf("failed to add job %s: %w", job.Name, err)
    }

    s.entries.Store(job.Name, &scheduledJobEntry{
        job:     job,
        entryID: entryID,
    })
    s.statusMap.Store(job.Name, &jobs.JobStatus{
        JobName:       job.Name,
        Status:        jobs.RunStatusIdle,
        Schedule:      job.Schedule,
        OverlapPolicy: job.OverlapPolicy,
        NextRunTime:   s.cron.Entry(entryID).Next,
    })

    s.events.EmitJobScheduled(ctx, job.Name, job.Schedule, s.cron.Entry(entryID).Next)
    return nil
}

// executeJob wraps execution with concurrency controls and circuit-breaker
func (s *Scheduler) executeJob(ctx context.Context, entry *scheduledJobEntry) {
    // Acquire global semaphore
    if err := s.globalSem.Acquire(ctx, 1); err != nil {
        return
    }
    defer s.globalSem.Release(1)

    // Acquire per-stack locks (sorted order prevents deadlocks)
    if err := s.stackLocks.LockAll(ctx, entry.job.TargetStacks); err != nil {
        return
    }
    defer s.stackLocks.UnlockAll(entry.job.TargetStacks)

    // Execute
    runID := uuid.New().String()
    s.events.EmitJobStarted(ctx, entry.job.Name, runID)
    startTime := time.Now()

    _, err := s.executor.Execute(ctx, entry.job.Name, ports.ExecuteOptions{})
    duration := time.Since(startTime)

    if err != nil {
        entry.consecutiveFailures++
        s.events.EmitJobFailed(ctx, entry.job.Name, runID, err, duration)

        // Circuit-breaker: auto-disable after 3 consecutive failures
        if entry.consecutiveFailures >= 3 {
            s.disableJob(ctx, entry.job.Name)
            s.events.EmitJobCircuitBroken(ctx, entry.job.Name, entry.consecutiveFailures)
        }
    } else {
        entry.consecutiveFailures = 0
        s.events.EmitJobCompleted(ctx, entry.job.Name, runID, duration)
    }

    // Persist state (M4: in-memory no-op; #177: durable write)
    now := time.Now()
    _ = s.stateStore.SaveJobState(ctx, ports.JobState{
        JobName:             entry.job.Name,
        LastRunAt:           &now,
        LastResult:          lastResult(err),
        ConsecutiveFailures: entry.consecutiveFailures,
    })
}
```

**Verify**: `go test ./internal/app/scheduler/...` passes with accelerated cron.

---

### Checkpoint: Phase 2 Complete

**Before proceeding**: Ensure #172 and #173 are closed, scheduler unit tests pass.

```bash
go test ./internal/app/scheduler/...
go test ./internal/app/executor/...
```

---

## Phase 3: User-Facing Features (Week 3)

### Day 11-12: CLI Commands (#175)

**Goal**: Add `bosun daemon` and `bosun job list` commands.

**Files to create**:
- `internal/cmd/daemon.go`
- `internal/cmd/job_list.go`

**Steps**:
1. Implement daemon command:
   ```go
   var daemonCmd = &cobra.Command{
       Use:   "daemon",
       Short: "Run Bosun as a long-lived daemon",
       RunE:  runDaemon,
   }

   func init() {
       daemonCmd.Flags().IntP("parallelism", "p", 1, "Max concurrent jobs")
       daemonCmd.Flags().Duration("refresh-interval", 5*time.Minute, "Config refresh interval")
   }

   func runDaemon(cmd *cobra.Command, args []string) error {
       parallelism, _ := cmd.Flags().GetInt("parallelism")
       refreshInterval, _ := cmd.Flags().GetDuration("refresh-interval")

       ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
       defer stop()

       services, err := app.Bootstrap(ctx, app.BootstrapOptions{
           DaemonMode:      true,
           Parallelism:     parallelism,
           RefreshInterval: refreshInterval,
       })
       if err != nil {
           return fmt.Errorf("failed to bootstrap daemon: %w", err)
       }

       slog.Info("Starting Bosun daemon", "parallelism", parallelism)

       if err := services.Scheduler.Start(ctx); err != nil {
           return fmt.Errorf("failed to start scheduler: %w", err)
       }

       <-ctx.Done()
       slog.Info("Shutdown signal received, stopping gracefully...")

       shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
       defer cancel()

       if err := services.Scheduler.Stop(shutdownCtx); err != nil {
           slog.Error("Graceful shutdown failed", "error", err)
           return err
       }

       slog.Info("Daemon stopped")
       return nil
   }
   ```

2. Implement job list command:
   ```go
   var jobListCmd = &cobra.Command{
       Use:   "list",
       Short: "List scheduled jobs",
       RunE:  runJobList,
   }

   func init() {
       jobListCmd.Flags().String("format", "text", "Output format: text, json, yaml")
   }

   func runJobList(cmd *cobra.Command, args []string) error {
       format, _ := cmd.Flags().GetString("format")

       services, err := app.Bootstrap(context.Background(), app.BootstrapOptions{
           DaemonMode: true,
       })
       if err != nil {
           return err
       }

       jobs := services.Scheduler.ListJobs()

       switch format {
       case "json":
           json.NewEncoder(os.Stdout).Encode(jobs)
       case "yaml":
           yaml.NewEncoder(os.Stdout).Encode(jobs)
       default:
           // Text table output
           for _, job := range jobs {
               fmt.Printf("%s\t%s\t%s\t%s\n", job.JobName, job.Status, job.Schedule, job.NextRunTime)
           }
       }
       return nil
   }
   ```

3. Register commands in `root.go`.

4. **Verify**: Build and test manually.

---

### Day 13-15: Integration Tests (#174)

**Goal**: End-to-end tests with real Docker daemon.

**Files to create**:
- `integration/scheduling_test.go`
- `integration/concurrency_test.go`
- `integration/daemon_lifecycle_test.go`
- `integration/testdata/scheduling-compose.yml`

**Example test**:
```go
//go:build integration

func TestScheduler_RealCron(t *testing.T) {
    ctx := context.Background()

    // Start test compose stack
    compose := testutil.StartComposeFixture(t, "testdata/scheduling-compose.yml")
    defer compose.Down(ctx)

    // Bootstrap daemon
    services, err := app.Bootstrap(ctx, app.BootstrapOptions{
        DaemonMode:      true,
        Parallelism:     1,
        RefreshInterval: 10 * time.Second,
    })
    require.NoError(t, err)

    // Start scheduler
    err = services.Scheduler.Start(ctx)
    require.NoError(t, err)
    defer services.Scheduler.Stop(ctx)

    // Wait for 2 minutes
    time.Sleep(2 * time.Minute)

    // Verify execution counts
    jobs := services.Scheduler.ListJobs()
    assert.Len(t, jobs, 3)

    // Job with */10s schedule should run ~12 times in 2 minutes
    job1 := findJob(jobs, "test-job-1")
    assert.Greater(t, job1.ExecutionCount, 10)
}
```

**Verify**: `go test -race -tags integration ./integration/...` passes.

---

### Checkpoint: Phase 3 Complete

**Before proceeding**: All tests pass, manual testing successful.

```bash
# Unit tests
go test ./...

# Integration tests
go test -race -tags integration ./integration/...

# Manual testing
./bin/bosun daemon
./bin/bosun job list
```

---

## Quick Reference

### Dependencies

```bash
go get github.com/robfig/cron/v3@v3.0.1
go get golang.org/x/sync@v0.19.0
```

### Running Tests

```bash
# Unit tests (fast)
go test ./internal/...

# Integration tests (slow, requires Docker)
go test -tags integration ./integration/...

# Race detector
go test -race ./...

# Coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

```bash
# Build binary
go build -o bin/bosun ./cmd/bosun/

# Run locally
./bin/bosun daemon --parallelism 3
./bin/bosun job list --format json
```

### Debugging

```bash
# Enable debug logs
export BOSUN_LOG_LEVEL=debug
./bin/bosun daemon

# Docker daemon logs
docker logs -f <container-id>

# Check running jobs
docker ps --filter label=bosun.run-id
```

### Common Issues

**Issue**: "Failed to acquire stack lock: context deadline exceeded"
- **Cause**: Job taking too long, context timeout too short
- **Fix**: Increase timeout or check for deadlocks

**Issue**: Race detector failures
- **Cause**: Concurrent access to shared state without proper locking
- **Fix**: Use sync.Mutex, sync.Map, or channels

**Issue**: Jobs not scheduling
- **Cause**: Invalid cron expression or job disabled
- **Fix**: Check logs for parse errors, verify `enabled=true` label

---

## Next Steps

After completing M4:
1. Close milestone #86
2. Update Serena memories (see [plan.md](plan.md) memory update section)
3. Consider #176 (cancel-and-restart) if user demand exists
4. Begin M5 planning (observability & robustness)
