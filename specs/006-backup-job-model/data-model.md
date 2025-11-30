# Data Model: Job Model & Planning

**Feature Branch**: `006-backup-job-model`
**Date**: 2025-11-30

## Overview

This document defines the domain entities for the Job Model & Planning feature. All types are pure Go structs following hexagonal architecture—no external dependencies in the domain layer.

---

## Entity Definitions

### Job

Represents a discovered job definition assembled from Docker container labels.

```go
// Package: internal/domain/jobs

// Job represents a discovered job assembled from container labels.
// A job can be contributed to by multiple containers (merged by name).
type Job struct {
    // Name is the unique identifier for this job (from bosun.job.name).
    // Required field.
    Name string

    // Schedule is a validated cron expression (from bosun.job.schedule).
    // Default: "0 0 * * *" (daily at midnight).
    Schedule string

    // TargetContainers lists container IDs that participate in this job.
    // These containers will be stopped before the worker runs.
    TargetContainers []string

    // TargetStacks lists unique stack names derived from TargetContainers.
    // Used for display/grouping purposes only; actual stop is per-container.
    TargetStacks []string

    // WorkerImage specifies the container image for executing the job.
    // Default: "bosun-worker:local" (placeholder for this milestone).
    WorkerImage string

    // AttachVolumes lists volumes to mount in the worker container.
    // Discovered from volumes with bosun.job.attach=<this-job> label.
    AttachVolumes []VolumeAttachment

    // SourceContainers tracks which containers contributed to this job.
    // Used for traceability and conflict detection.
    SourceContainers []string
}
```

**Validation Rules**:
- `Name` must be non-empty
- `Schedule` must be a valid cron expression (validated via `robfig/cron/v3`)
- `TargetContainers` may be empty (job discovered but no containers currently running)
- If multiple containers set the same field with conflicting values → validation error

---

### VolumeAttachment

Represents a volume to be mounted in the worker container.

```go
// VolumeAttachment represents a volume attached to a job.
type VolumeAttachment struct {
    // Name is the Docker volume name.
    Name string

    // MountPath is the path where the volume is mounted in the worker.
    // Default: "/data/<volume-name>" if not specified.
    MountPath string

    // Mode is the mount access mode: "ro" (read-only) or "rw" (read-write).
    // Default: "ro" for safety.
    Mode string
}
```

**Validation Rules**:
- `Name` must be non-empty
- `Mode` must be one of: `"ro"`, `"rw"`

---

### ExecutionPlan

An ordered sequence of steps to execute a job.

```go
// ExecutionPlan represents the computed steps to execute a job.
// Plans are deterministic: same inputs produce identical outputs.
type ExecutionPlan struct {
    // JobName references the originating job.
    JobName string

    // Steps contains ordered actions to execute.
    Steps []PlanStep

    // CreatedAt records when the plan was generated.
    // Set by the caller, not the planner (for determinism).
    CreatedAt time.Time
}
```

**Invariants**:
- `Steps` order is significant (stop before run-worker)
- Plans are immutable once created

---

### PlanStep

An individual action in an execution plan.

```go
// StepType identifies the kind of action in a plan step.
type StepType string

const (
    // StepTypeStopContainers stops the target containers.
    StepTypeStopContainers StepType = "stop_containers"

    // StepTypeRunWorker runs the worker container with attached volumes.
    StepTypeRunWorker StepType = "run_worker"

    // StepTypeStartContainers restarts the stopped containers.
    // Note: Out of scope for Milestone 2; included for forward compatibility.
    StepTypeStartContainers StepType = "start_containers"
)

// PlanStep represents a single action in an execution plan.
type PlanStep struct {
    // Type identifies the kind of action.
    Type StepType

    // Description is a human-readable explanation of the step.
    Description string

    // ContainerIDs lists containers affected by this step (for stop/start).
    // Empty for run_worker steps.
    ContainerIDs []string

    // ContainerNames provides human-readable names for ContainerIDs.
    ContainerNames []string

    // WorkerImage is the image to run (for run_worker steps only).
    WorkerImage string

    // VolumeMounts lists volumes to attach (for run_worker steps only).
    VolumeMounts []VolumeAttachment

    // UseComposeStop indicates if `docker compose stop` should be used.
    // True when all containers in a stack are being stopped.
    UseComposeStop bool

    // ComposeProject is the project name for compose commands.
    ComposeProject string
}
```

**Step Examples**:

| Type | Description Example |
|------|---------------------|
| `stop_containers` | "Stop 3 containers in stack 'myapp': web, api, worker" |
| `run_worker` | "Run worker 'restic/restic:latest' with 2 volumes attached" |
| `start_containers` | "Restart 3 containers in stack 'myapp'" |

---

### Stack

A logical grouping of Docker entities.

```go
// Stack represents a logical grouping of Docker entities.
// Determined by bosun.stack label or com.docker.compose.project metadata.
type Stack struct {
    // Name is the stack identifier.
    Name string

    // Containers lists container IDs in this stack.
    Containers []string

    // Volumes lists volume names associated with this stack.
    Volumes []string

    // Networks lists network names associated with this stack.
    Networks []string
}
```

**Resolution Rules**:
1. If `bosun.stack` label is present → use that value
2. Else if `com.docker.compose.project` metadata is present → use that value
3. Else → container is not part of any stack

---

### ContainerDependency

Parsed dependency information from Compose labels.

```go
// ContainerDependency represents a parsed dependency from
// the com.docker.compose.depends_on container label.
type ContainerDependency struct {
    // ServiceName is the dependency service name.
    ServiceName string

    // Condition is the start condition (service_started, service_healthy, etc.).
    Condition string

    // Required indicates if the dependency is required.
    Required bool
}
```

---

## Relationships

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Discovery Flow                             │
│                                                                     │
│  Snapshot                                                           │
│  (LabeledEntities)                                                  │
│        │                                                            │
│        ▼                                                            │
│  ┌─────────────┐    merge by    ┌─────────────┐                     │
│  │  Container  │───────────────▶│    Job      │                     │
│  │  Labels     │    job.name    │             │                     │
│  └─────────────┘                └─────────────┘                     │
│        │                              │                             │
│        │                              │                             │
│        ▼                              ▼                             │
│  ┌─────────────┐               ┌─────────────────┐                  │
│  │   Volume    │──────────────▶│ VolumeAttachment│                  │
│  │   Labels    │  job.attach   │                 │                  │
│  └─────────────┘               └─────────────────┘                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                          Planning Flow                              │
│                                                                     │
│  ┌─────────────┐                ┌─────────────────┐                 │
│  │    Job      │───────────────▶│  ExecutionPlan  │                 │
│  │             │    Planner     │                 │                 │
│  └─────────────┘                └────────┬────────┘                 │
│                                          │                          │
│                                          │                          │
│                                          ▼                          │
│                                 ┌─────────────────┐                 │
│                                 │    PlanStep     │ ×N              │
│                                 │                 │                 │
│                                 └─────────────────┘                 │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## State Transitions

Jobs and plans do not have explicit state machines in this milestone. Jobs are discovered (stateless), plans are generated (immutable).

Future milestones may introduce:
- Job execution state: `pending` → `running` → `completed` | `failed`
- Plan execution tracking

---

## Indexes / Lookups

For efficient operations, adapters should support:

| Lookup | Description | Implementation |
|--------|-------------|----------------|
| Jobs by name | Find job by unique name | `map[string]Job` |
| Containers by stack | Group containers by stack name | `map[string][]string` |
| Volumes by job | Find volumes attached to a job | Filter by `bosun.job.attach` |

---

## Serialization

All domain types are serializable to JSON/YAML for CLI output.

```go
// JSON tags for serialization (in actual implementation)
type Job struct {
    Name             string            `json:"name"`
    Schedule         string            `json:"schedule"`
    TargetContainers []string          `json:"targetContainers"`
    TargetStacks     []string          `json:"targetStacks"`
    WorkerImage      string            `json:"workerImage"`
    AttachVolumes    []VolumeAttachment `json:"attachVolumes"`
    SourceContainers []string          `json:"sourceContainers"`
}
```

---

## Validation Summary

| Entity | Field | Validation |
|--------|-------|------------|
| Job | Name | Non-empty string |
| Job | Schedule | Valid cron expression |
| Job | TargetContainers | No conflicting field values across containers |
| VolumeAttachment | Mode | One of: `"ro"`, `"rw"` |
| VolumeAttachment | Name | Non-empty string |
| ExecutionPlan | Steps | Non-empty, ordered |
| ContainerDependency | ServiceName | Non-empty string |
