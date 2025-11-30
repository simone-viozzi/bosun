# Testing Structure

## Overview
Bosun uses a comprehensive testing strategy with both unit and integration tests. The integration testing infrastructure uses CLI-based Docker Compose via `exec.Command`.

## Test Categories
- **Unit Tests**: Standard Go tests for individual components
- **Integration Tests**: End-to-end tests using Docker Compose stacks

## Integration Testing Infrastructure

### testutil Package (`internal/testutil/`)
Core utilities for integration testing:

#### harness.go
- `ComposeFS`: Embedded filesystem containing compose YAML files from `testutil/compose/*.yaml`
- `Stack` struct: Represents a running compose stack with:
  - `Project`: Unique project name
  - `ComposeDir`: Temp directory with compose files
  - `Files`: List of compose file names
  - `T`: Testing reference
- `StartCompose(t, ctx, files...)`: Main harness function
  - Generates unique project names using slug.Make (format: `bosun-{test_name}-{nanotime}`)
  - Writes embedded compose files to temp directory
  - Runs `docker compose up -d --wait` via `exec.Command`
  - Registers cleanup via `t.Cleanup()` (runs `docker compose down -v`)
  - Returns `*Stack` for further operations
- `Stack.Down(ctx)`: Stops and removes the compose stack
- `Stack.Exec(ctx, service, command...)`: Runs command in service container
- `Stack.Logs(ctx, service)`: Returns service logs
- `Stack.WaitForHealthy(ctx, service, timeout)`: Waits for service health

#### docker.go
- `mustDocker(t)`: Creates Docker client with error handling
- `HostPort(t, ctx, project, service, containerPort)`: Returns published host port
- `DumpLogs(t, ctx, project, outDir)`: Saves container logs to files
- `atoiOrFail(t, s)`: Helper to convert string to int

### Compose Files (`internal/testutil/compose/`)
Embedded Docker Compose configurations:
- `docker-compose.yaml`: Basic nginx service for smoke testing
- `dockerlabels-compose.yaml`: Labeled containers, volumes, networks for label discovery tests
- `joblabels-compose.yaml`: Job labels for plan/job discovery tests
- `joblabels-invalid-compose.yaml`: Invalid job labels for validation error tests
- `validate-valid.yaml`: Valid config labels for validation tests
- `validate-invalid.yaml`: Invalid config labels for validation error tests

### Integration Tests (`integration/`)
- Located in `integration/` package
- Use `//go:build integration` build tag
- Import `internal/testutil` for harness functionality
- Run with `make test-integration`

## Test Execution
```bash
# Unit tests only
make test

# Integration tests only (requires Docker)
make it       # or make itv for verbose

# All tests
make test && make it
```

## Test Patterns
Integration tests follow this pattern:
1. Create context with timeout
2. Call `testutil.StartCompose()` to start stack
3. Validate stack properties (project name, files)
4. Use `testutil.HostPort()` to get service ports
5. Perform HTTP requests or other validations
6. Use `testutil.DumpLogs()` for debugging
7. Automatic cleanup via `t.Cleanup()`

## Dependencies
- `github.com/gosimple/slug` (for unique project names)
- Docker CLI with `docker compose` support
- Docker daemon for integration tests

## Logging
Tests use standard `log` package for visibility into:
- Test start/completion
- Compose stack operations
- Project names and ports

## Troubleshooting

### Common Issues
- **Docker Not Running**: Error `Cannot connect to the Docker daemon` - ensure Docker is running
- **Port Collisions**: Use dynamic ports in compose files (e.g., `"80"` not `"8080:80"`)
- **Test Timeouts**: Increase with `-timeout=30m` flag if needed
- **Permission Denied (Linux)**: Add user to docker group
- **Stale Resources**: Clean up with `docker container/network/volume prune -f`

### Writing Integration Tests
Integration tests follow this pattern:
```go
//go:build integration

func Test_Integration_Feature(t *testing.T) {
    t.Parallel()
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    stack := testutil.StartCompose(t, ctx, "compose-file.yaml")
    // Stack automatically cleaned up via t.Cleanup()

    // Get published port for HTTP requests
    port := testutil.HostPort(t, ctx, stack.Project, "service", 80)

    // Or run bosun CLI and check output
    cmd := exec.CommandContext(ctx, binaryPath)
    output, err := cmd.CombinedOutput()
    // Assert on output...
}
```
