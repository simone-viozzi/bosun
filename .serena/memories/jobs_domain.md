# Jobs Domain

**Location**: `internal/domain/jobs/`

## Overview

Core business entities for backup job modeling and execution planning. Contains domain types that represent jobs discovered from Docker labels and the execution plans generated from them.

## Constants

- `DefaultSchedule = "0 0 * * *"` - Daily at midnight
- `DefaultWorkerImage = "bosun-worker:local"` - Placeholder worker image
- `DefaultMountMode = "ro"` - Read-only mount mode

## Key Types

### `Job`
Represents a backup job discovered from Docker labels:
- `Name string` - Unique identifier (from `bosun.job.name`)
- `Schedule string` - Validated cron expression
- `TargetContainers []string` - Container IDs to stop during backup
- `TargetStacks []string` - Unique stack names (for display)
- `WorkerImage string` - Container image for executing the job
- `AttachVolumes []VolumeAttachment` - Volumes to mount in worker
- `SourceContainers []string` - Containers that contributed to this job

### `VolumeAttachment`
Describes a volume mount for the worker container:
- `Name string` - Docker volume name
- `MountPath string` - Path in worker container (default: `/data/<volume-name>`)
- `Mode string` - Mount mode: `ro` or `rw` (default: `ro`)

### `ExecutionPlan`
Ordered sequence of steps to execute a job:
- `JobName string` - Reference to the job
- `Steps []PlanStep` - Ordered actions
- `CreatedAt time.Time` - Plan generation timestamp

### `PlanStep`
Single action in an execution plan:
- `Type StepType` - Step type enum
- `Description string` - Human-readable explanation
- `ContainerIDs []string` - Containers affected (for stop/start)
- `ContainerNames []string` - Human-readable names
- `WorkerImage string` - Image to run (for run_worker)
- `VolumeMounts []VolumeAttachment` - Volumes to attach
- `UseComposeStop bool` - Whether to use `docker compose stop`
- `ComposeProject string` - Project name for compose commands

### `StepType` (enum)
- `StepTypeStopContainers = "stop_containers"`
- `StepTypeRunWorker = "run_worker"`
- `StepTypeStartContainers = "start_containers"`

### `Stack`
Utility type for stack-level operations:
- `Project string` - Compose project name
- `Containers []ContainerDependency` - Containers in the stack

### `ContainerDependency`
Container within a stack:
- `ID string` - Container ID
- `Name string` - Container name

## Design Notes

- All types use `json` struct tags for serialization
- JSON output is suitable for machine-readable CLI formats
- Plans are deterministic given the same input
- `CreatedAt` is set by the caller (not the planner) for determinism in tests

## Usage

Jobs are discovered by `joblabels.Discoverer` from Docker labels and passed to `planner.Planner` to generate execution plans.
