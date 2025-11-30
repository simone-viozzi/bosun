# Contract: CLI Commands

**Feature Branch**: `006-backup-job-model`
**Date**: 2025-11-30
**Package**: `internal/cmd/`

## Overview

This document specifies the CLI commands for job discovery and plan visualization. Commands follow existing Bosun CLI patterns established in `config validate` and `labels snapshot`.

---

## Command: `bosun plan list`

### Synopsis

```
bosun plan list [flags]
```

List all discovered jobs from the current Docker environment.

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-o` | string | `text` | Output format: `text`, `json`, `yaml` |
| `--stopped` | | bool | `false` | Include stopped containers in discovery |
| `--stack` | `-s` | string | | Filter jobs by stack name |

### Behavior

1. Connect to Docker API
2. Take snapshot of all entities with `bosun.*` labels
3. Discover jobs from container labels
4. Apply `--stack` filter if specified
5. Output job list in requested format

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (jobs found or no jobs) |
| 1 | Validation errors in job labels |
| 2 | Runtime error (Docker unavailable) |

### Output Examples

**Text (default)**:
```
JOBS DISCOVERED: 2

NAME               SCHEDULE        CONTAINERS  VOLUMES  STACK
daily-backup       0 2 * * *       3           2        myapp
weekly-cleanup     0 0 * * 0       1           0        myapp

Run 'bosun plan show <name>' for execution details.
```

**Text (no jobs)**:
```
No jobs discovered.

To define a job, add labels to your containers:
  bosun.job.enabled=true
  bosun.job.name=my-job
  bosun.job.schedule=0 2 * * *
```

**JSON**:
```json
{
  "jobs": [
    {
      "name": "daily-backup",
      "schedule": "0 2 * * *",
      "targetContainers": ["abc123", "def456", "ghi789"],
      "targetStacks": ["myapp"],
      "workerImage": "bosun-worker:local",
      "attachVolumes": [
        {"name": "data-vol", "mode": "ro"},
        {"name": "config-vol", "mode": "ro"}
      ],
      "sourceContainers": ["abc123"]
    }
  ],
  "errors": []
}
```

**JSON (with validation errors)**:
```json
{
  "jobs": [],
  "errors": [
    {
      "entityKind": "container",
      "entityId": "abc123def456",
      "entityName": "myapp-web-1",
      "field": "bosun.job.schedule",
      "message": "invalid cron expression: unexpected character '?'"
    }
  ]
}
```

---

## Command: `bosun plan show`

### Synopsis

```
bosun plan show <job-name> [flags]
```

Show the execution plan for a specific job.

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `job-name` | Yes | Name of the job to show plan for |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-o` | string | `text` | Output format: `text`, `json`, `yaml` |
| `--stopped` | | bool | `false` | Include stopped containers in discovery |

### Behavior

1. Connect to Docker API
2. Take snapshot of all entities with `bosun.*` labels
3. Discover jobs from container labels
4. Find job matching `<job-name>`
5. Generate execution plan
6. Output plan in requested format

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Job not found or validation error |
| 2 | Runtime error (Docker unavailable) |

### Output Examples

**Text (default)**:
```
EXECUTION PLAN: daily-backup

Job:      daily-backup
Schedule: 0 2 * * * (Daily at 2:00 AM)
Stack:    myapp

STEPS:

1. STOP CONTAINERS
   Using: docker compose stop (all stack containers included)
   Project: myapp
   Containers:
     - myapp-web-1 (abc123)
     - myapp-api-1 (def456)
     - myapp-worker-1 (ghi789)

2. RUN WORKER
   Image: restic/restic:latest
   Volumes:
     - data-vol → /data/data-vol (ro)
     - config-vol → /data/config-vol (ro)
   Command: [to be configured]

Note: Container restart (step 3) is not included in this milestone.
```

**Text (job not found)**:
```
Error: job "nonexistent" not found

Available jobs:
  - daily-backup
  - weekly-cleanup

Use 'bosun plan list' to see all discovered jobs.
```

**Text (dependency validation failure)**:
```
Error: cannot generate plan for job "partial-backup"

Stopping the following containers would orphan dependents:
  - myapp-db-1 is required by: myapp-api-1, myapp-web-1

To fix, either:
  1. Add the dependent containers to the job, or
  2. Remove the dependency in your docker-compose.yml
```

**JSON**:
```json
{
  "jobName": "daily-backup",
  "steps": [
    {
      "type": "stop_containers",
      "description": "Stop 3 containers in stack 'myapp'",
      "containerIds": ["abc123", "def456", "ghi789"],
      "containerNames": ["myapp-web-1", "myapp-api-1", "myapp-worker-1"],
      "useComposeStop": true,
      "composeProject": "myapp"
    },
    {
      "type": "run_worker",
      "description": "Run worker with 2 volumes attached",
      "workerImage": "restic/restic:latest",
      "volumeMounts": [
        {"name": "data-vol", "mountPath": "/data/data-vol", "mode": "ro"},
        {"name": "config-vol", "mountPath": "/data/config-vol", "mode": "ro"}
      ]
    }
  ],
  "createdAt": "2025-11-30T12:00:00Z"
}
```

**YAML**:
```yaml
jobName: daily-backup
steps:
  - type: stop_containers
    description: "Stop 3 containers in stack 'myapp'"
    containerIds:
      - abc123
      - def456
      - ghi789
    containerNames:
      - myapp-web-1
      - myapp-api-1
      - myapp-worker-1
    useComposeStop: true
    composeProject: myapp
  - type: run_worker
    description: "Run worker with 2 volumes attached"
    workerImage: restic/restic:latest
    volumeMounts:
      - name: data-vol
        mountPath: /data/data-vol
        mode: ro
      - name: config-vol
        mountPath: /data/config-vol
        mode: ro
createdAt: "2025-11-30T12:00:00Z"
```

---

## Error Handling Patterns

### Docker Unavailable

```
Error: failed to connect to Docker

Cannot reach Docker daemon at unix:///var/run/docker.sock
Is Docker running?

Exit code: 2
```

### Invalid Job Labels (plan list)

```
VALIDATION ERRORS:

Container "myapp-web-1" (abc123):
  - bosun.job.schedule: invalid cron expression: unexpected '?'

Container "myapp-api-1" (def456):
  - bosun.job.name: required when bosun.job.enabled=true

Found 2 error(s). Fix labels and retry.

Exit code: 1
```

### Conflicting Job Configuration

```
VALIDATION ERRORS:

Job "daily-backup" has conflicting configurations:
  - bosun.job.schedule:
    - myapp-web-1: "0 2 * * *"
    - myapp-api-1: "0 3 * * *"

Multiple containers define the same job with different values.
Ensure all containers contributing to a job use identical values.

Exit code: 1
```

---

## Command Registration

```go
// internal/cmd/plan.go

func NewPlanCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "plan",
        Short: "Manage job execution plans",
        Long:  "Discover jobs and view their execution plans.",
    }

    cmd.AddCommand(NewPlanListCmd())
    cmd.AddCommand(NewPlanShowCmd())

    return cmd
}
```

```go
// internal/cmd/root.go (addition)

func NewRootCmd(ctx context.Context) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "bosun",
        Short: "Docker Compose job runner",
    }

    cmd.AddCommand(NewConfigCmd())
    cmd.AddCommand(NewLabelsCmd())
    cmd.AddCommand(NewPlanCmd())  // NEW

    return cmd
}
```

---

## Implementation Notes

1. **Reuse Snapshot Logic**: Use existing `dockerlabels.NewFromEnv()` and snapshot infrastructure
2. **Output Buffering**: Buffer output to prevent partial writes on error
3. **Color Support**: Use `isatty` detection; disable colors when piped
4. **Help Text**: Include examples in `--help` output
5. **Signal Handling**: Respect SIGINT/SIGTERM for graceful shutdown
