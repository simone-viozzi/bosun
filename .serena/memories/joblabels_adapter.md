# Job Labels Adapter

**Location**: `internal/adapters/joblabels/`

## Overview

Adapter that discovers backup jobs from Docker labels. Implements `ports.JobDiscoverer` interface. Transforms labeled entities from `labels.Snapshot` into domain `jobs.Job` objects.

## Label Constants

Container labels:
- `LabelJobEnabled = "bosun.job.enabled"` - Enables job participation (boolean)
- `LabelJobName = "bosun.job.name"` - Unique job identifier
- `LabelJobSchedule = "bosun.job.schedule"` - Cron expression
- `LabelJobWorkerImage = "bosun.job.worker.image"` - Worker container image

Volume labels:
- `LabelJobAttach = "bosun.job.attach"` - Job name to attach volume to
- `LabelJobMountPath = "bosun.job.mount.path"` - Mount path in worker
- `LabelJobMountMode = "bosun.job.mount.mode"` - Mount mode (ro|rw)

Stack resolution:
- `LabelStack = "bosun.stack"` - Explicit stack name (highest priority)
- `LabelComposeProject = "com.docker.compose.project"` - Docker Compose project name

## Key Types

### `Discoverer`
Implements `ports.JobDiscoverer`:
- `NewDiscoverer() *Discoverer` - Constructor with cron parser
- `DiscoverJobs(ctx, snapshot) ([]Job, []ValidationError, error)` - Main discovery method

## Discovery Algorithm

Three-phase process:

### Phase 1: Collect Job Definitions
- Scan containers with `bosun.job.enabled=true`
- Validate job name is present
- Build job map by name
- Merge fields from multiple containers (detect conflicts)
- Validate cron expressions
- Resolve stack names

### Phase 2: Attach Volumes
- Scan volumes with `bosun.job.attach=<job-name>`
- Validate referenced job exists
- Apply default mount path (`/mnt/<volume-name>`)
- Validate mount mode (ro|rw, default: ro)

### Phase 3: Build Jobs
- Apply defaults for missing fields
- Construct final `jobs.Job` objects

## Validation Errors

Non-fatal errors are collected and returned alongside valid jobs:
- Missing job name when enabled
- Invalid cron expression
- Conflicting schedule between containers
- Conflicting worker image
- Volume referencing non-existent job
- Invalid mount mode

## Stack Name Resolution

Priority order:
1. `bosun.stack` label (explicit override)
2. `com.docker.compose.project` (from Docker Compose)
3. Empty string (no stack)

## Cron Validation

Uses `robfig/cron/v3` parser with:
- 5-field format (minute hour day-of-month month day-of-week)
- Standard cron syntax

## Testing

Unit tests cover:
- Single container job discovery
- Multiple containers merged into one job
- Volume attachment
- Validation error scenarios
- Stack name resolution
- Cron validation

Integration tests in `integration/joblabels_test.go` verify end-to-end discovery.
