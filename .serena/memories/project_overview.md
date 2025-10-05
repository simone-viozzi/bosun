# Project Overview

Bosun is a Go CLI application designed with hexagonal architecture. The project is currently in its initial development phase, with the main application logic yet to be implemented. It includes adapters for Docker, HTTP, and storage operations, suggesting it may be a tool for managing or orchestrating containerized services or data handling.

## Tech Stack
- **Language**: Go 1.24.6
- **Architecture**: Hexagonal (Ports and Adapters)
- **Dependencies**: Testcontainers for integration testing
- **Build Tool**: Makefile
- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Linting**: golangci-lint
- **Pre-commit**: pre-commit hooks

## Codebase Structure
- `cmd/bosun/main.go`: Application entrypoint
- `internal/app/`: Application core logic
- `internal/adapters/`: External system adapters (docker, http, storage, dockerlabels)
- `internal/config/`: Configuration handling
- `internal/domain/`: Business domain logic (labels domain implemented)
- `internal/ports/`: Interface definitions (labels ports implemented)
- `internal/testutil/`: Integration testing utilities and harness
- `integration/`: Integration tests with Docker Compose stacks
- `api/`: API related code (if any)
- `bin/`: Built binaries (ignored in git)
- `Makefile`: Build and development commands
- `Dockerfile`: Container build configuration
- `.pre-commit-config.yaml`: Pre-commit hooks for code quality
- `.github/workflows/ci.yml`: CI pipeline for testing and linting

## Current Implementation Status
- **Domain & Ports**: Label discovery domain and ports fully implemented
- **Adapters**: Docker labels adapter with container (#23), volume (#24), and network (#24) discovery implemented; utility functions, and Docker client; placeholder directories for docker, http, and storage
- **Application**: Basic wiring in place, TODO for full integration

## Testing Infrastructure
- **Unit Tests**: Standard Go testing with `go test`
- **Integration Tests**: Docker Compose-based tests using testcontainers
- **Test Utilities**: `internal/testutil` package provides:
  - `StartCompose()`: Launches Docker Compose stacks with unique project names
  - `HostPort()`: Gets published ports for services
  - `DumpLogs()`: Saves container logs for debugging
  - Automatic cleanup via `t.Cleanup()`
- **Test Execution**:
  - Unit tests: `make test`
  - Integration tests: `make test-integration` (requires Docker)
  - All tests: `make test && make test-integration`

## Development Guidelines
- Follow hexagonal architecture principles
- Use standard Go formatting and conventions
- Run pre-commit hooks before committing
- Ensure all code passes linting and tests
- Integration tests should be ~10-20 lines using the testutil harness
