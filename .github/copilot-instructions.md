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
- **Branching**: Semi-linear history enforced - rebase PRs, no merge commits
- **Testing**: Integration tests use build tag `//go:build integration`
- **Tools**: Use serena as main tool, use context7 to get updated dependencies. Do not use the terminal tools directly unless absolutely necessary. Prefer using serena to navigate the codebase.
- **Memories**: Always list and read memories upon starting a task. Then create and update Serena memories for important concepts, patterns, and decisions.

## Project State
Early-stage project with hexagonal architecture scaffolding. Focus on implementing domain logic, ports, and adapters following the established structure.

### Testing Infrastructure
The project includes a comprehensive testing setup:
- **Unit Tests**: Standard Go tests in package-specific `_test.go` files
- **Integration Tests**: Docker Compose-based end-to-end tests in `integration/` package
- **Test Utilities**: `internal/testutil/` provides harness for starting Docker Compose stacks with unique project names and automatic cleanup
- **Test Execution**: Use `make test` for unit tests, `make test-integration` for integration tests (requires Docker)

### Label Discovery Module
The `dockerlabels` adapter provides filtered label discovery for containers, volumes, and networks. It uses `bosun.*` prefix filtering, excludes stopped entities by default, and ignores image labels in v1.
