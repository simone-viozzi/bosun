package ports

import (
	"context"
	"errors"
	"time"
)

// ErrJobStateNotFound is returned by LoadJobState when no state exists for the job.
var ErrJobStateNotFound = errors.New("job state not found")

// JobStateStore persists per-job state across daemon restarts.
//
// M4 ships with InMemoryStateStore (no durability — state lost on restart).
// Issue #177 adds a durable adapter (BoltDB or SQLite) as a drop-in replacement,
// enabling catch-up runs and persistent circuit-breaker state.
//
// M4 adapter:  internal/adapters/state/memory.go
// #177 adapter: internal/adapters/state/bolt.go (planned)
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
