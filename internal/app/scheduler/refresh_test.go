package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// --- Internal test mocks (package scheduler, for white-box access) ---

type refreshMockExecutor struct {
	execFunc func(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error)
}

func (m *refreshMockExecutor) Execute(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, job, opts)
	}
	return ports.ExecutionResult{
		Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
	}, nil
}

func (m *refreshMockExecutor) DryRun(_ context.Context, _ jobs.Job) (jobs.ExecutionPlan, error) {
	return jobs.ExecutionPlan{}, nil
}

type refreshMockEvents struct {
	mu     sync.Mutex
	events []string
}

func newRefreshMockEvents() *refreshMockEvents { return &refreshMockEvents{} }

func (m *refreshMockEvents) record(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, name)
}

func (m *refreshMockEvents) hasEvent(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.events {
		if e == name {
			return true
		}
	}
	return false
}

func (m *refreshMockEvents) EmitJobScheduled(_ context.Context, _ string, _ string, _ time.Time) {
	m.record("JobScheduled")
}
func (m *refreshMockEvents) EmitJobStarted(_ context.Context, _ string, _ string) {
	m.record("JobStarted")
}
func (m *refreshMockEvents) EmitJobCompleted(_ context.Context, _ string, _ string, _ time.Duration) {
	m.record("JobCompleted")
}
func (m *refreshMockEvents) EmitJobFailed(_ context.Context, _ string, _ string, _ error, _ time.Duration) {
	m.record("JobFailed")
}
func (m *refreshMockEvents) EmitJobSkipped(_ context.Context, _ string, _ string) {
	m.record("JobSkipped")
}
func (m *refreshMockEvents) EmitJobAdded(_ context.Context, _ string) {
	m.record("JobAdded")
}
func (m *refreshMockEvents) EmitJobRemoved(_ context.Context, _ string) {
	m.record("JobRemoved")
}
func (m *refreshMockEvents) EmitJobChanged(_ context.Context, _, _, _ string) {
	m.record("JobChanged")
}
func (m *refreshMockEvents) EmitJobCircuitBroken(_ context.Context, _ string, _ int) {
	m.record("JobCircuitBroken")
}
func (m *refreshMockEvents) EmitJobDuplicateName(_ context.Context, _ string, _ string) {
	m.record("JobDuplicateName")
}

type refreshMockStateStore struct {
	mu    sync.Mutex
	store map[string]ports.JobState
}

func newRefreshMockStateStore() *refreshMockStateStore {
	return &refreshMockStateStore{store: make(map[string]ports.JobState)}
}

func (m *refreshMockStateStore) LoadJobState(_ context.Context, jobName string) (ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.store[jobName]
	if !ok {
		return ports.JobState{}, ports.ErrJobStateNotFound
	}
	return s, nil
}

func (m *refreshMockStateStore) SaveJobState(_ context.Context, state ports.JobState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[state.JobName] = state
	return nil
}

func (m *refreshMockStateStore) DeleteJobState(_ context.Context, jobName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, jobName)
	return nil
}

func (m *refreshMockStateStore) ListJobStates(_ context.Context) ([]ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]ports.JobState, 0, len(m.store))
	for _, s := range m.store {
		result = append(result, s)
	}
	return result, nil
}

// --- T053: Detect new jobs added since last refresh ---

func TestDiffJobs_DetectsNewJobs(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register job A.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Discover jobs A and B (B is new).
	jobB := makeTestJob("job-b", "0 30 * * * *")
	diff := sched.diffJobs([]jobs.Job{jobA, jobB})

	if len(diff.Added) != 1 {
		t.Fatalf("expected 1 added job, got %d", len(diff.Added))
	}
	if diff.Added[0].Name != "job-b" {
		t.Errorf("added job = %q, want %q", diff.Added[0].Name, "job-b")
	}
	if len(diff.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(diff.Removed))
	}
	if len(diff.Changed) != 0 {
		t.Errorf("expected 0 changed, got %d", len(diff.Changed))
	}
}

// --- T054: Detect removed jobs and changed schedules/policies ---

func TestDiffJobs_DetectsRemovedJobs(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register jobs A and B.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	jobB := makeTestJob("job-b", "0 30 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob A: %v", err)
	}
	if err := sched.AddJob(ctx, jobB); err != nil {
		t.Fatalf("AddJob B: %v", err)
	}

	// Discovery only returns A (B is removed).
	diff := sched.diffJobs([]jobs.Job{jobA})

	if len(diff.Removed) != 1 {
		t.Fatalf("expected 1 removed job, got %d", len(diff.Removed))
	}
	if diff.Removed[0] != "job-b" {
		t.Errorf("removed job = %q, want %q", diff.Removed[0], "job-b")
	}
}

func TestDiffJobs_DetectsChangedSchedule(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register job A with original schedule.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Discover job A with changed schedule.
	changedA := makeTestJob("job-a", "0 15 * * * *")
	diff := sched.diffJobs([]jobs.Job{changedA})

	if len(diff.Changed) != 1 {
		t.Fatalf("expected 1 changed job, got %d", len(diff.Changed))
	}
	if diff.Changed[0].NewJob.Name != "job-a" {
		t.Errorf("changed job = %q, want %q", diff.Changed[0].NewJob.Name, "job-a")
	}
	if diff.Changed[0].OldSchedule != "0 0 * * * *" {
		t.Errorf("old schedule = %q, want %q", diff.Changed[0].OldSchedule, "0 0 * * * *")
	}
	if diff.Changed[0].NewJob.Schedule != "0 15 * * * *" {
		t.Errorf("new schedule = %q, want %q", diff.Changed[0].NewJob.Schedule, "0 15 * * * *")
	}
}

func TestDiffJobs_DetectsChangedOverlapPolicy(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register job A with queue policy.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	jobA.OverlapPolicy = jobs.OverlapPolicyQueue
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Discover job A with skip policy.
	changedA := makeTestJob("job-a", "0 0 * * * *")
	changedA.OverlapPolicy = jobs.OverlapPolicySkip
	diff := sched.diffJobs([]jobs.Job{changedA})

	if len(diff.Changed) != 1 {
		t.Fatalf("expected 1 changed job, got %d", len(diff.Changed))
	}
	if diff.Changed[0].OldOverlapPolicy != jobs.OverlapPolicyQueue {
		t.Errorf("old policy = %v, want %v", diff.Changed[0].OldOverlapPolicy, jobs.OverlapPolicyQueue)
	}
	if diff.Changed[0].NewJob.OverlapPolicy != jobs.OverlapPolicySkip {
		t.Errorf("new policy = %v, want %v", diff.Changed[0].NewJob.OverlapPolicy, jobs.OverlapPolicySkip)
	}
}

func TestDiffJobs_NoChanges(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	diff := sched.diffJobs([]jobs.Job{jobA})
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.Changed) != 0 {
		t.Errorf("expected no diff, got added=%d removed=%d changed=%d",
			len(diff.Added), len(diff.Removed), len(diff.Changed))
	}
}

// --- T055: Circuit-broken jobs are NOT re-enabled by config refresh (FR-041) ---

func TestDiffJobs_CircuitBrokenNotReEnabled(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register a job, then mark it circuit-broken.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	sched.mu.Lock()
	sched.entries["job-a"].circuitBroken = true
	sched.mu.Unlock()

	// Discovery returns job-a with a changed schedule.
	changedA := makeTestJob("job-a", "0 30 * * * *")
	diff := sched.diffJobs([]jobs.Job{changedA})

	// Circuit-broken job should NOT appear in changed list.
	if len(diff.Changed) != 0 {
		t.Errorf("expected 0 changed (circuit-broken), got %d", len(diff.Changed))
	}
	// And should NOT appear in added list either.
	if len(diff.Added) != 0 {
		t.Errorf("expected 0 added (circuit-broken), got %d", len(diff.Added))
	}
}

func TestDiffJobs_CircuitBrokenCanBeRemoved(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()

	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	sched.mu.Lock()
	sched.entries["job-a"].circuitBroken = true
	sched.mu.Unlock()

	// Discovery returns empty: circuit-broken job is removed.
	diff := sched.diffJobs([]jobs.Job{})

	if len(diff.Removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(diff.Removed))
	}
	if diff.Removed[0] != "job-a" {
		t.Errorf("removed = %q, want %q", diff.Removed[0], "job-a")
	}
}

// --- ApplyRefresh integration tests ---

func TestApplyRefresh_AddsAndRemovesJobs(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()
	events := sched.events.(*refreshMockEvents)

	// Register jobs A and B.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	jobB := makeTestJob("job-b", "0 30 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob A: %v", err)
	}
	if err := sched.AddJob(ctx, jobB); err != nil {
		t.Fatalf("AddJob B: %v", err)
	}

	// Refresh discovers A (unchanged), C (new). B is removed.
	jobC := makeTestJob("job-c", "0 45 * * * *")
	diff := sched.ApplyRefresh(ctx, []jobs.Job{jobA, jobC})

	if len(diff.Added) != 1 || diff.Added[0].Name != "job-c" {
		t.Errorf("expected job-c added, got %v", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0] != "job-b" {
		t.Errorf("expected job-b removed, got %v", diff.Removed)
	}

	// Verify events were emitted.
	if !events.hasEvent("JobAdded") {
		t.Error("expected JobAdded event")
	}
	if !events.hasEvent("JobRemoved") {
		t.Error("expected JobRemoved event")
	}
}

func TestApplyRefresh_ChangedJobEmitsEvent(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx := context.Background()
	events := sched.events.(*refreshMockEvents)

	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Refresh with changed schedule.
	changedA := makeTestJob("job-a", "0 30 * * * *")
	diff := sched.ApplyRefresh(ctx, []jobs.Job{changedA})

	if len(diff.Changed) != 1 {
		t.Fatalf("expected 1 changed, got %d", len(diff.Changed))
	}
	if !events.hasEvent("JobChanged") {
		t.Error("expected JobChanged event")
	}
}

// --- Refresh loop test ---

func TestStartRefreshLoop_PeriodicDiscovery(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var callCount int32
	discoverFn := func(_ context.Context) ([]jobs.Job, error) {
		c := atomic.AddInt32(&callCount, 1)
		// Return a new job on the 2nd call.
		if c >= 2 {
			return []jobs.Job{makeTestJob("dynamic-job", "0 0 * * * *")}, nil
		}
		return []jobs.Job{}, nil
	}

	go sched.startRefreshLoop(ctx, 50*time.Millisecond, discoverFn)

	// Wait for at least 2 refresh cycles.
	time.Sleep(200 * time.Millisecond)
	cancel()

	if atomic.LoadInt32(&callCount) < 2 {
		t.Fatalf("expected at least 2 discover calls, got %d", atomic.LoadInt32(&callCount))
	}

	// Verify the dynamic job was added.
	jobList := sched.ListJobs()
	found := false
	for _, js := range jobList {
		if js.JobName == "dynamic-job" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'dynamic-job' to be registered after refresh")
	}
}

func TestStartRefreshLoop_NilFuncDisablesRefresh(t *testing.T) {
	t.Parallel()

	sched := newTestScheduler(t)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.startRefreshLoop(ctx, 10*time.Millisecond, nil)
		close(done)
	}()

	// Should return immediately since discoverFn is nil.
	select {
	case <-done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("startRefreshLoop did not return for nil discoverFn")
	}
	cancel()
}

// --- Test helpers ---

func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()
	return New(
		&refreshMockExecutor{},
		newRefreshMockEvents(),
		newRefreshMockStateStore(),
		Options{Parallelism: 1},
		slog.Default(),
	)
}

func makeTestJob(name, schedule string) jobs.Job {
	return jobs.Job{
		Name:          name,
		Schedule:      schedule,
		OverlapPolicy: jobs.OverlapPolicyQueue,
		Enabled:       true,
	}
}

// --- T070: enabled→disabled transition removes job on refresh ---

func TestDiffJobs_EnabledToDisabled_RemovesJob(t *testing.T) {
	sched := newTestScheduler(t)
	ctx := context.Background()

	// Register an enabled job.
	job := makeTestJob("backup-db", "0 0 3 * * *")
	if err := sched.AddJob(ctx, job); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Simulate: on next discovery, the job is no longer returned
	// (because isJobEnabled now returns false for it).
	discovered := []jobs.Job{} // empty = job was disabled

	diff := sched.diffJobs(discovered)

	if len(diff.Removed) != 1 || diff.Removed[0] != "backup-db" {
		t.Errorf("Removed = %v, want [backup-db]", diff.Removed)
	}
	if len(diff.Added) != 0 {
		t.Errorf("Added = %v, want empty", diff.Added)
	}
	if len(diff.Changed) != 0 {
		t.Errorf("Changed = %v, want empty", diff.Changed)
	}
}

// --- T071: disabled→enabled transition adds job on refresh ---

func TestDiffJobs_DisabledToEnabled_AddsJob(t *testing.T) {
	sched := newTestScheduler(t)

	// Initially no jobs registered (the job was disabled).
	// After enabling via label, the discoverer now returns it.
	discovered := []jobs.Job{
		makeTestJob("backup-db", "0 0 3 * * *"),
	}

	diff := sched.diffJobs(discovered)

	if len(diff.Added) != 1 || diff.Added[0].Name != "backup-db" {
		t.Errorf("Added = %v, want [{Name: backup-db}]", diff.Added)
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", diff.Removed)
	}
	if len(diff.Changed) != 0 {
		t.Errorf("Changed = %v, want empty", diff.Changed)
	}
}

// --- T070/T071 integration: ApplyRefresh handles enable/disable transitions ---

func TestApplyRefresh_EnableDisableTransitions(t *testing.T) {
	sched := newTestScheduler(t)
	ctx := context.Background()

	// Start with one enabled job.
	jobA := makeTestJob("job-a", "0 0 * * * *")
	if err := sched.AddJob(ctx, jobA); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Simulate: job-a gets disabled (disappears), job-b gets enabled (appears).
	jobB := makeTestJob("job-b", "0 30 * * * *")
	discovered := []jobs.Job{jobB}

	sched.ApplyRefresh(ctx, discovered)

	statuses := sched.ListJobs()
	if len(statuses) != 1 {
		t.Fatalf("ListJobs() = %d jobs, want 1", len(statuses))
	}
	if statuses[0].JobName != "job-b" {
		t.Errorf("remaining job = %q, want %q", statuses[0].JobName, "job-b")
	}

	// Verify events were emitted.
	ev := sched.events.(*refreshMockEvents) //nolint:errcheck // test mock
	if !ev.hasEvent("JobRemoved") {
		t.Error("expected JobRemoved event for disabled job-a")
	}
	if !ev.hasEvent("JobAdded") {
		t.Error("expected JobAdded event for enabled job-b")
	}
}
