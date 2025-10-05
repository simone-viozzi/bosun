# Code Style and Conventions

## Go Standards
- Follow the official Go formatting standards using `go fmt`
- Use `gofmt` for consistent code formatting
- Adhere to Go naming conventions:
  - Exported identifiers start with capital letters
  - Unexported with lowercase
  - Use camelCase for multi-word identifiers
- Use meaningful, descriptive names for variables, functions, and types

## Code Quality
- Run `go vet` for static analysis to catch common mistakes
- Use golangci-lint for comprehensive linting
- Write clear, concise code with appropriate comments
- Use type hints and interfaces effectively

## Architecture Specific
- Follow hexagonal architecture principles:
  - Business logic in `internal/domain/`
  - Interfaces (ports) in `internal/ports/`
  - External adapters in `internal/adapters/`
- Keep dependencies pointing inward (domain doesn't depend on adapters)

## Documentation
- Use comments for exported functions and types
- Keep TODO.md updated for tasks and FIXME items
- Use pre-commit hook for TODO management

## File Organization
- One package per directory
- Main entrypoint in `cmd/bosun/main.go`
- Internal code in `internal/` subdirectories
- Test files alongside source files with `_test.go` suffix
