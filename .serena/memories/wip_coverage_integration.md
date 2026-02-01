# WIP: Test Coverage Integration (Issue #150)

## Current Phase: Integration Test Coverage (Phase 2) — COMPLETED ✅

### Phase 1 Completed
- ✅ Add `make coverage` and `make coverage-html` targets (tested locally, 11% coverage baseline)
- ✅ Update CI to use `-coverpkg=./...` for cross-package coverage
- ✅ Generate and upload HTML coverage artifacts in CI
- ✅ Add VS Code coverage settings (opt-in, gutter-style decorators)
- ✅ Update .gitignore for coverage files

### Phase 2 Completed ✅
- ✅ Created `internal/testutil/coverage.go` with `BuildBosunBinary()` and `RunBosunWithCoverage()` helpers
- ✅ Added `coverageEnabled()` helper and global `globalCoverDir` in `integration/main_test.go`
- ✅ Updated all integration tests (validate, plan, job_execution) to use `runBosun()` wrapper that auto-enables coverage via `BOSUN_TEST_COVERAGE=1` env var
- ✅ Added Makefile targets:
  - `make it-cover` — runs integration tests with coverage, converts to `coverage.integration.out`
  - `make coverage-integration` — standalone target to convert `covdata/` to profile
  - `make coverage-all` — merges unit + integration coverage, generates unified HTML report
- ✅ Updated CI `.github/workflows/ci.yml` integration job:
  - Collects integration coverage with `GOCOVERDIR`
  - Downloads unit coverage artifacts from test job
  - Merges profiles with `gocovmerge`
  - Uploads merged `coverage.all.out` to Codecov (Option C: clean dashboard)
  - Archives all three profiles (unit, integration, merged) as artifacts for debugging
- ✅ Updated `.gitignore` for `covdata/`, `coverage.all.html`, `coverage.all.txt`

**Implementation Notes:**
- Used Go 1.20+ binary coverage: build with `-cover`, run with `GOCOVERDIR`, convert with `go tool covdata textfmt`
- Global `covdata/` directory shared across parallel tests (Go handles merging automatically)
- Each test builds its own binary (per-test isolation for parallel safety)
- Backward compatible: `make it` still runs without coverage (fast), `make it-cover` opts in

### Phase 3: Coverage Thresholds/Gating (Future)
**Approach Used:** Go 1.20+ binary coverage (Advanced method), because integration tests execute the `bosun` binary via `exec.Command()`.

### Phase 3: Coverage Thresholds/Gating (Future)
Add CI step to fail build if coverage drops below threshold using `vladopajic/go-test-coverage` action.

### Phase 4: Enhanced Visualization (Future)
Consider Codecov PR comments, coverage badges, or other UI enhancements.

## References
- Go coverage docs: https://go.dev/doc/build-cover
- Integration test coverage: https://go.dev/blog/integration-test-coverage
- Current CI: `.github/workflows/ci.yml` already uploads to Codecov
- VS Code extension: golang.go (confirmed installed)
