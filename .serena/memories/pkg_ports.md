# Ports Package

## Scope
Interface contracts in `internal/ports/` defining boundaries between domain and adapters.

## What

### `LabelSource` (`labels.go`)
Discovers labeled Docker entities.
```go
type LabelSource interface {
    Snapshot(ctx context.Context, sel Selector) (labels.Snapshot, error)
}
```

**`Selector`** - Query parameters
- `Prefixes []string` - Label key prefixes to filter (e.g., `["bosun."]`)
- `IncludeStopped bool` - Include stopped containers
- `ProjectFilter []string` - Filter by `com.docker.compose.project`
- `StackFilter []string` - Filter by `bosun.stack`

### `JobDiscoverer` (`planner.go`)
Transforms labels into Jobs.
```go
type JobDiscoverer interface {
    DiscoverJobs(ctx context.Context, snap labels.Snapshot) ([]jobs.Job, []ValidationError, error)
}
```

### `JobPlanner` (`planner.go`)
Generates execution plans.
```go
type JobPlanner interface {
    Plan(ctx context.Context, job jobs.Job) (jobs.ExecutionPlan, error)
}
```

### `ValidationError` (`planner.go`)
Non-fatal discovery error.
- `EntityKind string` - "container", "volume", or "network"
- `EntityID string` - Docker ID
- `EntityName string` - Human-readable name
- `Field string` - Label key that failed validation
- `Message string` - Human-readable description

### `ComposeController` (`compose.go`)
Controls Compose stack lifecycle.
```go
type ComposeController interface {
    StopStack(ctx context.Context, projectName string, opts StopOptions) error
    StartStack(ctx context.Context, projectName string, opts StartOptions) error
    ListStackContainers(ctx context.Context, projectName string) ([]StackContainer, error)
    IsStackRunning(ctx context.Context, projectName string) (bool, error)
}
```

**`StopOptions`** / **`StartOptions`** - Timeout configuration
- `Timeout time.Duration` - Per-container timeout (default: 30s)

**`StackContainer`** - Container in a Compose stack
- `ID`, `Name`, `ServiceName`, `State`, `DependsOn []string`

### `WorkerRunner` (`worker.go`)
Executes backup worker containers.
```go
type WorkerRunner interface {
    Run(ctx context.Context, config WorkerConfig) (WorkerResult, error)
}
```

**`WorkerConfig`** - Worker container configuration
- `Image string` - Container image
- `Env map[string]string` - Environment variables (BOSUN_* + user pass-through)
- `Mounts []VolumeMount` - Volume attachments
- `Timeout time.Duration` - Execution timeout (default: 1h)
- `KeepOnFailure bool` - Preserve container on non-zero exit

**`WorkerResult`** - Execution outcome
- `ExitCode int`, `Logs string`, `ContainerID string`

**Environment Variables** injected by Bosun:
- `BOSUN_JOB_NAME`, `BOSUN_RUN_ID`, `BOSUN_STACK`, `BOSUN_DRY_RUN`

User pass-through via `bosun.job.worker.env.*` labels.

### `JobExecutor` (`executor.go`)
Orchestrates complete job execution using plan-driven interpretation.
```go
type JobExecutor interface {
    Execute(ctx context.Context, job jobs.Job, opts ExecuteOptions) (ExecuteResult, error)
    DryRun(ctx context.Context, job jobs.Job) (DryRunResult, error)
}
```

**Execution Flow** (plan-driven):
1. Generate execution plan from job
2. Pre-validate worker image exists (fail fast)
3. Execute each plan step in order:
   - `StopContainers` - Stop stack (reverse dependency order)
   - `RunWorker` - Run worker container
   - `StartContainers` - Start stack (dependency order) - always attempted, even if worker fails

## Why
Ports enable dependency inversion: domain depends on interfaces, adapters implement them. Easy to mock for testing.

## Related
- `arch_overview` - Hexagonal architecture
- `pkg_adapters_dockerlabels` - Implements LabelSource
- `pkg_adapters_joblabels` - Implements JobDiscoverer
- `pkg_app_planner` - Implements JobPlanner
- `pkg_adapters_docker_compose` - Implements ComposeController
- `pkg_adapters_docker_worker` - Implements WorkerRunner
- `pkg_app_executor` - Implements JobExecutor
