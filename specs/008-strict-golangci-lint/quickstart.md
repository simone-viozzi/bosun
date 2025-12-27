# Quickstart: Strict golangci-lint Checks

**Feature**: 008-strict-golangci-lint
**Date**: 2025-12-27
**Purpose**: Commands to verify implementation and run lint checks

## Prerequisites

```bash
# Ensure golangci-lint is installed
golangci-lint --version
# Expected: golangci-lint has version v2.x.x or higher
```

## Verification Commands

### 1. Run Full Lint Suite

```bash
# Primary success criteria - must exit 0 with no issues
golangci-lint run ./...
```

### 2. Verify Configuration

```bash
# Config validation - must pass without errors
golangci-lint config verify
```

### 3. Check Specific Linters (for debugging)

```bash
# Check prealloc specifically
golangci-lint run --enable prealloc ./...

# Check unparam specifically
golangci-lint run --enable unparam ./...

# Check errcheck specifically
golangci-lint run --enable errcheck ./...
```

### 4. Run with Integration Tests Included

```bash
# Include integration test files in lint scope
golangci-lint run --build-tags integration ./...
```

### 5. Verify No Disabled Linters

```bash
# Should NOT contain 'disable:' section with active linters
grep -A5 "disable:" .golangci.yml

# Should NOT contain -ST exclusions
grep "ST10" .golangci.yml
```

## Expected Final State

After implementation, the following must be true:

1. `golangci-lint run` exits with code 0
2. `golangci-lint config verify` passes
3. `.golangci.yml` has no linters in `disable:` section
4. `.golangci.yml` has no `-ST*` exclusions in staticcheck checks
5. All tests still pass: `make test`

## Troubleshooting

### If prealloc violations appear

Check if slice can be pre-allocated:
```go
// Use make() with capacity
out := make([]T, 0, len(source))
```

### If unparam violations appear

Either:
- Remove the unused parameter
- Change return type if error is always nil
- Use `//nolint:unparam` with justification (last resort)

### If ST1000 violations appear

Create `doc.go` with package comment:
```go
// Package name provides description.
package name
```

### If unusedwrite violations appear in tests

Either:
- Add assertion using the field
- Remove the unnecessary field assignment
- Use exclusion rule for test files (if pattern is intentional)
