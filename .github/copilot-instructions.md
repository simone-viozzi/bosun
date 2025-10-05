# Bosun - AI Coding Guidelines

## Architecture Overview
Bosun follows **hexagonal/clean architecture** principles:
- `internal/domain/` - Business logic entities and rules
- `internal/ports/` - Interface definitions for external dependencies
- `internal/adapters/` - Concrete implementations (docker, http, storage)
- `internal/app/` - Application wiring and orchestration
- `cmd/bosun/main.go` - Application entry point

## Development Workflow
```bash
# Build and run
make build          # Builds to bin/bosun
make run           # Runs from source
make test          # Unit tests
make it            # Integration tests (requires -tags=integration)
make itv           # Integration tests with verbose output

# Code quality
make fmt           # go fmt
make vet           # go vet
make tidy          # go mod tidy
```

## Key Conventions
- **Commit Messages**: See `.github/instructions/commit-msg.instructions.md` for detailed conventional commits format
- **Issue Implementation**: Follow `.github/instructions/issue-implementation.instructions.md` for structured issue handling
- **Code Style**: Follow Go standard conventions; pre-commit hooks enforce formatting
- **Testing**: Write unit tests alongside code; integration tests go in `integration/` with build tags

## Path-Specific Instructions

The repository uses path-specific instructions for context-aware guidance. These are automatically loaded by Copilot based on the files you're working with:

- **Go source files** (`.github/instructions/go-files.instructions.md`): Architecture patterns, error handling, code style
- **Test files** (`.github/instructions/test-files.instructions.md`): Testing patterns, AAA structure, integration test setup
- **Makefile** (`.github/instructions/makefile.instructions.md`): Build target conventions
- **Dockerfile** (`.github/instructions/dockerfile.instructions.md`): Multi-stage builds, security practices
- **Docker Compose** (`.github/instructions/docker-compose.instructions.md`): Testing patterns, label conventions
- **Commit messages** (`.github/instructions/commit-msg.instructions.md`): Conventional commits format

## Serena-first workflow (MANDATORY)

**Always do this at the start of any task, issue, PR review, or chat reply:**
1) Activate Serena on this repo/project → `serena.activate(project="bosun")`.
2) List memories → `serena.memories.list()` and read the most relevant.
3) Update or create memories as needed to save relevant context for future reference.
4) Prefer Serena’s navigation/edit tools for all code work. Only use terminal tools when Serena can’t do it.

**Tool policy**
- Primary: `serena` (code navigation/edits, context).
- Secondary: `context7` (check/lookup updated dependencies or APIs).
- Avoid direct terminal commands unless absolutely necessary; prefer Serena for file ops, search, refactors.

## Project State
Early-stage project with hexagonal architecture scaffolding. Domain logic, ports, and adapters for label discovery are fully implemented. Focus on expanding functionality and integrating additional features.

### Testing Infrastructure
The project includes a comprehensive testing setup:
- **Unit Tests**: Standard Go tests in package-specific `_test.go` files
- **Integration Tests**: Docker Compose-based end-to-end tests in `integration/` package
- **Test Utilities**: `internal/testutil/` provides harness for starting Docker Compose stacks with unique project names and automatic cleanup
- **Test Execution**: Use `make test` for unit tests, `make test-integration` for integration tests (requires Docker)

### Label Discovery Module
The `dockerlabels` adapter provides filtered label discovery for containers, volumes, and networks. It uses `bosun.*` prefix filtering, excludes stopped entities by default, and enriches entities with metadata (containers: image, compose info; volumes: driver; networks: driver, scope). See Serena memories `dockerlabels_adapter` and `label_discovery_domain` for detailed implementation.

## Quality Standards

Before committing changes:
1. **Format**: Run `make fmt` to apply Go formatting
2. **Lint**: Run `make vet` for static analysis
3. **Test**: Run `make test` (unit) and optionally `make it` (integration)
4. **Pre-commit**: Hooks run automatically if installed (`pre-commit install`)

The CI pipeline runs on all PRs and checks:
- All tests pass (unit and integration)
- Code is properly formatted
- No linting issues
- Coverage is maintained

## CLI Structure

Bosun uses Cobra for CLI commands:
- Root command: `internal/cmd/root.go`
- Subcommands organized by feature (e.g., `internal/cmd/labels.go`)
- Main entry point: `cmd/bosun/main.go`

Current commands:
- `bosun labels snapshot [--stopped]` - Capture Docker entity labels

## Dependencies

Key external dependencies:
- **Docker SDK**: `github.com/docker/docker` - Docker client operations
- **Cobra**: CLI framework
- **Testcontainers**: Integration testing with Docker
- **Pre-commit**: Code quality automation

Use `go mod tidy` after adding/removing dependencies.

## Troubleshooting

- **Docker connection issues**: Ensure Docker daemon is running
- **Integration test failures**: Check Docker availability with `docker ps`
- **Pre-commit failures**: Run `pre-commit run --all-files` to see specific issues
- **Build issues**: Ensure Go 1.25+ is installed

## Getting Help

- Check Serena memories for project-specific context
- Review path-specific instructions in `.github/instructions/`
- See `docs/` directory for detailed guides (e.g., `docs/testing.md`)
- Check `README.md` for quick start guide
