//go:build integration

package integration

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/app/scheduler"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// --- Integration test mocks ---

type integrationMockExecutor struct {
	execCount int32
	execFunc  func(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error)
	called    chan string // sends job name on each execution
}

func newIntegrationMockExecutor() *integrationMockExecutor {
	return &integrationMockExecutor{
		called: make(chan string, 100),
	}
}

func (m *integrationMockExecutor) Execute(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error) {
	atomic.AddInt32(&m.execCount, 1)
	if m.execFunc != nil {
		result, err := m.execFunc(ctx, job, opts)
		select {
		case m.called <- job.Name:
		default:
		}
		return result, err
	}
	select {
	case m.called <- job.Name:
	default:
	}
	return ports.ExecutionResult{
		Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
	}, nil
}

func (m *integrationMockExecutor) DryRun(_ context.Context, _ jobs.Job) (jobs.ExecutionPlan, error) {
	return jobs.ExecutionPlan{}, nil
}

type integrationMockEvents struct {
	mu     sync.Mutex
	events []string
}

func (m *integrationMockEvents) record(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, name)
}

func (m *integrationMockEvents) hasEvent(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.events {
		if e == name {
			return true
		}
	}
	return false
}

func (m *integrationMockEvents) EmitJobScheduled(_ context.Context, _ string, _ string, _ time.Time) {
	m.record("JobScheduled")
}
func (m *integrationMockEvents) EmitJobStarted(_ context.Context, _ string, _ string) {
	m.record("JobStarted")
}
func (m *integrationMockEvents) EmitJobCompleted(_ context.Context, _ string, _ string, _ time.Duration) {
	m.record("JobCompleted")
}
func (m *integrationMockEvents) EmitJobFailed(_ context.Context, _ string, _ string, _ error, _ time.Duration) {
	m.record("JobFailed")
}
func (m *integrationMockEvents) EmitJobSkipped(_ context.Context, _ string, _ string) {
	m.record("JobSkipped")
}
func (m *integrationMockEvents) EmitJobAdded(_ context.Context, _ string) {
	m.record("JobAdded")
}
func (m *integrationMockEvents) EmitJobRemoved(_ context.Context, _ string) {
	m.record("JobRemoved")
}
func (m *integrationMockEvents) EmitJobChanged(_ context.Context, _ string, _ string, _ string) {
	m.record("JobChanged")
}
func (m *integrationMockEvents) EmitJobCircuitBroken(_ context.Context, _ string, _ int) {
	m.record("JobCircuitBroken")
}
func (m *integrationMockEvents) EmitJobDuplicateName(_ context.Context, _ string, _ string) {
	m.record("JobDuplicateName")
}

type integrationMockStateStore struct {
	mu     sync.Mutex
	states map[string]ports.JobState
}

func newIntegrationMockStateStore() *integrationMockStateStore {
	return &integrationMockStateStore{states: make(map[string]ports.JobState)}
}

func (m *integrationMockStateStore) SaveJobState(_ context.Context, state ports.JobState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[state.JobName] = state
	return nil
}

func (m *integrationMockStateStore) LoadJobState(_ context.Context, jobName string) (ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[jobName]
	if !ok {
		return ports.JobState{}, ports.ErrJobStateNotFound
	}
	return s, nil
}

func (m *integrationMockStateStore) ListJobStates(_ context.Context) ([]ports.JobState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]ports.JobState, 0, len(m.states))
	for _, s := range m.states {
		result = append(result, s)
	}
	return result, nil
}

func (m *integrationMockStateStore) DeleteJobState(_ context.Context, jobName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, jobName)
	return nil
}

// --- T077: Scheduler registers and fires jobs with real cron ---

func TestIntegration_Scheduler_RealCronFiring(t *testing.T) {
	t.Parallel()

	exec := newIntegrationMockExecutor()
	ev := &integrationMockEvents{}
	store := newIntegrationMockStateStore()

	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 2}, slog.Default())
	ctx := context.Background()

	// Register a job that fires every minute (5-field cron).
	job := jobs.Job{
		Name:     "every-second-job",
		Schedule: "@every 500ms",
		Enabled:  true,
	}
	if err := s.AddJob(ctx, job); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Start cron (non-blocking).
	s.StartCron()
	defer s.StopCron()

	// Wait for at least 2 executions (should happen within 3 seconds).
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 2 {
		select {
		case <-exec.called:
			received++
		case <-timeout:
			t.Fatalf("timed out waiting for cron executions, got %d", received)
		}
	}

	// Verify events were emitted.
	if !ev.hasEvent("JobScheduled") {
		t.Error("expected JobScheduled event")
	}
	if !ev.hasEvent("JobStarted") {
		t.Error("expected JobStarted event")
	}
	if !ev.hasEvent("JobCompleted") {
		t.Error("expected JobCompleted event")
	}
}

// --- T081: End-to-end scheduling integration (mock-based, no Docker) ---

func TestIntegration_Scheduler_MultipleJobsFiring(t *testing.T) {
	t.Parallel()

	exec := newIntegrationMockExecutor()
	ev := &integrationMockEvents{}
	store := newIntegrationMockStateStore()

	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 3}, slog.Default())
	ctx := context.Background()

	// Register multiple jobs.
	jobNames := []string{"job-alpha", "job-beta", "job-gamma"}
	for _, name := range jobNames {
		job := jobs.Job{
			Name:     name,
			Schedule: "@every 500ms",
			Enabled:  true,
		}
		if err := s.AddJob(ctx, job); err != nil {
			t.Fatalf("AddJob(%q): %v", name, err)
		}
	}

	s.StartCron()
	defer s.StopCron()

	// Collect executions for 3 seconds.
	seen := make(map[string]int)
	timeout := time.After(4 * time.Second)
	for {
		select {
		case name := <-exec.called:
			seen[name]++
			// Once each job has fired at least once, we're good.
			allFired := true
			for _, n := range jobNames {
				if seen[n] < 1 {
					allFired = false
					break
				}
			}
			if allFired {
				goto done
			}
		case <-timeout:
			t.Fatalf("timed out: seen=%v, wanted all jobs to fire", seen)
		}
	}
done:

	// Verify all jobs appeared in ListJobs.
	statuses := s.ListJobs()
	if len(statuses) != 3 {
		t.Errorf("ListJobs returned %d, want 3", len(statuses))
	}
}
