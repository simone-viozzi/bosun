# Planner Service

**Location**: `internal/app/planner/`

## Overview

Application service that generates execution plans from backup jobs. Implements the `ports.JobPlanner` interface. Produces deterministic, ordered steps for job execution.

## Key Types

### `Planner`
Stateless struct implementing `ports.JobPlanner`:
- `New() *Planner` - Constructor
- `Plan(ctx, job) (ExecutionPlan, error)` - Generate execution plan

## Execution Plan Generation

The `Plan` method transforms a `jobs.Job` into an `ExecutionPlan`:

### Steps Generated

1. **StopContainers** (if `TargetContainers` is non-empty)
   - Stops containers before backup
   - Sets `UseComposeStop=true` if all containers belong to one stack
   - Includes container IDs and human-readable names

2. **RunWorker**
   - Runs the worker container with attached volumes
   - Uses `job.WorkerImage` and `job.AttachVolumes`

3. **StartContainers** (future milestone - not yet implemented)

### Determinism

- Container IDs are sorted alphabetically
- Volume attachments are sorted by name
- `CreatedAt` is set to `time.Now().UTC()` (caller can override for tests)

## Helper Functions

- `extractContainerNames(containerIDs)` - Extracts name portion from container ID/name
- `generateStopDescription(names, useCompose, project)` - Human-readable stop step description
- `generateRunWorkerDescription(image, volumes)` - Human-readable run step description

## Context Handling

Respects context cancellation - returns `ctx.Err()` if context is already cancelled.

## Usage

```go
planner := planner.New()
job := jobs.Job{Name: "daily-backup", ...}
plan, err := planner.Plan(ctx, job)
// plan.Steps contains ordered execution steps
```

## Testing

Unit tests verify:
- Empty job produces plan with only run_worker step
- Job with containers produces stop + run_worker steps
- Single stack sets `UseComposeStop=true`
- Multiple stacks leave `UseComposeStop=false`
- Output is deterministic regardless of input order
