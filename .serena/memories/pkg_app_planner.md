# Planner Service

## Scope
Execution plan generation in `internal/app/planner/`.

## What

Implements `ports.JobPlanner` to generate deterministic execution plans from Jobs.

### `Planner`
- `New() *Planner` - Constructor (stateless)
- `Plan(ctx, job) (ExecutionPlan, error)` - Generate plan

### Generated Steps

1. **StopContainers** (if `TargetContainers` non-empty)
   - Sets `UseComposeStop=true` if all containers belong to one stack
   - Sorted container IDs for determinism

2. **RunWorker**
   - Uses `job.WorkerImage` and `job.AttachVolumes`
   - Sorted volume attachments

3. **StartContainers** (if `TargetContainers` non-empty)
   - Restarts containers stopped in step 1
   - Uses `docker compose start` if `UseComposeStop=true`

### Determinism Guarantees
- Container IDs sorted alphabetically
- Volume attachments sorted by name
- `CreatedAt` set to `time.Now().UTC()` (tests can override)

### Helper Functions
- `extractContainerNames()` - Parse name from ID/name
- `generateStopDescription()` - Human-readable stop step
- `generateRunWorkerDescription()` - Human-readable run step
- `generateStartDescription()` - Human-readable start step

## Why
Pure, stateless planner enables easy testing. Deterministic output ensures reproducible plans.

## Related
- `pkg_ports` - JobPlanner interface
- `pkg_domain_jobs` - Job, ExecutionPlan types
- `pkg_adapters_joblabels` - Provides Job input
