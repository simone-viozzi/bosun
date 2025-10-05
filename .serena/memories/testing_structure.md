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
make test-integration

# All tests
make test && make test-integration
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
- Project names and ports</content>
<parameter name="memory_name">testing_structure
