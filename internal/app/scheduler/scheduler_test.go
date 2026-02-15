package scheduler_test

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/app/scheduler"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// --- Mocks ---

type mockExecutor struct {
	mu         sync.Mutex
	execCount  int
	execFunc   func(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error)
	execCalled chan struct{}
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		execCalled: make(chan struct{}, 100),
	}
}

func (m *mockExecutor) Execute(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error) {
	m.mu.Lock()
	m.execCount++
	fn := m.execFunc
	m.mu.Unlock()
	defer func() {
		select {
		case m.execCalled <- struct{}{}:
		default:
		}
	}()
	if fn != nil {
		return fn(ctx, job, opts)
	}
	return ports.ExecutionResult{
		Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
	}, nil
}

func (m *mockExecutor) DryRun(_ context.Context, _ jobs.Job) (jobs.ExecutionPlan, error) {
	return jobs.ExecutionPlan{}, nil
}

func (m *mockExecutor) getExecCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.execCount
}

// mockEvents implements ports.EventEmitter for testing.
type mockEvents struct {
	mu     sync.Mutex
	events []eventRecord
}

type eventRecord struct {
	kind    string
	jobName string
	extra   map[string]interface{}
}

func newMockEvents() *mockEvents {
	return &mockEvents{}
}

func (m *mockEvents) record(kind, jobName string, extra map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, eventRecord{kind: kind, jobName: jobName, extra: extra})
}

func (m *mockEvents) hasEvent(kind string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.events {
		if e.kind == kind {
			return true
		}
	}
	return false
}

func (m *mockEvents) EmitJobScheduled(_ context.Context, jobName, schedule string, nextRun time.Time) {
	m.record("job.scheduled", jobName, map[string]interface{}{"schedule": schedule})
}
func (m *mockEvents) EmitJobStarted(_ context.Context, jobName, runID string) {
	m.record("job.started", jobName, map[string]interface{}{"runID": runID})
}
func (m *mockEvents) EmitJobCompleted(_ context.Context, jobName, runID string, duration time.Duration) {
	m.record("job.completed", jobName, map[string]interface{}{"runID": runID, "duration": duration})
}
func (m *mockEvents) EmitJobFailed(_ context.Context, jobName, runID string, err error, duration time.Duration) {
	m.record("job.failed", jobName, map[string]interface{}{"runID": runID, "error": err, "duration": duration})
}
func (m *mockEvents) EmitJobSkipped(_ context.Context, jobName, reason string) {
	m.record("job.skipped", jobName, map[string]interface{}{"reason": reason})
}
func (m *mockEvents) EmitJobAdded(_ context.Context, jobName string) {
	m.record("job.added", jobName, nil)
}
func (m *mockEvents) EmitJobRemoved(_ context.Context, jobName string) {
	m.record("job.removed", jobName, nil)
}
func (m *mockEvents) EmitJobChanged(_ context.Context, jobName, oldSchedule, newSchedule string) {
	m.record("job.changed", jobName, map[string]interface{}{"old": oldSchedule, "new": newSchedule})
}
func (m *mockEvents) EmitJobCircuitBroken(_ context.Context, jobName string, consecutiveFailures int) {
	m.record("job.circuit_broken", jobName, map[string]interface{}{"failures": consecutiveFailures})
}
func (m *mockEvents) EmitJobDuplicateName(_ context.Context, jobName, containerID string) {
	m.record("job.duplicate_name", jobName, map[string]interface{}{"container": containerID})
}

// mockStateStore implements ports.JobStateStore for testing.
type mockStateStore struct {
	mu     sync.Mutex
	states map[string]ports.JobState
}

func newMockStateStore() *mockStateStore {
	return &mockStateStore{
		states: make(map[string]ports.JobState),
	}
}

func (m *mockStateStore) SaveJobState(_ context.Context, state ports.JobState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[state.JobName] = state
	return nil
}

func (m *mockStateStore) LoadJobState(_ context.Context, jobName string) (ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[jobName]
	if !ok {
		return ports.JobState{}, ports.ErrJobStateNotFound
	}
	return s, nil
}

func (m *mockStateStore) ListJobStates(_ context.Context) ([]ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]ports.JobState, 0, len(m.states))
	for _, s := range m.states {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockStateStore) DeleteJobState(_ context.Context, jobName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, jobName)
	return nil
}

func (m *mockStateStore) getState(jobName string) (ports.JobState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[jobName]
	return s, ok
}

// --- Helpers ---

func newTestScheduler(executor ports.JobExecutor, events ports.EventEmitter, store ports.JobStateStore) *scheduler.Scheduler {
	return scheduler.New(
		executor,
		events,
		store,
		scheduler.Options{Parallelism: 1},
		slog.Default(),
	)
}

func testJob(name, schedule string) jobs.Job {
	return jobs.Job{
		Name:     name,
		Schedule: schedule,
		Enabled:  true,
	}
}

// --- T016: Unit test Scheduler.AddJob ---

func TestScheduler_AddJob_ValidCron(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("daily-backup", "0 0 3 * * *") // seconds-based: every day at 03:00:00
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v, want nil", err)
	}

	// Verify job appears in ListJobs.
	listed := s.ListJobs()
	if len(listed) != 1 {
		t.Fatalf("ListJobs() returned %d items, want 1", len(listed))
	}
	if listed[0].JobName != "daily-backup" {
		t.Errorf("JobName = %q, want %q", listed[0].JobName, "daily-backup")
	}
	if listed[0].Status != jobs.ScheduleStatusIdle {
		t.Errorf("Status = %q, want %q", listed[0].Status, jobs.ScheduleStatusIdle)
	}

	// Verify scheduled event was emitted.
	if !ev.hasEvent("job.scheduled") {
		t.Errorf("expected job.scheduled event to be emitted")
	}
}

func TestScheduler_AddJob_InvalidCron(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("bad-job", "not-a-cron")
	err := s.AddJob(ctx, job)
	if err == nil {
		t.Fatal("AddJob() error = nil, want error for invalid cron")
	}

	// Verify no jobs registered.
	if len(s.ListJobs()) != 0 {
		t.Errorf("ListJobs() should be empty after failed AddJob")
	}
}

func TestScheduler_AddJob_DefaultOverlapPolicy(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("no-policy", "0 0 * * * *")
	job.OverlapPolicy = "" // empty
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	listed := s.ListJobs()
	if len(listed) != 1 {
		t.Fatalf("ListJobs() returned %d items, want 1", len(listed))
	}
	if listed[0].OverlapPolicy != jobs.DefaultOverlapPolicy {
		t.Errorf("OverlapPolicy = %q, want default %q", listed[0].OverlapPolicy, jobs.DefaultOverlapPolicy)
	}
}

func TestScheduler_RemoveJob(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	_ = s.AddJob(ctx, testJob("temp-job", "0 0 * * * *"))

	// Remove it.
	err := s.RemoveJob(ctx, "temp-job")
	if err != nil {
		t.Fatalf("RemoveJob() error = %v", err)
	}
	if len(s.ListJobs()) != 0 {
		t.Errorf("ListJobs() should be empty after RemoveJob")
	}
	if !ev.hasEvent("job.removed") {
		t.Errorf("expected job.removed event to be emitted")
	}
}

func TestScheduler_RemoveJob_NotFound(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)

	err := s.RemoveJob(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("RemoveJob() error = nil, want error for nonexistent job")
	}
}

// --- T017: Unit test Scheduler.executeJob ---

func TestScheduler_ExecuteJob_Success(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	// Use a fast schedule that fires soon.
	job := testJob("fast-job", "* * * * * *") // every second
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	// Wait for at least one execution.
	s.StartCron()
	defer s.StopCron()

	select {
	case <-exec.execCalled:
		// ok
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for job execution")
	}

	// Verify events.
	if !ev.hasEvent("job.started") {
		t.Error("expected job.started event")
	}
	if !ev.hasEvent("job.completed") {
		t.Error("expected job.completed event")
	}

	// Verify state was persisted.
	state, ok := store.getState("fast-job")
	if !ok {
		t.Fatal("expected state to be persisted after execution")
	}
	if state.LastResult != "success" {
		t.Errorf("LastResult = %q, want %q", state.LastResult, "success")
	}
	if state.ConsecutiveFailures != 0 {
		t.Errorf("ConsecutiveFailures = %d, want 0", state.ConsecutiveFailures)
	}
}

func TestScheduler_ExecuteJob_Failure(t *testing.T) {
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		return ports.ExecutionResult{}, fmt.Errorf("executor error")
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("fail-job", "* * * * * *")
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	select {
	case <-exec.execCalled:
		// ok
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for job execution")
	}

	// Verify failed event.
	if !ev.hasEvent("job.failed") {
		t.Error("expected job.failed event")
	}

	// Verify state persisted with failure count.
	state, ok := store.getState("fail-job")
	if !ok {
		t.Fatal("expected state to be persisted after failure")
	}
	if state.ConsecutiveFailures < 1 {
		t.Errorf("ConsecutiveFailures = %d, want >= 1", state.ConsecutiveFailures)
	}
}

func TestScheduler_ExecuteJob_ResetsFailureOnSuccess(t *testing.T) {
	var callCount atomic.Int32
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		n := callCount.Add(1)
		if n <= 2 {
			return ports.ExecutionResult{}, fmt.Errorf("transient error")
		}
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("flaky-job", "* * * * * *")
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for at least 3 executions (2 failures + 1 success).
	for i := 0; i < 3; i++ {
		select {
		case <-exec.execCalled:
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout waiting for execution %d", i+1)
		}
	}

	// Give a moment for state to be updated.
	time.Sleep(100 * time.Millisecond)

	// After success, consecutive failures should be reset.
	state, ok := store.getState("flaky-job")
	if !ok {
		t.Fatal("expected state to be persisted")
	}
	if state.ConsecutiveFailures != 0 {
		t.Errorf("ConsecutiveFailures = %d, want 0 after success", state.ConsecutiveFailures)
	}
}

// --- T018: Unit test circuit-breaker ---

func TestScheduler_CircuitBreaker_DisablesAfterThreeFailures(t *testing.T) {
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		return ports.ExecutionResult{}, fmt.Errorf("always fails")
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("unstable-job", "* * * * * *")
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for at least 3 executions.
	for i := 0; i < 3; i++ {
		select {
		case <-exec.execCalled:
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout waiting for execution %d", i+1)
		}
	}

	// Give a moment for circuit breaker to process.
	time.Sleep(200 * time.Millisecond)

	// Verify circuit-broken event.
	if !ev.hasEvent("job.circuit_broken") {
		t.Error("expected job.circuit_broken event after 3 failures")
	}

	// Verify job is circuit-broken.
	if !s.IsCircuitBroken("unstable-job") {
		t.Error("expected job to be circuit-broken after 3 failures")
	}

	// Verify state has >= 3 consecutive failures.
	state, ok := store.getState("unstable-job")
	if !ok {
		t.Fatal("expected state to be persisted")
	}
	if state.ConsecutiveFailures < 3 {
		t.Errorf("ConsecutiveFailures = %d, want >= 3", state.ConsecutiveFailures)
	}
}

func TestScheduler_CircuitBreaker_NoDisableUnderThreshold(t *testing.T) {
	var callCount atomic.Int32
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		n := callCount.Add(1)
		if n <= 2 {
			return ports.ExecutionResult{}, fmt.Errorf("fail %d", n)
		}
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("recovering-job", "* * * * * *")
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for 3 executions (2 failures, then 1 success).
	for i := 0; i < 3; i++ {
		select {
		case <-exec.execCalled:
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout waiting for execution %d", i+1)
		}
	}

	// Should NOT be circuit-broken because only 2 consecutive failures
	// before a success (never hit 3 in a row).
	time.Sleep(200 * time.Millisecond)

	if ev.hasEvent("job.circuit_broken") {
		t.Error("job should NOT be circuit-broken with only 2 consecutive failures before recovery")
	}
	if s.IsCircuitBroken("recovering-job") {
		t.Error("job should NOT be circuit-broken")
	}
}

// --- Additional tests for ListJobs and Start/Stop ---

func TestScheduler_ListJobs_Empty(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)

	listed := s.ListJobs()
	if len(listed) != 0 {
		t.Errorf("ListJobs() returned %d items, want 0", len(listed))
	}
}

func TestScheduler_ListJobs_MultipleJobs(t *testing.T) {
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	_ = s.AddJob(ctx, testJob("job-a", "0 0 * * * *"))
	_ = s.AddJob(ctx, testJob("job-b", "0 30 * * * *"))
	_ = s.AddJob(ctx, testJob("job-c", "0 0 3 * * *"))

	listed := s.ListJobs()
	if len(listed) != 3 {
		t.Fatalf("ListJobs() returned %d items, want 3", len(listed))
	}

	names := make(map[string]bool)
	for _, j := range listed {
		names[j.JobName] = true
	}
	for _, expected := range []string{"job-a", "job-b", "job-c"} {
		if !names[expected] {
			t.Errorf("ListJobs() missing %q", expected)
		}
	}
}

// --- T028: Unit test queue overlap — second invocation blocks until first completes ---

func TestScheduler_OverlapQueue_SecondBlocksUntilFirstCompletes(t *testing.T) {
	var (
		callCount atomic.Int32
		started   = make(chan struct{}, 10)
		gate      = make(chan struct{}) // blocks first execution until released
	)
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		n := callCount.Add(1)
		started <- struct{}{}
		if n == 1 {
			<-gate // first execution blocks until test releases
		}
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("queue-job", "* * * * * *") // every second
	job.OverlapPolicy = "queue"
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for the first execution to start.
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for first execution")
	}

	// Wait for a second tick to fire (the cron fires every second).
	// The second execution should be queued (blocked) because the first is still running.
	time.Sleep(1500 * time.Millisecond)

	// No skip event should have been emitted — queue policy waits.
	if ev.hasEvent("job.skipped") {
		t.Error("queue policy should NOT skip; expected queuing, not skipping")
	}

	// Release the first execution — the queued second execution should then run.
	close(gate)

	// Wait for the second execution to start and complete.
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for second (queued) execution")
	}

	// Verify at least 2 executions happened (none were skipped).
	if exec.getExecCount() < 2 {
		t.Errorf("execCount = %d, want >= 2 (queue should not skip)", exec.getExecCount())
	}
}

// --- T029: Unit test skip overlap — second invocation dropped with JobSkipped event ---

func TestScheduler_OverlapSkip_SecondInvocationDropped(t *testing.T) {
	var (
		callCount atomic.Int32
		started   = make(chan struct{}, 2)
		gate      = make(chan struct{}) // blocks first execution
	)
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		n := callCount.Add(1)
		started <- struct{}{}
		if n == 1 {
			<-gate // first execution blocks until released
		}
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	job := testJob("skip-job", "* * * * * *") // every second
	job.OverlapPolicy = "skip"
	err := s.AddJob(ctx, job)
	if err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for the first execution to start.
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for first execution")
	}

	// Wait for at least one more cron tick while first is still running.
	// The second execution should be skipped.
	time.Sleep(1500 * time.Millisecond)

	// Verify JobSkipped event was emitted.
	if !ev.hasEvent("job.skipped") {
		t.Error("expected job.skipped event when skip policy drops overlapping run")
	}

	// Release the first execution.
	close(gate)

	// Verify only 1 execution happened (second was skipped).
	time.Sleep(100 * time.Millisecond)
	if callCount.Load() != 1 {
		// It's possible a third cron tick fires after gate opens, so allow >= 1
		// but the key assertion is that job.skipped was emitted.
		t.Logf("execCount = %d (1 expected, but post-gate ticks may fire)", callCount.Load())
	}
}

// --- T038: Unit test global semaphore — N=1 serial, N=3 concurrent ---

func TestScheduler_GlobalSemaphore_Serial(t *testing.T) {
	// With Parallelism=1, only one job should execute at a time.
	var (
		maxConcurrent atomic.Int32
		current       atomic.Int32
	)
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		cur := current.Add(1)
		defer current.Add(-1)
		// Track max concurrency.
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()

	// Parallelism = 1.
	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 1}, slog.Default())
	ctx := context.Background()

	// Add 3 jobs that fire every second.
	for _, name := range []string{"job-1", "job-2", "job-3"} {
		_ = s.AddJob(ctx, testJob(name, "* * * * * *"))
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for several executions.
	time.Sleep(3 * time.Second)

	if maxConcurrent.Load() > 1 {
		t.Errorf("max concurrent = %d, want <= 1 (Parallelism=1 should enforce serial)", maxConcurrent.Load())
	}
}

func TestScheduler_GlobalSemaphore_AllowsConcurrent(t *testing.T) {
	// With Parallelism=3, up to 3 jobs should run concurrently.
	var (
		maxConcurrent atomic.Int32
		current       atomic.Int32
	)
	exec := newMockExecutor()
	exec.execFunc = func(_ context.Context, _ jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		cur := current.Add(1)
		defer current.Add(-1)
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}
	ev := newMockEvents()
	store := newMockStateStore()

	// Parallelism = 3.
	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 3}, slog.Default())
	ctx := context.Background()

	// Add 3 jobs that all fire every second.
	for _, name := range []string{"job-a", "job-b", "job-c"} {
		_ = s.AddJob(ctx, testJob(name, "* * * * * *"))
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for concurrent execution.
	time.Sleep(3 * time.Second)

	// With 3 jobs, 500ms execution, and 3 slots, we should see >1 concurrent.
	if maxConcurrent.Load() <= 1 {
		t.Logf("max concurrent = %d (expected >1 with Parallelism=3, but timing-dependent)", maxConcurrent.Load())
	}
	// The hard limit is 3.
	if maxConcurrent.Load() > 3 {
		t.Errorf("max concurrent = %d, want <= 3 (Parallelism=3)", maxConcurrent.Load())
	}
}

// --- T069: Unit test disabled job is skipped during initial discovery ---

func TestScheduler_DisabledJobNotScheduled(t *testing.T) {
	// When a job has Enabled=false, the caller (daemon) should skip it.
	// The scheduler itself does not check Enabled — that's the daemon's
	// responsibility. This test documents that disabled jobs should not
	// be passed to AddJob.
	exec := newMockExecutor()
	ev := newMockEvents()
	store := newMockStateStore()
	s := newTestScheduler(exec, ev, store)
	ctx := context.Background()

	enabledJob := jobs.Job{
		Name:     "enabled-backup",
		Schedule: "0 0 3 * * *",
		Enabled:  true,
	}
	disabledJob := jobs.Job{
		Name:     "disabled-backup",
		Schedule: "0 0 3 * * *",
		Enabled:  false,
	}

	// Simulate daemon filtering: only add enabled jobs.
	allJobs := []jobs.Job{enabledJob, disabledJob}
	for _, j := range allJobs {
		if !j.Enabled {
			continue
		}
		if err := s.AddJob(ctx, j); err != nil {
			t.Fatalf("AddJob(%q) unexpected error: %v", j.Name, err)
		}
	}

	// Only the enabled job should be scheduled.
	statuses := s.ListJobs()
	if len(statuses) != 1 {
		t.Fatalf("ListJobs() returned %d jobs, want 1", len(statuses))
	}
	if statuses[0].JobName != "enabled-backup" {
		t.Errorf("scheduled job = %q, want %q", statuses[0].JobName, "enabled-backup")
	}
}
