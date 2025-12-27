# Code Style & Conventions

## Scope
Go coding standards and project conventions.

## What

### Go Standards
- `go fmt` for formatting
- `go vet` for static analysis
- `golangci-lint` for comprehensive linting
- Naming: exported = Capital, unexported = lowercase, camelCase

### Architecture Rules
- Business logic in `internal/domain/` - no external dependencies
- Interfaces in `internal/ports/` - contracts only
- Implementations in `internal/adapters/` - external integrations
- Dependencies point inward (domain ← ports ← adapters)

### File Organization
- One package per directory
- Test files alongside source (`*_test.go`)
- Main entrypoint: `cmd/bosun/main.go`
- All internal code under `internal/`

### Documentation
- Comments for exported functions and types
- TODO.md for tasks (auto-generated from todos at commit)
- Commit messages follow `.github/instructions/commit-msg.instructions.md`

### Error Handling
- Return all validation errors, not just first
- Use typed errors (e.g., `ValidationError`, `ValidationErrors`)
- Standardized exit codes in CLI

## Why
Consistent style enables collaboration and maintainability.

## Related
- `arch_development_lifecycle` - Commands to enforce style
- `arch_overview` - Architecture context
