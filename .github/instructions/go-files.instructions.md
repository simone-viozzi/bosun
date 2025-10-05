---
applyTo: "**/*.go,!**/*_test.go"
---

# Go Source Code Guidelines

When working with Go source files in the Bosun project, follow these conventions:

## Architecture Principles

Bosun follows **hexagonal/clean architecture**:
- **Domain layer** (`internal/domain/`): Pure business logic, no external dependencies
- **Ports layer** (`internal/ports/`): Interface definitions for external systems
- **Adapters layer** (`internal/adapters/`): Concrete implementations of ports
- **Application layer** (`internal/app/`): Wires everything together

**Dependency Rule**: Dependencies point inward. Domain never depends on adapters or ports.

## Code Style

- Follow standard Go conventions (use `gofmt`, `goimports`)
- Keep exported identifiers capitalized, unexported lowercase
- Use camelCase for multi-word identifiers
- Write clear, descriptive names (avoid abbreviations unless widely understood)
- Add godoc comments for all exported types, functions, and constants
- Keep functions focused and single-purpose (prefer small functions)

## Error Handling

- Return errors, don't panic (except for truly unrecoverable situations)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Check all errors (don't ignore with `_`)
- Provide user-friendly error messages for CLI commands

## Package Organization

- One package per directory
- Keep internal packages in `internal/` to prevent external imports
- Use package names that are short, lowercase, single-word when possible
- Avoid generic names like `util` or `common`

## Interfaces

- Define interfaces in the consumer package (ports), not the implementer
- Keep interfaces small (1-3 methods is ideal)
- Name single-method interfaces with `-er` suffix (e.g., `Reader`, `Writer`)

## Context Usage

- Pass `context.Context` as first parameter when needed
- Use context for cancellation, timeouts, and request-scoped values
- Don't store contexts in structs (pass them through function calls)

## Before Committing

Run these commands:
```bash
make fmt    # Format code
make vet    # Static analysis
make test   # Unit tests
```

Pre-commit hooks will also run automatically if installed.
