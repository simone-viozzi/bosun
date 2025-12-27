# Testing Infrastructure

## Scope
Unit and integration testing patterns, testutil harness, and test organization.

## What

### Test Categories
- **Unit tests**: Standard Go tests alongside source files (`*_test.go`)
- **Integration tests**: Docker Compose-based, in `integration/` package

### Integration Test Harness (`internal/testutil/`)

#### Core Functions
- `StartCompose(t, ctx, files...)` - Start compose stack with unique project name
- `Stack.Down(ctx)` - Stop and remove stack
- `Stack.Exec(ctx, service, cmd...)` - Run command in container
- `Stack.Logs(ctx, service)` - Get service logs
- `Stack.WaitForHealthy(ctx, service, timeout)` - Wait for health

#### Helper Functions
- `HostPort(t, ctx, project, service, port)` - Get published port
- `DumpLogs(t, ctx, project, outDir)` - Save logs for debugging

#### Compose Files (`internal/testutil/compose/`)
- `docker-compose.yaml` - Basic smoke test
- `dockerlabels-compose.yaml` - Label discovery tests
- `joblabels-compose.yaml` - Job discovery tests
- `validate-*.yaml` - Config validation tests

### Integration Test Pattern
```go
//go:build integration

func TestFeature(t *testing.T) {
    t.Parallel()
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    stack := testutil.StartCompose(t, ctx, "compose-file.yaml")
    // Automatic cleanup via t.Cleanup()

    port := testutil.HostPort(t, ctx, stack.Project, "service", 80)
    // Perform assertions...
}
```

### Troubleshooting
- Docker not running: Ensure daemon is active
- Port collisions: Use dynamic ports (`"80"` not `"8080:80"`)
- Stale resources: `docker container/network/volume prune -f`

## Why
Integration tests with real Docker provide high confidence; testutil harness keeps tests concise (~10-20 lines).

## Related
- `arch_development_lifecycle` - Test commands
- `pkg_adapters_dockerlabels` - Label discovery tested here
