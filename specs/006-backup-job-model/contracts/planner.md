# Contract: Job Planner Interface

**Feature Branch**: `006-backup-job-model`
**Date**: 2025-11-30
**Package**: `internal/ports/planner.go`

## Overview

The Planner interface defines the contract for transforming discovered jobs into executable plans. Following hexagonal architecture, this interface lives in the ports layer and is implemented by adapters/application logic.

---

## Interface Definition

```go
// Package: internal/ports

import (
    "context"

    dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
    djobs "github.com/simone-viozzi/bosun/internal/domain/jobs"
)

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

// ValidationError represents a validation error for a specific entity.
type ValidationError struct {
    // EntityKind is "container", "volume", or "network".
    EntityKind string

    // EntityID is the Docker ID of the entity.
    EntityID string

    // EntityName is the human-readable name.
    EntityName string

    // Field is the label key that failed validation.
    Field string

    // Message describes the validation failure.
    Message string
}
```

---

## Method Contracts

### JobDiscoverer.DiscoverJobs

**Purpose**: Extract job definitions from a snapshot of Docker entities.

**Preconditions**:
- `snapshot` is a valid `Snapshot` (may be empty)
- Context is not cancelled

**Postconditions**:
- Returns all jobs that can be assembled from the snapshot
- Jobs are merged by `bosun.job.name` (multiple containers → one job)
- Validation errors are collected, not short-circuited
- System errors (non-validation) are returned separately

**Behavior**:

| Input | Output |
|-------|--------|
| Empty snapshot | `[]Job{}`, `[]ValidationError{}`, `nil` |
| No job labels | `[]Job{}`, `[]ValidationError{}`, `nil` |
| Valid job labels | `[]Job{...}`, `[]ValidationError{}`, `nil` |
| Invalid labels | `[]Job{...}`, `[]ValidationError{...}`, `nil` |
| Cancelled context | `nil`, `nil`, `context.Canceled` |

**Validation Errors Collected**:
- `bosun.job.enabled` not a valid boolean
- `bosun.job.schedule` not a valid cron expression
- `bosun.job.name` missing when `enabled=true`
- Conflicting values for same field across containers (same job name)
- `bosun.job.attach` referencing non-existent job (warning, not error)

---

### JobPlanner.Plan

**Purpose**: Generate an execution plan for a single job.

**Preconditions**:
- `job` is a valid `Job` (has been validated by discoverer)
- Context is not cancelled

**Postconditions**:
- Returns a deterministic `ExecutionPlan`
- Same `job` input always produces identical plan
- Steps are ordered: stop → run-worker (→ start, future milestone)

**Behavior**:

| Input | Output |
|-------|--------|
| Job with containers | Plan with stop + run-worker steps |
| Job with no containers | Plan with only run-worker step |
| Job with volumes | run-worker step includes volume mounts |
| Job with no volumes | run-worker step has empty mounts |
| Dependency validation failure | `nil`, `ErrOrphanedDependents` |

**Error Conditions**:
- `ErrOrphanedDependents`: Stopping target containers would leave dependent containers running that are not in the job
- Context cancelled

---

## Error Types

```go
// Package: internal/ports

import "errors"

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
```

---

## Usage Examples

### Discovering Jobs

```go
discoverer := joblabels.NewDiscoverer()
labelSource := dockerlabels.NewFromEnv()

snapshot, err := labelSource.Snapshot(ctx, ports.Selector{
    Prefixes: []string{labels.DefaultLabelPrefix},
})
if err != nil {
    return fmt.Errorf("snapshot failed: %w", err)
}

jobs, validationErrors, err := discoverer.DiscoverJobs(ctx, snapshot)
if err != nil {
    return fmt.Errorf("discovery failed: %w", err)
}

if len(validationErrors) > 0 {
    for _, ve := range validationErrors {
        fmt.Fprintf(os.Stderr, "%s %q: %s\n", ve.EntityKind, ve.EntityName, ve.Message)
    }
}
```

### Generating a Plan

```go
planner := appplanner.New()

for _, job := range jobs {
    plan, err := planner.Plan(ctx, job)
    if errors.Is(err, ports.ErrOrphanedDependents) {
        fmt.Fprintf(os.Stderr, "Job %q: cannot stop containers - dependents would be orphaned\n", job.Name)
        continue
    }
    if err != nil {
        return fmt.Errorf("planning %q failed: %w", job.Name, err)
    }

    // Use plan...
}
```

---

## Implementation Notes

1. **Separation of Concerns**: `JobDiscoverer` handles label parsing/validation; `JobPlanner` handles plan generation
2. **Validation vs Planning**: Validation errors are non-fatal during discovery; planning errors are fatal
3. **Determinism**: Planner must sort inputs before processing to ensure stable output
4. **No Docker Calls**: Neither interface makes Docker API calls; they work on in-memory data only
