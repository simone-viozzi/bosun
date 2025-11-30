package ports

import (
	"context"
	"errors"

	djobs "github.com/simone-viozzi/bosun/internal/domain/jobs"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

// Error sentinels for job discovery and planning.
var (
	// ErrOrphanedDependents is returned when stopping job containers would
	// leave dependent containers running that are not part of the job.
	ErrOrphanedDependents = errors.New("stopping containers would orphan dependents")

	// ErrInvalidSchedule is returned when a cron expression is invalid.
	ErrInvalidSchedule = errors.New("invalid cron schedule")

	// ErrMissingJobName is returned when bosun.job.enabled=true but no name.
	ErrMissingJobName = errors.New("job enabled but name not specified")

	// ErrConflictingJobField is returned when containers set different values.
	ErrConflictingJobField = errors.New("conflicting job field values")
)

// ValidationError represents a validation error for a specific entity.
type ValidationError struct {
	// EntityKind is "container", "volume", or "network".
	EntityKind string `json:"entityKind"`

	// EntityID is the Docker ID of the entity.
	EntityID string `json:"entityId"`

	// EntityName is the human-readable name.
	EntityName string `json:"entityName"`

	// Field is the label key that failed validation.
	Field string `json:"field"`

	// Message describes the validation failure.
	Message string `json:"message"`
}

// JobDiscoverer discovers jobs from a label snapshot.
type JobDiscoverer interface {
	// DiscoverJobs extracts jobs from a snapshot of labeled entities.
	// Returns all valid jobs found, plus any validation errors encountered.
	// A snapshot with no job labels returns an empty slice (not an error).
	DiscoverJobs(ctx context.Context, snapshot dlabels.Snapshot) ([]djobs.Job, []ValidationError, error)
}

// JobPlanner generates execution plans for discovered jobs.
type JobPlanner interface {
	// Plan generates an ExecutionPlan for the given job.
	// The plan is deterministic: same job input produces identical output.
	// Returns an error if the job is invalid or cannot be planned.
	Plan(ctx context.Context, job djobs.Job) (djobs.ExecutionPlan, error)
}
