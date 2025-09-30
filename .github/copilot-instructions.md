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
make test-integration  # Integration tests (requires -tags=integration)

# Code quality
make fmt           # go fmt
make vet           # go vet
make tidy          # go mod tidy
```

## Key Conventions
- **Commit Messages**: See `.github/instructions/commit-msg.instructions.md` for detailed conventional commits format
- **Branching**: Semi-linear history enforced - rebase PRs, no merge commits
- **Testing**: Integration tests use build tag `//go:build integration`
- **Dependencies**: Managed via Renovate with standard config
- **CI**: Tests + golangci-lint on PRs to main

## Development Environment
Use the provided `shell.nix` for consistent tooling:
- Go 1.24.6 with gopls, delve, gofumpt
- Docker/Colima for container development
- Pre-commit hooks for quality gates

## Project State
Early-stage project with hexagonal architecture scaffolding. Focus on implementing domain logic, ports, and adapters following the established structure.
