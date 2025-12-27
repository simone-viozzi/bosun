# Jobs Domain Package

## Scope
Domain types for job modeling and execution planning in `internal/domain/jobs/`.

## What

### Core Types

**`Job`** - Discovered from Docker labels
- `Name string` - Unique identifier (from `bosun.job.name`)
- `Schedule string` - Cron expression for scheduling
- `TargetContainers []string` - Container IDs to stop during execution
- `TargetStacks []string` - Unique stack names (for display)
- `WorkerImage string` - Container image for the worker
- `AttachVolumes []VolumeAttachment` - Volumes to mount in worker
- `SourceContainers []string` - Containers that defined this job

**`VolumeAttachment`** - Volume mount specification
- `Name string` - Docker volume name
- `MountPath string` - Path in worker (default: `/data/<volume-name>`)
- `Mode string` - `ro` or `rw` (default: `ro`)

**`ExecutionPlan`** - Ordered steps to execute a job
- `JobName string` - Reference to the job
- `Steps []PlanStep` - Ordered actions
- `CreatedAt time.Time` - Generation timestamp

**`PlanStep`** - Single execution action
- `Type StepType` - `stop_containers`, `run_worker`, `start_containers`
- `Description string` - Human-readable explanation
- `ContainerIDs/Names []string` - Affected containers
- `WorkerImage string` - Image to run (for run_worker)
- `VolumeMounts []VolumeAttachment` - Volumes to attach
- `UseComposeStop bool` - Use `docker compose stop`
- `ComposeProject string` - Project name for compose

### Defaults
Defaults are defined as functions in `internal/config/schema/job_labels.go`:
- `schema.DefaultJobSchedule()` - Returns default cron expression
- `schema.DefaultJobWorkerImage()` - Returns default worker image
- `schema.DefaultJobMountMode()` - Returns default mount mode ("ro")

### Additional Types

**`Stack`** - Stack-level grouping
- `Name string` - Stack identifier
- `Containers []string` - Container IDs
- `Volumes []string` - Volume names
- `Networks []string` - Network names

**`ContainerDependency`** - Compose-style dependency
- `ServiceName string` - Dependency service name
- `Condition string` - Start condition (service_started, service_healthy)
- `Required bool` - Whether dependency is required

## Why
Domain types are vendor-agnostic and serializable (JSON tags). Plans are deterministic for testing.

## Related
- `pkg_adapters_joblabels` - Discovers Jobs from labels
- `pkg_app_planner` - Generates ExecutionPlans
- `arch_overview` - Hexagonal architecture context
