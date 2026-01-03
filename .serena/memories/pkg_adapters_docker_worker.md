# Docker Worker Adapter

## Scope
Worker container execution in `internal/adapters/docker/worker/`.

## What
Implements `ports.WorkerRunner` for executing backup worker containers.

### Key Files
- `runner.go` - Main `Runner` struct implementing `WorkerRunner`
- `doc.go` - Package documentation

### Worker Lifecycle
1. Create container with image, env, mounts
2. Start container
3. Wait for exit (with timeout via context)
4. Capture logs
5. Remove container (unless `KeepOnFailure` and exit != 0)

### Environment Variables
**Bosun-injected** (BOSUN_* prefix):
- `BOSUN_JOB_NAME` - Job identifier from labels
- `BOSUN_RUN_ID` - Unique execution ID (UUID v4)
- `BOSUN_STACK` - Target compose stack name
- `BOSUN_DRY_RUN` - "true" or "false"

**User pass-through** via `bosun.job.worker.env.*` labels:
```yaml
labels:
  bosun.job.worker.env.RESTIC_REPOSITORY: "s3:..."
  bosun.job.worker.env.RESTIC_PASSWORD_FILE: "/run/secrets/..."
```
Prefix stripped, value passed as-is.

### Signal Protocol
On timeout:
1. Send SIGTERM to container
2. Wait 10s grace period
3. Send SIGKILL if still running
4. Exit code 137 (128+9) indicates SIGKILL

### Container Cleanup
- **Default**: Always remove after execution
- **`--keep-worker` flag**: Preserve on non-zero exit for debugging
- **Never remove**: Attached volumes (user data)

### Worker Requirements
Workers MUST:
- Have default ENTRYPOINT/CMD that performs backup
- Exit 0 on success, non-zero on failure

Workers SHOULD:
- Handle SIGTERM gracefully
- Write logs to stdout/stderr
- Complete within configured timeout

## Why
BYOI (Bring Your Own Image) approach chosen:
- Backup tools have their own conventions (RESTIC_*, PGHOST, BORG_*)
- Pass-through respects existing configurations
- No Bosun-specific image dependencies

## Related
- `pkg_ports` - Defines `WorkerRunner` interface
- `pkg_app_executor` - Orchestrates worker execution
