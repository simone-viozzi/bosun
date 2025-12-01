# Data Model: Milestone 2.5 – Polish Job Model, Config Schema & Test Suite

**Date**: 2025-11-30
**Status**: Complete

## Overview

This milestone is primarily a **refactoring/polish** effort. No new entities are introduced, but existing entities are clarified and consolidated.

## Entity Changes

### 1. Selector (ports/labels.go)

**Current State**:
```go
type Selector struct {
    Prefixes       []string
    IncludeStopped bool
    ProjectFilter  []string // exists but not implemented
}
```

**Target State**:
```go
type Selector struct {
    Prefixes       []string
    IncludeStopped bool
    ProjectFilter  []string // Filter by com.docker.compose.project label
    StackFilter    []string // Filter by bosun.stack label (NEW)
}
```

**Validation Rules**:
- `ProjectFilter` and `StackFilter` are optional (empty = no filter)
- Multiple values in filter arrays are OR'd (match any)
- If both are specified, results must match both (AND)

---

### 2. JobLabelConfig (config/schema/job_labels.go)

**Current State**: Exists with struct tags defining label metadata.

**Target State**: Unchanged structurally, but becomes the **canonical source** for:
- Label key constants (extracted from tags)
- Default values (extracted from tags)
- Validation rules (derived from type + constraints in tags)

**New Exported Functions** (to be added):
```go
// LabelKeys returns all job label key strings for use in validation
func JobLabelKeys() []string

// DefaultJobSchedule returns the default schedule from JobLabelConfig tag
func DefaultJobSchedule() string

// DefaultWorkerImage returns the default worker image from JobLabelConfig tag
func DefaultWorkerImage() string

// DefaultMountMode returns the default mount mode from JobVolumeConfig tag
func DefaultMountMode() string
```

---

### 3. ValidationError (ports/planner.go)

**Current State**: Exists in ports package.

**Target State**: Unified error type used by both:
- `joblabels.Discoverer.DiscoverJobs()`
- `loader.ValidateJobLabels()`

Both should return `[]ValidationError` with consistent error codes and messages.

---

### 4. Exit Codes (cmd/exitcodes.go - NEW FILE)

**New Constants**:
```go
const (
    ExitSuccess          = 0  // Command completed successfully
    ExitRuntimeError     = 1  // Docker unavailable, I/O failure, etc.
    ExitValidationError  = 2  // Invalid labels, missing required fields
)
```

---

## Relationships

```
┌─────────────────────┐     uses      ┌─────────────────────┐
│  joblabels.Discoverer│─────────────▶│ schema.JobLabelConfig│
└─────────────────────┘               └─────────────────────┘
         │                                     ▲
         │ returns                             │ uses
         ▼                                     │
┌─────────────────────┐               ┌─────────────────────┐
│ ports.ValidationError│◀─────────────│ loader.ValidateJobLabels│
└─────────────────────┘               └─────────────────────┘

┌─────────────────────┐     uses      ┌─────────────────────┐
│ dockerlabels.Source │─────────────▶│  ports.Selector     │
└─────────────────────┘               │  (ProjectFilter,    │
         │                            │   StackFilter)      │
         │ filtered by                └─────────────────────┘
         ▼
┌─────────────────────┐
│ Docker API filters  │
└─────────────────────┘
```

## State Transitions

N/A - This milestone doesn't introduce stateful entities.

## Validation Rules Summary

| Field | Type | Validation | Default |
|-------|------|------------|---------|
| `bosun.job.enabled` | bool | "true"/"false" (case-insensitive) | false |
| `bosun.job.name` | string | Required when enabled | - |
| `bosun.job.schedule` | string | Valid cron expression | "0 0 * * *" |
| `bosun.job.worker.image` | string | Non-empty | "bosun-worker:local" |
| `bosun.job.mount.mode` | enum | "ro"/"rw" (case-insensitive, normalized to lowercase) | "ro" |
