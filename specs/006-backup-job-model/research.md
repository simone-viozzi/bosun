# Research: Job Model & Planning

**Feature Branch**: `006-backup-job-model`
**Date**: 2025-11-30

## Research Tasks

This document resolves all NEEDS CLARIFICATION items and documents technology decisions.

---

## 1. Cron Expression Parsing Library

**Task**: Select a Go library for parsing and validating cron expressions.

### Decision: `github.com/robfig/cron/v3`

### Rationale

- **Maturity**: Industry-standard Go cron library with high GitHub stars and extensive production use
- **Parser Flexibility**: Supports standard 5-field cron syntax (minute, hour, dom, month, dow) with optional customization
- **Validation**: `cron.ParseStandard()` returns errors for invalid expressions—exactly what we need for validation
- **No Scheduling Required**: For this milestone, we only need parsing/validation, not scheduling. The library's `Parser` can be used standalone.
- **Predefined Schedules**: Supports `@daily`, `@hourly`, `@weekly`, `@monthly`, `@yearly`, `@every <duration>` for user convenience

### Alternatives Considered

| Library | Rejected Because |
|---------|------------------|
| `github.com/go-co-op/gocron` | Higher-level scheduler; we only need parsing |
| Custom regex | Error-prone; cron has edge cases (day-of-week vs day-of-month) |
| `github.com/gorhill/cronexpr` | Less maintained; robfig/cron is the de facto standard |

### Usage Pattern

```go
import "github.com/robfig/cron/v3"

// Validate cron expression (standard 5-field: minute hour dom month dow)
parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
_, err := parser.Parse("0 2 * * *")
if err != nil {
    return fmt.Errorf("invalid schedule: %w", err)
}
```

---

## 2. Domain Model Design Pattern

**Task**: Best practices for Go domain models in hexagonal architecture.

### Decision: Pure Value Types + Factory Functions

### Rationale

Following existing Bosun patterns from `internal/domain/labels/`:

1. **Pure Structs**: Domain types are plain Go structs with no external dependencies
2. **Exported Fields**: Use exported fields (no getters/setters) for simplicity; Go convention
3. **Factory Functions**: `NewJob()`, `NewExecutionPlan()` for construction with validation
4. **No Methods on Domain Types**: Keep domain types as data containers; logic lives in service/planner layer
5. **Immutability**: Prefer returning new instances over mutation

### Pattern Example

```go
// internal/domain/jobs/types.go
package jobs

import "time"

type Job struct {
    Name             string
    Schedule         string           // Validated cron expression
    TargetContainers []string         // Container IDs
    TargetStacks     []string         // Derived stack names
    WorkerImage      string
    AttachVolumes    []VolumeAttachment
    SourceContainers []string         // Containers that contributed labels
}

type VolumeAttachment struct {
    Name string
    Mode string // "ro" or "rw"
}
```

---

## 3. Docker Compose Dependency Label Format

**Task**: Verify format of `com.docker.compose.depends_on` container label.

### Decision: Confirmed via User Research

### Format

```
com.docker.compose.depends_on=<service>:<condition>:<required>,...
```

**Example**: `db:service_started:true,redis:service_healthy:false`

### Fields

| Field | Description | Values |
|-------|-------------|--------|
| `service` | Service name from compose.yml | string |
| `condition` | Start condition | `service_started`, `service_healthy`, `service_completed_successfully` |
| `required` | Whether dependency is required | `true`, `false` |

### Usage in Bosun

1. Parse label from container metadata
2. Extract service names (first field before each `:`)
3. Cross-reference with job's target containers
4. If any dependent service is NOT in the target list, fail validation

### Edge Cases

- Label may be absent for non-Compose containers → skip dependency validation
- Multiple dependencies are comma-separated
- Label is on the **dependent** container, not the dependency

---

## 4. Job Label Schema Design

**Task**: Design label schema following existing `FieldSpec` patterns.

### Decision: Extend Schema Package with Job Labels

### New Labels

| Label Key | Type | Scope | Default | Required | Description |
|-----------|------|-------|---------|----------|-------------|
| `bosun.job.enabled` | bool | container | `false` | No | Marks container as job definition |
| `bosun.job.name` | string | container | — | **Yes** (if enabled) | Unique job identifier |
| `bosun.job.schedule` | string | container | `0 0 * * *` | No | Cron expression |
| `bosun.job.worker.image` | string | container | (built-in) | No | Worker container image |
| `bosun.job.attach` | string | volume | — | No | Job attachment (`<job>` or `<job>:ro`/`<job>:rw`) |

### Schema Implementation

```go
// internal/config/schema/job_labels.go
type JobLabelConfig struct {
    Enabled     bool   `bosun:"key=bosun.job.enabled,scope=container,type=bool,default=false,doc='Enable job definition'"`
    Name        string `bosun:"key=bosun.job.name,scope=container,type=string,required,doc='Job identifier'"`
    Schedule    string `bosun:"key=bosun.job.schedule,scope=container,type=string,default=0 0 * * *,doc='Cron schedule'"`
    WorkerImage string `bosun:"key=bosun.job.worker.image,scope=container,type=string,doc='Worker container image'"`
}

type JobVolumeConfig struct {
    Attach string `bosun:"key=bosun.job.attach,scope=volume,type=string,doc='Job to attach volume to'"`
}
```

---

## 5. Default Worker Image Strategy

**Task**: Define default worker image for jobs without explicit `bosun.job.worker.image`.

### Decision: Built-in Placeholder (Milestone 2 Only)

### Rationale

- This milestone introduces no Docker side effects
- The default image value is stored but never executed
- Using `bosun-worker:local` as placeholder documents intent
- Actual image publishing (`ghcr.io/simone-viozzi/bosun-worker`) deferred to Milestone 3

### Implementation

```go
const DefaultWorkerImage = "bosun-worker:local" // Placeholder, not executed in M2
```

---

## 6. Planner Determinism Guarantee

**Task**: Ensure planner produces identical output for identical inputs.

### Decision: Deterministic Sorting + No Side Effects

### Guarantees

1. **Sort entities**: Containers and volumes sorted by ID before processing
2. **Stable map iteration**: Use `sort.Strings(keys)` before iterating maps
3. **No timestamps in plan**: `ExecutionPlan.CreatedAt` set by caller, not planner
4. **No random values**: No UUIDs or random tokens in plan generation

### Verification

Unit tests will compare JSON serialization of plans generated from identical inputs.

---

## 7. CLI Output Format Strategy

**Task**: Design output formats for `plan list` and `plan show`.

### Decision: Follow Existing CLI Patterns

### Formats

| Format | Use Case | Implementation |
|--------|----------|----------------|
| `text` (default) | Human-readable, colorized terminal output | Custom formatting |
| `json` | Scripting, automation, piping to `jq` | `encoding/json` |
| `yaml` | Human-readable structured output | `gopkg.in/yaml.v3` |

### Pattern (from `bosun labels snapshot`)

```go
switch format {
case "json":
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(plan)
case "yaml":
    enc := yaml.NewEncoder(os.Stdout)
    return enc.Encode(plan)
default: // text
    return renderTextPlan(plan)
}
```

---

## Summary

All research tasks complete. No remaining NEEDS CLARIFICATION items.

| Research Item | Status | Decision |
|---------------|--------|----------|
| Cron library | ✅ Resolved | `github.com/robfig/cron/v3` |
| Domain model pattern | ✅ Resolved | Pure value types + factory functions |
| Compose depends_on label | ✅ Resolved | `service:condition:required,...` format |
| Job label schema | ✅ Resolved | Extend schema package with job labels |
| Default worker image | ✅ Resolved | `bosun-worker:local` placeholder |
| Planner determinism | ✅ Resolved | Sorted iteration, no side effects |
| CLI output formats | ✅ Resolved | text/json/yaml following existing patterns |
