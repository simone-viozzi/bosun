# M3 Decision: Worker Container Contract

**Decision Date**: 2025-12-28
**GitHub Issue**: #110
**Status**: ✅ DECIDED

## Decision Summary

| Area | Decision |
|------|----------|
| **Environment Variables** | Minimal BOSUN_* metadata + label-based pass-through |
| **Signal Protocol** | SIGTERM → 10s grace → SIGKILL (Docker standard) |
| **Container Cleanup** | Always remove; `--keep-failed` flag to preserve |
| **Base Images** | BYOI only for M3; example Dockerfiles in docs |
| **Communication** | Exit codes only (no stdout parsing) |

---

## 1. Environment Variables (FR-010)

### Bosun-Injected Variables

Bosun injects a minimal set of metadata variables with `BOSUN_` prefix:

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `BOSUN_JOB_NAME` | string | Job identifier from labels | `daily-backup` |
| `BOSUN_RUN_ID` | string | Unique execution ID (UUID v4) | `550e8400-e29b-41d4-a716-446655440000` |
| `BOSUN_STACK` | string | Target compose stack name | `myapp` |
| `BOSUN_DRY_RUN` | string | Whether this is a dry run | `true` or `false` |

### User-Defined Pass-Through

Workers define their own environment variables via labels. Bosun passes these through without modification:

```yaml
services:
  db:
    labels:
      bosun.backup.enabled: "true"
      bosun.backup.worker-image: "restic/restic:latest"
      # Pass-through env vars for the worker
      bosun.backup.worker-env.RESTIC_REPOSITORY: "s3:s3.amazonaws.com/mybucket"
      bosun.backup.worker-env.RESTIC_PASSWORD_FILE: "/run/secrets/restic-password"
      bosun.backup.worker-env.AWS_ACCESS_KEY_ID: "AKIA..."
```

**Label format**: `bosun.backup.worker-env.<VAR_NAME>: "<value>"`

### Rationale

- Backup tools expect their OWN conventions (RESTIC_*, PGHOST, BORG_*, etc.)
- Bosun shouldn't duplicate or translate tool-specific variables
- Pass-through respects user's existing configuration patterns
- BOSUN_* prefix provides observability without conflicts

### Implementation Notes

```go
// WorkerConfig built from labels
type WorkerConfig struct {
    Image      string            // bosun.backup.worker-image
    Env        map[string]string // bosun.backup.worker-env.*
    Volumes    []VolumeMount     // derived from bosun.backup.volumes
    Timeout    time.Duration     // bosun.backup.timeout
}

// Injected by Bosun at runtime
func (w *WorkerRunner) buildEnv(job Job, runID string) []string {
    env := []string{
        fmt.Sprintf("BOSUN_JOB_NAME=%s", job.Name),
        fmt.Sprintf("BOSUN_RUN_ID=%s", runID),
        fmt.Sprintf("BOSUN_STACK=%s", job.Stack),
        fmt.Sprintf("BOSUN_DRY_RUN=%t", job.DryRun),
    }
    // Append user-defined pass-through vars
    for k, v := range job.WorkerEnv {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }
    return env
}
```

---

## 2. Signal Protocol (FR-011)

### Timeout Behavior

When the job timeout is reached:

```
┌─────────────────────────────────────────────────────────┐
│ Timeout reached                                         │
│     │                                                   │
│     ▼                                                   │
│ Send SIGTERM to worker PID 1                            │
│     │                                                   │
│     ▼                                                   │
│ Wait grace period (default: 10 seconds)                 │
│     │                                                   │
│     ├── Worker exits cleanly ──► Capture exit code      │
│     │                                                   │
│     └── Still running ──► Send SIGKILL                  │
│                               │                         │
│                               ▼                         │
│                          Exit code 137 (128 + 9)        │
└─────────────────────────────────────────────────────────┘
```

### Configuration

| Parameter | Default | Source |
|-----------|---------|--------|
| Job timeout | `1h` | `bosun.backup.timeout` label or `--timeout` flag |
| Grace period | `10s` | Hardcoded for M3 (future: `bosun.backup.stop-grace`) |
| Stop signal | `SIGTERM` | Docker default (future: `bosun.backup.stop-signal`) |

### Worker Expectations

Workers SHOULD:
1. Use exec form in Dockerfile (`ENTRYPOINT ["backup.sh"]`) so signals reach PID 1
2. Trap SIGTERM and perform graceful cleanup (flush buffers, close connections)
3. Exit with meaningful code before grace period expires

Workers MAY:
- Use `--init` flag if spawning child processes (Bosun can inject this)
- Ignore SIGTERM if cleanup requires longer (will be SIGKILL'd)

### Implementation Notes

```go
// Timeout handling in WorkerRunner
func (w *WorkerRunner) Run(ctx context.Context, config WorkerConfig) (int, error) {
    // Create container
    containerID, err := w.docker.ContainerCreate(ctx, &container.Config{
        Image: config.Image,
        Env:   config.Env,
    }, &container.HostConfig{
        Binds: config.Volumes,
    }, nil, nil, "")

    // Start container
    w.docker.ContainerStart(ctx, containerID, container.StartOptions{})

    // Wait with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
    defer cancel()

    statusCh, errCh := w.docker.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

    select {
    case err := <-errCh:
        // Timeout or error - stop container
        gracePeriod := 10 * time.Second
        w.docker.ContainerStop(ctx, containerID, container.StopOptions{
            Timeout: &gracePeriod,
        })
        return -1, fmt.Errorf("worker timeout: %w", err)
    case status := <-statusCh:
        return int(status.StatusCode), nil
    }
}
```

---

## 3. Container Cleanup (FR-012)

### Default Behavior

Workers are removed after execution regardless of exit code:

```go
defer w.docker.ContainerRemove(ctx, containerID, container.RemoveOptions{
    Force: true,
    RemoveVolumes: false, // NEVER remove volumes - user data!
})
```

### Debug Mode

`--keep-failed` flag preserves containers with non-zero exit codes:

```bash
# Normal: container removed after run
bosun job run daily-backup

# Debug: keep container if it fails
bosun job run daily-backup --keep-failed
```

When kept, container name includes run ID for identification:
`bosun-worker-daily-backup-550e8400`

### Cleanup Scope

| What | Removed? |
|------|----------|
| Worker container | ✅ Yes (unless `--keep-failed` and failed) |
| Attached volumes | ❌ Never |
| Worker image | ❌ Never |
| Worker logs | ✅ Yes (captured to stdout before removal) |

### Rationale

- Clean default prevents container accumulation
- Debug option available when needed
- Volumes are user data - never auto-remove
- Logs captured before container removal

---

## 4. Base Images

### M3 Decision: BYOI (Bring Your Own Image)

Bosun does NOT provide base images for M3. Users specify any image via labels:

```yaml
labels:
  bosun.backup.worker-image: "restic/restic:0.16.0"
  # or
  bosun.backup.worker-image: "postgres:15-alpine"
  # or
  bosun.backup.worker-image: "myregistry/custom-backup:latest"
```

### Image Requirements

Workers MUST:
- Have a default `ENTRYPOINT` or `CMD` that performs the backup
- Exit with code 0 on success, non-zero on failure

Workers SHOULD:
- Handle SIGTERM gracefully
- Write logs to stdout/stderr (captured by Bosun)
- Complete within the configured timeout

### Documentation Deliverables

Provide example Dockerfiles/compose snippets for common patterns:

1. **PostgreSQL pg_dump**
2. **Restic backup**
3. **MySQL/MariaDB dump**
4. **Generic tar/rsync**

### Future (M6+)

Consider "blessed" example images published to Docker Hub:
- `bosun/worker-postgres:1.0`
- `bosun/worker-restic:1.0`

These would add convenience features (progress reporting, retry logic) but remain optional.

---

## 5. Worker Communication

### M3: Exit Codes Only

| Exit Code | Meaning | Bosun Interpretation |
|-----------|---------|---------------------|
| `0` | Success | Job succeeded |
| `1-125` | Application error | Job failed |
| `126` | Command not executable | Job failed (config error) |
| `127` | Command not found | Job failed (config error) |
| `137` (128+9) | SIGKILL | Job failed (timeout) |
| `143` (128+15) | SIGTERM | Job failed (interrupted) |

### Logs

- stdout/stderr captured via Docker API (`ContainerLogs`)
- Displayed to user in CLI (unless `--quiet`)
- Not persisted to disk in M3 (future: log storage)

### Future Considerations (Not M3)

**Stdout markers** for progress reporting:
```
BOSUN:PROGRESS:50
BOSUN:STATUS:Uploading to S3
BOSUN:METRIC:bytes_uploaded:1048576
```

**HTTP callback** for real-time updates:
```
POST http://bosun:8080/api/runs/{run_id}/progress
```

These add complexity and are deferred to future milestones.

---

## Port Interface Implications

### WorkerRunner Port (#116)

```go
// Package ports

// WorkerRunner executes backup worker containers
type WorkerRunner interface {
    // Run executes a worker container and returns exit code
    // Blocks until container exits or timeout
    Run(ctx context.Context, config WorkerConfig) (exitCode int, logs string, err error)
}

// WorkerConfig defines worker container configuration
type WorkerConfig struct {
    Image       string            // Container image
    Env         map[string]string // Environment variables (BOSUN_* + user pass-through)
    Mounts      []VolumeMount     // Volume mounts
    Timeout     time.Duration     // Execution timeout
    KeepFailed  bool              // Keep container on failure
}

// VolumeMount defines a volume attachment
type VolumeMount struct {
    Source   string // Volume name or host path
    Target   string // Container mount path
    ReadOnly bool   // Mount as read-only
}
```

---

## Research Sources

### Web Search Findings
- Restic, Borg, pg_dump all expect their OWN env var conventions
- Docker standard: SIGTERM → grace → SIGKILL
- K8s pattern: JOB_ID, RUN_ID for job metadata
- No universal standard - pass-through is pragmatic

### Codebase Research
- **Watchtower**: SIGTERM + configurable signal, timeout waiting, container removal after stop
- **Portainer**: Compose v2 library, status polling, context-based timeouts

---

## Related Issues

- #110 - This research (CLOSED)
- #116 - WorkerRunner port interface
- #119 - WorkerRunner adapter implementation

---

*This contract defines the minimum viable worker interface for M3. Extensions (progress reporting, base images, log persistence) are deferred to future milestones.*
