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

## Why
Ports enable dependency inversion: domain depends on interfaces, adapters implement them. Easy to mock for testing.

## Related
- `arch_overview` - Hexagonal architecture
- `pkg_adapters_dockerlabels` - Implements LabelSource
- `pkg_adapters_joblabels` - Implements JobDiscoverer
- `pkg_app_planner` - Implements JobPlanner
