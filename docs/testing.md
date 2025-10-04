# Testing Guide

This guide explains how to run and write tests for the Bosun project.

## Prerequisites

### Docker

Integration tests require Docker to be installed and running. The test suite uses Docker Compose to spin up test services.

**Check Docker is available:**
```bash
docker --version
docker ps
```

If Docker is not running, start it before running integration tests.

## Running Tests

### Unit Tests

Unit tests validate individual components in isolation without external dependencies.

**Run all unit tests:**
```bash
make test
```

This executes `go test ./...` which runs all standard Go tests across the project.

**Run tests for a specific package:**
```bash
go test ./internal/app
```

**Run with verbose output:**
```bash
go test -v ./internal/app
```

**Run a single test:**
```bash
go test ./internal/app -run TestAppRuns
```

### Integration Tests

Integration tests validate end-to-end behavior using real Docker containers. These tests are marked with the `integration` build tag to keep them separate from fast unit tests.

**Run all integration tests:**
```bash
make it
```

This executes:
```bash
go test -tags=integration -parallel 6 -timeout=20m ./integration/...
```

**Run integration tests with verbose output:**
```bash
make itv
```

This adds the `-v` flag for detailed test output, useful for debugging.

**Run a single integration test:**
```bash
go test -tags=integration -v ./integration -run Test_Integration_Smoke_Placeholder
```

#### Why `-tags=integration`?

The integration tests use the build tag `//go:build integration` to:
- **Separate slow tests from fast tests**: Unit tests complete in milliseconds, while integration tests may take minutes
- **Avoid Docker dependency during development**: Developers can run `make test` without Docker installed
- **Enable targeted CI workflows**: CI can run unit tests quickly on every commit and integration tests less frequently

Without the `-tags=integration` flag, Go will skip all integration test files.

## Test Organization

### Unit Tests
- Located alongside source files (e.g., `internal/app/app_test.go`)
- No build tags required
- Should be fast (< 100ms per test)
- Mock external dependencies

### Integration Tests
- Located in `integration/` package
- Require `//go:build integration` tag
- Use real Docker services via `internal/testutil/` harness
- Test complete workflows end-to-end

## Writing Tests

### Writing a Unit Test

Create a `*_test.go` file in the same package:

```go
package app_test

import (
    "testing"
    "github.com/simone-viozzi/bosun/internal/app"
)

func TestFeature(t *testing.T) {
    // Arrange
    a := app.New()

    // Act
    result := a.DoSomething()

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Writing an Integration Test

Create a test file in `integration/` with the build tag:

```go
//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    "time"
    "github.com/simone-viozzi/bosun/internal/testutil"
)

func Test_Integration_MyFeature(t *testing.T) {
    t.Parallel()

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Start Docker Compose stack
    stack := testutil.StartCompose(t, ctx, "docker-compose.yaml")

    // Get service port
    port := testutil.HostPort(t, ctx, stack.Project, "service-name", 80)

    // Test your feature...
}
```

## Test Utilities

The `internal/testutil/` package provides helpers for integration tests:

- **`StartCompose(t, ctx, files...)`**: Starts a Docker Compose stack with automatic cleanup
- **`HostPort(t, ctx, project, service, containerPort)`**: Gets the published host port for a service
- **`DumpLogs(t, ctx, project, outDir)`**: Dumps container logs to a directory (useful for debugging failures)

Each test gets a unique Docker Compose project name to enable parallel execution.

## Troubleshooting

### Docker Not Running

**Error:** `Cannot connect to the Docker daemon`

**Solution:** Ensure Docker is installed and running. See the [Docker installation guide](https://docs.docker.com/get-docker/) for your platform.

### Port Collisions

**Error:** `Bind for 0.0.0.0:XXXX failed: port is already allocated`

**Solution:** The test harness assigns random host ports to avoid collisions. Ensure the compose file doesn't specify fixed host ports (use `"80"` not `"8080:80"`). See [Docker port binding documentation](https://docs.docker.com/network/#published-ports) for details.

### Integration Test Timeouts

**Error:** `test timed out after 20m`

**Solution:**
1. Check Docker is running and healthy
2. Increase timeout if legitimately needed: `go test -tags=integration -timeout=30m ./integration/...`
3. Check container logs for startup failures:
   ```bash
   docker ps -a
   docker logs <container-id>
   ```

### Permission Denied (Linux)

**Error:** `permission denied while trying to connect to Docker daemon`

**Solution:** Add your user to the docker group. See the [Docker post-installation steps](https://docs.docker.com/engine/install/linux-postinstall/) for instructions.

### Stale Docker Resources

**Error:** Tests fail due to leftover containers/networks from previous runs

**Solution:** Clean up Docker resources:
```bash
# Remove all stopped containers
docker container prune -f

# Remove all unused networks
docker network prune -f

# Remove all unused volumes (careful - this removes all unused volumes)
docker volume prune -f
```

## CI Integration

The CI workflow (if configured) should:
1. Run unit tests on every PR: `make test`
2. Run integration tests before merge: `make it`
3. Ensure Docker is available in the CI environment

## Performance Tips

- Run unit tests frequently during development (they're fast)
- Run integration tests before committing (they're comprehensive)
- Use `t.Parallel()` in integration tests to speed up the suite
- Keep unit tests under 100ms by mocking external dependencies
- Use `make itv` when debugging integration test failures

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Testcontainers Go](https://golang.testcontainers.org/) - Library used by `internal/testutil/`
