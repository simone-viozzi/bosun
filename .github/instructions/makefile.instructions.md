---
applyTo: "**/Makefile,**/makefile"
---

# Makefile Guidelines

When modifying the Makefile for Bosun:

## Current Targets

The Bosun Makefile provides these standard targets:
- `make build` - Compile to `bin/bosun`
- `make run` - Run from source with `go run`
- `make test` - Run unit tests
- `make it` - Run integration tests (requires Docker)
- `make itv` - Run integration tests with verbose output
- `make fmt` - Format code with `go fmt`
- `make vet` - Run `go vet` for static analysis
- `make tidy` - Run `go mod tidy` to clean dependencies

## Conventions

- Use `.PHONY` for all non-file targets
- Keep targets simple and focused
- Use `@` prefix for commands when output would be noise
- Include helpful comments for complex targets
- Default target should be `build` or `help`

## Adding New Targets

When adding targets:
1. Keep names short and intuitive
2. Document with inline comments if not obvious
3. Follow existing patterns for consistency
4. Use variables for repeated values (e.g., `BINARY_NAME`, `BUILD_DIR`)

## Testing Targets

- Separate unit tests (`test`) from integration tests (`it`)
- Integration tests require Docker - document this clearly
- Consider adding `test-all` target that runs both

## Build Targets

- Build to `bin/` directory (already in .gitignore)
- Use Go build flags for reproducible builds: `-trimpath`
- Consider adding version injection: `-ldflags "-X main.version=$(VERSION)"`

## Quality Targets

- Keep `fmt`, `vet`, `tidy` separate for granular control
- Consider adding a `quality` target that runs all checks
- Pre-commit hooks handle most quality checks automatically

## Example Pattern

```makefile
.PHONY: target-name
target-name: dependencies ## Short description
	@echo "Running target..."
	command-to-run
```

## Don't

- Don't make targets OS-specific without good reason
- Don't use overly complex shell scripts (move to separate scripts instead)
- Don't duplicate functionality that's in pre-commit hooks
