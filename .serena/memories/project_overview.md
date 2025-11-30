# Project Overview

Bosun is a Go CLI application for automated Docker volume backups using declarative labels. Containers and volumes are annotated with `bosun.job.*` labels to define backup jobs. Bosun discovers these labels, validates them, and generates execution plans for backup operations.

## Tech Stack
- **Language**: Go 1.24.9
- **Architecture**: Hexagonal (Ports and Adapters)
- **Integration Testing**: CLI-based `docker compose` via `exec.Command`
- **Build Tool**: Makefile
- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Pre-commit**: pre-commit hooks

## Codebase Structure
- `cmd/bosun/main.go`: Application entrypoint
- `internal/app/`: Application core logic
- `internal/adapters/`: External system adapters (dockerlabels, joblabels, docker, http, storage)
- `internal/app/`: Application services (planner)
- `internal/config/`: Configuration handling (schema, loader, merge)
- `internal/domain/`: Business domain logic (labels, jobs)
- `internal/ports/`: Interface definitions (LabelSource, JobDiscoverer, JobPlanner)
- `internal/testutil/`: Integration testing utilities and harness
- `integration/`: Integration tests with Docker Compose stacks
- `api/`: API related code (if any)
- `bin/`: Built binaries (ignored in git)
- `Makefile`: Build and development commands
- `Dockerfile`: Container build configuration
- `.pre-commit-config.yaml`: Pre-commit hooks for code quality
- `.github/workflows/ci.yml`: CI pipeline for testing and linting

## Implementation Status
- **Domain**: Labels and Jobs domain types fully implemented
- **Ports**: LabelSource, JobDiscoverer, JobPlanner interfaces defined
- **Config System**: Code-first schema, label loader, source merger, job label validation
- **CLI Commands**:
  - `bosun config validate` - validates config and job labels
  - `bosun labels snapshot` - shows Docker entities with bosun.* labels
  - `bosun plan list` - lists discovered backup jobs
  - `bosun plan show <job>` - shows execution plan for a job
- **Documentation**: Auto-generated `docs/config.md` and `docs/config.schema.json`
- **Adapters**: dockerlabels (label discovery), joblabels (job discovery from labels)
- **Application**: Planner service with topological dependency resolution

## Testing Infrastructure
- **Unit Tests**: Standard Go testing with `go test`
- **Integration Tests**: Docker Compose-based tests using CLI-based harness
- **Test Utilities**: `internal/testutil` package provides:
  - `StartCompose()`: Launches Docker Compose stacks with unique project names via CLI
  - `Stack.Down()`: Tears down compose stack
  - `Stack.Exec()`: Runs commands in service containers
  - `Stack.Logs()`: Gets service logs
  - Automatic cleanup via `t.Cleanup()`
- **Test Compose Files**: Embedded in `internal/testutil/compose/*.yaml`
- **Test Execution**:
  - Unit tests: `make test`
  - Integration tests: `make it` (requires Docker), or `make itv` for verbose
  - All tests: `make test && make it`

## Development Guidelines
- Follow hexagonal architecture principles
- Use standard Go formatting and conventions
- Run pre-commit hooks before committing
- Ensure all code passes linting and tests
- Integration tests should be ~10-20 lines using the testutil harness
