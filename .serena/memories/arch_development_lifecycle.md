# Development Lifecycle

## Scope
Commands, workflows, and conventions for developing and testing Bosun.

## What

### Build & Run
- `make build` - Compile to `bin/bosun`
- `make run` - Run with `go run`
- `docker build -t bosun .` - Container build

### Testing
- `make test` - Unit tests (`go test ./...`)
- `make it` - Integration tests (requires Docker)
- `make itv` - Verbose integration tests
- Integration tests use `//go:build integration` tag

### Code Quality
- `make fmt` - Format code (`go fmt ./...`)
- `make vet` - Static analysis (`go vet ./...`)
- `make tidy` - Dependency management (`go mod tidy`)
- `pre-commit run --all-files` - All quality checks
- `golangci-lint run` - Comprehensive linting

### Documentation
- `make docs` - Generate config docs and JSON schema
- Outputs: `docs/config.md`, `docs/config.schema.json`

### Task Completion Checklist
After any development task:
1. `make fmt` - Format
2. `make vet` - Analyze
3. `make test` - Unit tests
4. `pre-commit run --all-files` - Hooks
5. `make tidy` - If deps changed
6. Update TODO.md if needed

### CI/CD
- GitHub Actions on push/PR to `main`
- Test coverage uploaded to Codecov
- Linting with golangci-lint

## Why
Consistent workflows ensure code quality and prevent CI failures.

## Related
- `arch_overview` - Architecture context
- `arch_testing` - Testing infrastructure details
