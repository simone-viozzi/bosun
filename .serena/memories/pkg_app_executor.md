# Executor Service

## Scope
Job execution orchestration in `internal/app/executor/`.

## What
Implements `ports.JobExecutor` - coordinates ComposeController and WorkerRunner.

### Key Files
- `executor.go` - Main `Executor` struct implementing `JobExecutor`
- `doc.go` - Package documentation

### Execution Flow (Plan-Driven)
```
Execute(job)
    │
    ├─► 0. Generate execution plan (Planner.Plan)
    │
    ├─► 1. Pre-validate worker image (ImageInspect or Pull)
    │       └─► Fail fast if image missing (stack stays safe)
    │
    └─► 2. Execute plan steps in order:
            │
            ├─► StopContainers: Stop stack (ComposeController.StopStack)
            │       └─► Reverse dependency order
            │
            ├─► RunWorker: Run worker (WorkerRunner.Run)
            │       └─► Capture exit code and logs
            │
            └─► StartContainers: Start stack (ComposeController.StartStack)
                    └─► Always attempted, even if worker failed
```

The executor interprets the plan steps rather than hardcoding the flow, ensuring `bosun plan show` output matches actual execution.

### Timeout Configuration
| Step | Default | Label | CLI Flag |
|------|---------|-------|----------|
| Stop | 30s | `bosun.job.stop-timeout` | `--stop-timeout` |
| Worker | 1h | `bosun.job.timeout` | `--timeout` |
| Start | 30s | `bosun.job.start-timeout` | `--start-timeout` |

### Failure Handling
- **Stop fails**: Return error, do NOT proceed to worker
- **Worker fails**: Log warning, still restart stack
- **Start fails**: Return error, stack may be partial
- **SIGINT/SIGTERM**: Cancel context, always attempt stack restart

**Rationale**: Stack availability > backup success. Production workloads shouldn't stay down because backup failed.

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 10 | Worker failed (non-zero exit) |
| 11 | Stop failed |
| 12 | Start failed |
| 13 | Timeout |
| 14 | Image not found |
| 15 | Job not found |
| 16 | Interrupted |

### DryRun Mode
`DryRun(job)` validates without execution:
- Checks job exists
- Validates worker image
- Returns execution plan
- Does NOT stop stack or run worker

## Why
Centralizes orchestration logic with clear failure semantics. Pre-validation prevents stopping stack for missing images.

## Related
- `pkg_ports` - Defines `JobExecutor` interface
- `pkg_adapters_docker_compose` - Stack control
- `pkg_adapters_docker_worker` - Worker execution
- `pkg_cli` - `bosun job run` command
