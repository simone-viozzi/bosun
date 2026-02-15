// Package contracts defines the port interfaces for M4 Scheduling Engine.
package contracts

import (
	"context"
	"errors"
	"time"
)

// JobStateStore persists per-job state across daemon restarts.
//
// M4 ships with InMemoryStateStore (no durability — state lost on restart).
// Issue #177 adds a durable adapter (BoltDB or SQLite) as a drop-in replacement,
// enabling catch-up runs and persistent circuit-breaker state.
//
// Interface location: internal/ports/state.go
// M4 adapter:         internal/adapters/state/memory.go
// #177 adapter:       internal/adapters/state/bolt.go (planned)
type JobStateStore interface {
	// SaveJobState persists state for a single job.
	// Called after every execution completes (success or failure).
	SaveJobState(ctx context.Context, state JobState) error

	// LoadJobState retrieves persisted state for a single job.
	// Returns ErrJobStateNotFound if no state has been saved for this job.
	LoadJobState(ctx context.Context, jobName string) (JobState, error)

	// ListJobStates returns all persisted job states.
	// Used during startup for the reconcile/catch-up phase (#177).
	ListJobStates(ctx context.Context) ([]JobState, error)

	// DeleteJobState removes persisted state for a job.
	// Called when a job is permanently removed during config refresh.
	DeleteJobState(ctx context.Context, jobName string) error
}

// ErrJobStateNotFound is returned by LoadJobState when no state exists for the job.
var ErrJobStateNotFound = errors.New("job state not found")

// JobState is the subset of per-job runtime state that survives daemon restarts.
// It captures the minimum data needed for:
//   - Catch-up runs (#177): compare LastRunAt against cron schedule to detect missed runs
//   - Circuit-breaker persistence: restore ConsecutiveFailures so auto-disabled jobs
//     stay disabled across restarts
type JobState struct {
	// JobName is the unique identity key (from bosun.job.name label).
	JobName string

	// LastRunAt records when the job last completed execution (success or failure).
	// nil if the job has never run.
	LastRunAt *time.Time

	// LastResult is "success" or "error: <message>".
	LastResult string

	// ConsecutiveFailures tracks the circuit-breaker counter.
	// Reset to 0 on success; job auto-disabled at ≥3.
	ConsecutiveFailures int
}

// Design Notes:
//
// Why a port interface instead of direct sync.Map access?
//
// The Scheduler currently stores per-job state (consecutiveFailures, lastRunTime)
// in private fields of scheduledJobEntry and statusMap. This works for M4's
// in-memory-only model, but when #177 adds persistence, every state mutation
// would need to be refactored to also write to disk.
//
// By defining a port now and wrapping the existing sync.Map behind InMemoryStateStore,
// the Scheduler always goes through the interface. Adding BoltDB persistence becomes
// a Bootstrap() configuration change—not a scheduler refactor.
//
// Reconcile phase (extension point for #177):
//
// The Scheduler.Start() flow gains an explicit reconcile step:
//
//   Start → discover jobs → load persisted states → reconcile (catch-up) → AddJob → cron.Start()
//
// In M4, reconcile is a no-op (InMemoryStateStore returns empty on fresh start).
// In #177, reconcile compares LastRunAt against each job's cron schedule and fires
// catch-up runs for any jobs that missed their window during daemon downtime.

// Example Usage:
//
// In Scheduler.executeJob() — after execution completes:
//
//	now := time.Now()
//	state := ports.JobState{
//	    JobName:             entry.job.Name,
//	    LastRunAt:           &now,
//	    LastResult:          "success",
//	    ConsecutiveFailures: 0,
//	}
//	if err := s.stateStore.SaveJobState(ctx, state); err != nil {
//	    slog.Error("Failed to persist job state", "job", entry.job.Name, "error", err)
//	}
//
// In Scheduler.Start() — for catch-up reconciliation (#177):
//
//	states, _ := s.stateStore.ListJobStates(ctx)
//	for _, state := range states {
//	    job := findJobByName(currentJobs, state.JobName)
//	    if job != nil && state.LastRunAt != nil {
//	        if missedRun(job.Schedule, *state.LastRunAt) {
//	            s.executeJob(ctx, job) // catch-up
//	        }
//	    }
//	}
//
// InMemoryStateStore (M4 default):
//
//	store := state.NewInMemoryStateStore()
//	scheduler := scheduler.New(executor, discoverer, events, store, opts)
//
// BoltStateStore (#177):
//
//	store, _ := state.NewBoltStateStore("/var/lib/bosun/state.db")
//	scheduler := scheduler.New(executor, discoverer, events, store, opts)
