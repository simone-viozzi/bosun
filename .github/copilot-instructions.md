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
The `dockerlabels` adapter provides filtered label discovery for containers, volumes, and networks with `bosun.*` prefix filtering, stopped container exclusion by default, and metadata enrichment. **Key conventions:**
- Labels are case-sensitive
- Image labels are intentionally ignored in v1
- Networks may require manual label application
- Only entities with matching labels are included

See `internal/adapters/dockerlabels/README.md` for comprehensive documentation, examples, and gotchas. Serena memories `dockerlabels_adapter` and `label_discovery_domain` contain implementation details.
