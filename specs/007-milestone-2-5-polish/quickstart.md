# Quickstart: Milestone 2.5 Implementation

**Date**: 2025-11-30

## Overview

This document provides a quick reference for implementing the Milestone 2.5 polish work.

## Key Changes Summary

### 1. Unify Validation Logic (P1)

**Goal**: Single source of truth for job validation rules.

**Files to modify**:
- `internal/config/schema/job_labels.go` - Add helper functions to extract constants from tags
- `internal/adapters/joblabels/discoverer.go` - Import and use schema helpers
- `internal/config/loader/job_validation.go` - Import and use schema helpers

**Pattern**:
```go
// In schema/job_labels.go - export constants from tags
func JobLabelKey(field string) string { ... }
func DefaultSchedule() string { return "0 0 * * *" }

// In discoverer.go - use schema constants
import "github.com/simone-viozzi/bosun/internal/config/schema"
if labels[schema.JobLabelKey("enabled")] == "true" { ... }
```

### 2. Implement Project/Stack Filtering (P1)

**Goal**: Isolate CLI operations to specific Docker Compose projects.

**Files to modify**:
- `internal/ports/labels.go` - Add `StackFilter` field to `Selector`
- `internal/adapters/dockerlabels/source.go` - Implement Docker API filtering
- `internal/cmd/plan_list.go` - Add `--project` and `--stack` flags
- `internal/cmd/plan_show.go` - Add `--project` and `--stack` flags
- `internal/cmd/snapshot.go` - Add `--project` and `--stack` flags

**Docker filtering pattern**:
```go
import "github.com/docker/docker/api/types/filters"

args := filters.NewArgs()
if len(selector.ProjectFilter) > 0 {
    for _, p := range selector.ProjectFilter {
        args.Add("label", "com.docker.compose.project="+p)
    }
}
if len(selector.StackFilter) > 0 {
    for _, s := range selector.StackFilter {
        args.Add("label", "bosun.stack="+s)
    }
}
```

### 3. Update CLI Branding (P2)

**Files to modify**:
- `internal/cmd/root.go` - Update Short/Long descriptions
- `internal/cmd/plan.go` - Update subcommand descriptions
- `internal/cmd/config.go` - Update subcommand descriptions

### 4. Centralize Exit Codes (P2)

**New file**: `internal/cmd/exitcodes.go`

```go
package cmd

const (
    ExitSuccess         = 0
    ExitRuntimeError    = 1
    ExitValidationError = 2
)
```

**Update all commands** to use these constants instead of hardcoded values.

### 5. Document Testing Philosophy (P3)

**New file**: `internal/testutil/doc.go`

```go
// Package testutil provides utilities for testing Bosun.
//
// # Test Layer Responsibilities
//
// Unit tests (internal/*_test.go):
//   - Exhaustive edge cases with mocked inputs
//   - No Docker dependency
//   - Fast execution (<1s per test)
//
// Integration tests (integration/):
//   - Happy path end-to-end flows
//   - Real Docker Compose stacks
//   - CLI invocation and output verification
//
// # Adding New Tests
//
// For new validation edge cases: add to unit tests
// For new CLI flows: add to integration tests
// For new adapters: unit test with mocks, integration test for smoke
package testutil
```

### 6. Update Compose Fixtures (P3)

**Add header comments** to each file in `internal/testutil/compose/`:

```yaml
# docker-compose.yaml
# Purpose: Basic nginx service for smoke testing infrastructure
# Used by: integration/smoke_placeholder_test.go
```

### 7. Update README and Docs (P3)

**Run**: `go generate ./internal/config/schema/...`

**Verify**: `docs/config.md` includes job labels (`bosun.job.*`)

**Update**: README.md examples if any are stale

## Testing Checklist

- [ ] Unit tests pass: `make test`
- [ ] Integration tests pass: `make it`
- [ ] New filter flags work: `bosun plan list --project myproject`
- [ ] Empty filter returns exit 0: `bosun plan list --project nonexistent`
- [ ] Validation errors return exit 2
- [ ] Help text mentions backup/orchestration
- [ ] `go generate` produces up-to-date docs
