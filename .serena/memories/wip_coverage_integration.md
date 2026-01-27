# WIP: Test Coverage Integration (Issue #150)

## Current Phase: Basic Coverage Setup

### Completed
- ✅ Add `make coverage` and `make coverage-html` targets (tested locally, 11% coverage baseline)
- ✅ Update CI to use `-coverpkg=./...` for cross-package coverage
- ✅ Generate and upload HTML coverage artifacts in CI
- ✅ Add VS Code coverage settings (opt-in, gutter-style decorators)
- ✅ Update .gitignore for coverage files

### In Progress
None - basic phase complete.

## Future Phases (Out of Scope for Initial PR)

### Phase 2: Integration Test Coverage
Integration tests run with `-tags=integration` in `integration/` package. Two approaches:

1. **Simple**: If tests are still `_test.go` files, run with coverage:
   ```bash
   go test -tags=integration ./integration/... -coverprofile=coverage.integration.out -coverpkg=./...
   ```

2. **Advanced** (Go 1.20+): For E2E tests that run a binary:
   - Build with `go build -cover -o bin/bosun ./cmd/bosun`
   - Run with `GOCOVERDIR=covdata bin/bosun ...`
   - Convert with `go tool covdata textfmt -i=covdata -o coverage.e2e.out`

### Phase 3: Merge Multiple Coverage Profiles
Use `gocovmerge` to combine unit + integration coverage:
```bash
go install github.com/wadey/gocovmerge@latest
gocovmerge coverage.out coverage.integration.out > coverage.all.out
```

### Phase 4: Coverage Thresholds/Gating
Add CI step to fail build if coverage drops below threshold using `vladopajic/go-test-coverage` action.

### Phase 5: Enhanced Visualization
Consider Codecov PR comments, coverage badges, or other UI enhancements.

## References
- Go coverage docs: https://go.dev/doc/build-cover
- Integration test coverage: https://go.dev/blog/integration-test-coverage
- Current CI: `.github/workflows/ci.yml` already uploads to Codecov
- VS Code extension: golang.go (confirmed installed)
