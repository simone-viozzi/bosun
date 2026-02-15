# Quickstart: Milestone 3.75 — Critical Bug Fixes

**Branch**: `011-critical-bug-fixes`

## Overview

This milestone fixes 3 bugs. No new features, no API changes. All changes are backward-compatible.

## Changes At A Glance

### 1. Fix Corrupted Worker Logs (#153)

**Before**: Worker output shows binary garbage: `\x01\x00\x00\x00\x00\x00\x00\x0cHello World`
**After**: Worker output is clean: `Hello World`

**What changes**:
- `internal/adapters/docker/worker/runner.go` → `streamLogs()`: Replace `io.Copy` with `stdcopy.StdCopy`
- `internal/testutil/docker.go` → `DumpLogs()`: Same fix

### 2. Fix Worker Error Handling (#144)

**Before**: Errors silently dropped; exit code `137` hardcoded as magic number.
**After**: Errors logged/returned; exit codes use named constants.

**What changes**:
- `streamLogs()`: Returns `error` instead of silently swallowing
- `stopContainer()`: Returns `(int, error)` instead of defaulting to magic `137`
- New constants: `ExitCodeSIGKILL = 137`, `ExitCodeSIGTERM = 143`

### 3. Move Coverage Files (#161)

**Before**: `coverage.out`, `coverage.html`, etc. clutter the repo root.
**After**: All coverage output goes to `coverage/` subdirectory.

**What changes**:
- `Makefile`: Add `COVERAGE_DIR` variable, prefix all coverage paths
- `.gitignore`: Simplify to single `coverage/` entry

## Verification

```bash
# After fix #153 + #144: Run any job and check output is clean
bosun job run --job my-job

# After fix #161: Run tests with coverage and check root is clean
make coverage
ls coverage/        # All output files here
ls coverage*.* 2>/dev/null  # Should be empty
```

## No Breaking Changes

- `WorkerRunner` port interface: unchanged
- `WorkerConfig` / `WorkerResult` structs: unchanged
- CLI flags and exit codes: unchanged
- All existing tests should pass without modification
