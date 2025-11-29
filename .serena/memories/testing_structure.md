# Testing Structure

## Overview
Bosun uses a comprehensive testing strategy with both unit and integration tests. The integration testing infrastructure is built around Docker Compose stacks managed by testcontainers-go.

## Test Categories
- **Unit Tests**: Standard Go tests for individual components
- **Integration Tests**: End-to-end tests using Docker Compose stacks

## Integration Testing Infrastructure

### testutil Package (`internal/testutil/`)
Core utilities for integration testing:

#### harness.go
- `ComposeFS`: Embedded filesystem containing compose YAML files from `testutil/compose/*.yaml`
- `Stack` struct: Represents a running compose stack with project name, files, compose instance, and test reference
- `StartCompose(t, ctx, files...)`: Main harness function
  - Generates unique project names using slug.Make (format: `bosun-{test_name}-{nanotime}`)
  - Embeds compose YAML files and starts them without temp files
  - Starts Docker Compose stacks with automatic cleanup via t.Cleanup()
  - Returns `*Stack` with project name, file paths, and compose instance

#### docker.go
- `mustDocker(t)`: Creates Docker client with error handling
- `HostPort(t, ctx, project, service, containerPort)`: Returns published host port for a service in a compose project
- `DumpLogs(t, ctx, project, outDir)`: Saves container logs to files in outDir
- `atoiOrFail(t, s)`: Helper to convert string to int with test failure

### Compose Files (`internal/testutil/compose/`)
Embedded Docker Compose configurations:
- `docker-compose.yaml`: Basic nginx service on port 80 for smoke testing
- `dockerlabels-compose.yaml`: Labeled containers, volumes, and networks for integration testing

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
- `github.com/testcontainers/testcontainers-go`
- `github.com/testcontainers/testcontainers-go/modules/compose`
- `github.com/gosimple/slug`
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
    port := testutil.HostPort(t, ctx, stack.Project, "service", 80)
    // Test logic...
}
```</content>
<parameter name="memory_name">testing_structure
