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
Test compose files follow the naming convention `*-compose.yaml` or `validate-*.yaml`.
List them with `ls internal/testutil/compose/` to see the current set.
Each compose file targets a specific integration test scenario (smoke, label discovery, job execution, validation, etc.).

#### Worker Image (`internal/testutil/worker/`)
Contains a `Dockerfile` used to build the test worker image for integration tests.

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

### CI Notes
- CI pre-pulls `alpine:latest` before integration tests (worker image must exist locally; Bosun follows BYOI — Bring Your Own Image).

### Troubleshooting
- Docker not running: Ensure daemon is active
- Port collisions: Use dynamic ports (`"80"` not `"8080:80"`)
- Stale resources: `docker container/network/volume prune -f`
- Missing worker image: `docker pull alpine:latest` (CI does this automatically)

## Why
Integration tests with real Docker provide high confidence; testutil harness keeps tests concise (~10-20 lines).

## Related
- `arch_development_lifecycle` - Test commands
- `pkg_adapters_dockerlabels` - Label discovery tested here
