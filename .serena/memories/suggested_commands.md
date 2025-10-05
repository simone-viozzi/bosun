# Suggested Commands for Development

## Building and Running
- **Build the application**: `make build` - Compiles the Go code into `bin/bosun`
- **Run the application**: `make run` - Runs the application directly with `go run`
- **Build and run in Docker**: `docker build -t bosun . && docker run bosun`

## Testing and Quality
- **Run all tests**: `make test` - Executes `go test ./...`
- **Run tests with coverage**: `go test -v ./... -coverprofile=coverage.out`
- **Format code**: `make fmt` - Applies `go fmt ./...`
- **Static analysis**: `make vet` - Runs `go vet ./...`
- **Lint code**: Use golangci-lint (via CI or locally installed)
- **Tidy dependencies**: `make tidy` - Runs `go mod tidy`

## Pre-commit and Git
- **Run pre-commit hooks**: `pre-commit run --all-files` - Formats, vets, and checks code quality
- **Install pre-commit hooks**: `pre-commit install`

## Utility Commands (Linux)
- **List files**: `ls -la`
- **Change directory**: `cd <path>`
- **Search text**: `grep -r "pattern" .`
- **Find files**: `find . -name "*.go"`
- **Git status**: `git status`
- **Git add**: `git add .`
- **Git commit**: `git commit -m "message"`
- **Git push**: `git push origin <branch>`

## CI/CD
- CI runs automatically on pushes/PRs to `main` and `init` branches
- Includes test execution with coverage upload to Codecov
- Linting with golangci-lint
