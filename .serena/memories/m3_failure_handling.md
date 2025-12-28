# M3 Decision: Failure Handling Strategy

**Decision Date**: 2025-12-28
**GitHub Issue**: #117
**Status**: ✅ DECIDED

## Decision Summary

| Area | Decision |
|------|----------|
| **Stop Timeout** | 30s default (configurable via label) |
| **Start Timeout** | 30s default (configurable via label) |
| **Worker Timeout** | 1h default (existing `bosun.backup.timeout`) |
| **Overall Job Timeout** | None for M3 (sum of step timeouts) |
| **Restart on Worker Failure** | Always restart stack (default) |
| **Pre-Validation** | Verify worker image exists before stopping |
| **Partial Failure** | Report and leave in current state |

---

## 1. Timeout Configuration (FR-005)

### Per-Step Timeouts

| Step | Default | Label Override | CLI Override |
|------|---------|----------------|--------------|
| Stop stack | 30s | `bosun.backup.stop-timeout` | `--stop-timeout` |
| Run worker | 1h | `bosun.backup.timeout` | `--timeout` |
| Start stack | 30s | `bosun.backup.start-timeout` | `--start-timeout` |

### Rationale for 30s Default

- Docker default is 10s, but that's for simple containers
- Database containers often need longer for clean shutdown (fsync, connections)
- 30s balances safety vs reasonable wait time
- AWS ECS uses 30s as their default
- Label override allows per-stack tuning

### Timeout Behavior

```
┌─────────────────────────────────────────────────────────┐
│ Timeout reached during stop/start                       │
│     │                                                   │
│     ▼                                                   │
│ Docker StopOptions{Timeout: &timeout}                   │
│     │                                                   │
│     ▼                                                   │
│ Docker handles SIGTERM → grace → SIGKILL internally     │
│     │                                                   │
│     ▼                                                   │
│ Return error to Bosun                                   │
└─────────────────────────────────────────────────────────┘
```

### Implementation

```go
// Default timeouts
const (
    DefaultStopTimeout  = 30 * time.Second
    DefaultStartTimeout = 30 * time.Second  // For health check wait in future
    DefaultWorkerTimeout = 1 * time.Hour
)

// StopStack with timeout
func (c *ComposeController) StopStack(ctx context.Context, stack string, timeout time.Duration) error {
    containers := c.ListStackContainers(ctx, stack)

    // Stop in reverse dependency order
    for _, container := range reversedTopologicalOrder(containers) {
        stopCtx, cancel := context.WithTimeout(ctx, timeout)
        defer cancel()

        err := c.docker.ContainerStop(stopCtx, container.ID, container.StopOptions{
            Timeout: &timeout,
        })
        if err != nil {
            return fmt.Errorf("failed to stop container %s: %w", container.Name, err)
        }
    }
    return nil
}
```

---

## 2. Failure Matrix

### Scenario 1: Stop Hangs (Container Won't Stop)

| Aspect | Behavior |
|--------|----------|
| Detection | Context timeout after `stop-timeout` |
| Docker behavior | SIGTERM → 10s grace → SIGKILL |
| Bosun action | Return error, DO NOT proceed to worker |
| Stack state | Partially stopped (some containers may be stopped) |
| User message | `Error: timed out stopping container X. Stack may be in inconsistent state.` |
| Exit code | Non-zero |

**Recovery**: User must manually fix stack state.

### Scenario 2: Stop Fails (API Error)

| Aspect | Behavior |
|--------|----------|
| Detection | Docker API returns error |
| Bosun action | Return error immediately, DO NOT proceed |
| Stack state | May be partially stopped |
| User message | `Error: failed to stop container X: <error>` |
| Exit code | Non-zero |

### Scenario 3: Worker Fails (Non-Zero Exit)

| Aspect | Behavior |
|--------|----------|
| Detection | Worker exits with code != 0 |
| Bosun action | **Still restart stack** (default) |
| Stack state | Restarted (even though worker failed) |
| User message | `Warning: worker exited with code N, but stack was restarted` |
| Exit code | Non-zero (reflects worker failure) |
| Override | `--keep-stopped` flag skips restart |

**Rationale**: Stack availability is more important than backup success. Users can retry backup later, but a stopped production stack is a bigger problem.

### Scenario 4: Start Fails (After Worker)

| Aspect | Behavior |
|--------|----------|
| Detection | Container fails to start |
| Bosun action | Log error, report to user |
| Stack state | Partially started (some containers running) |
| User message | `Error: failed to start container X: <error>. Stack may be in inconsistent state.` |
| Exit code | Non-zero |

**No retry in M3**: Automatic retry could loop forever. User must fix issue manually.

### Scenario 5: Partial Failure During Restart

| Aspect | Behavior |
|--------|----------|
| Detection | Some containers start, others fail |
| Bosun action | Continue attempting all containers |
| Stack state | Mixed - some running, some failed |
| User message | `Warning: X/Y containers started. Failed: [container names]` |
| Exit code | Non-zero |

---

## 3. Restart Policy on Worker Failure (FR-014)

### Default: Always Restart Stack

Even if the worker fails (non-zero exit), Bosun will restart the stack.

**Rationale**:
1. Stack availability > backup success
2. Production workloads shouldn't stay down because backup failed
3. User can retry backup manually
4. Logs capture worker failure for debugging

### Optional: Skip Restart

```bash
# Normal: restart stack even if worker fails
bosun job run daily-backup

# Skip restart if worker fails
bosun job run daily-backup --keep-stopped
```

**Use case**: Maintenance mode, debugging, intentional downtime.

### Future (Not M3): Label Configuration

```yaml
labels:
  bosun.backup.on-worker-failure: "restart"  # default
  # bosun.backup.on-worker-failure: "keep-stopped"
```

---

## 4. Pre-Validation Strategy (FR-023)

### Decision: Verify Image Before Stop (Fail Fast)

Before stopping the stack, Bosun will verify the worker image exists locally or can be pulled.

```
┌─────────────────────────────────────────────────────────┐
│ Job starts                                              │
│     │                                                   │
│     ▼                                                   │
│ 1. Verify worker image exists (ImageInspect or Pull)   │
│     │                                                   │
│     ├── Image missing ──► Fail immediately, stack safe │
│     │                                                   │
│     ▼                                                   │
│ 2. Stop stack                                           │
│ 3. Run worker                                           │
│ 4. Start stack                                          │
└─────────────────────────────────────────────────────────┘
```

### Implementation

```go
func (e *Executor) Run(ctx context.Context, job Job) error {
    // Step 0: Pre-validate worker image
    if err := e.validateWorkerImage(ctx, job.WorkerImage); err != nil {
        return fmt.Errorf("worker image validation failed: %w", err)
    }

    // Step 1: Stop stack (now safe to proceed)
    if err := e.composeController.StopStack(ctx, job.Stack); err != nil {
        return err
    }

    // ... rest of execution
}

func (e *Executor) validateWorkerImage(ctx context.Context, image string) error {
    // First, try to inspect locally
    _, _, err := e.docker.ImageInspectWithRaw(ctx, image)
    if err == nil {
        return nil // Image exists locally
    }

    // Image not local, try to pull
    reader, err := e.docker.ImagePull(ctx, image, image.PullOptions{})
    if err != nil {
        return fmt.Errorf("image %s not found locally or in registry: %w", image, err)
    }
    defer reader.Close()
    io.Copy(io.Discard, reader) // Consume pull output

    return nil
}
```

### Tradeoffs

| Option | Pros | Cons |
|--------|------|------|
| **Pre-validate (chosen)** | Fail fast, stack stays safe | Extra API call, slight latency |
| Optimistic | Faster happy path | Stack down if image missing |

Pre-validation wins because a missing image is a **configuration error** that should be caught before any destructive action.

---

## 5. Signal and Ctrl+C Handling

### Graceful Shutdown on SIGINT/SIGTERM

When Bosun receives SIGINT (Ctrl+C) or SIGTERM:

```
┌─────────────────────────────────────────────────────────┐
│ Signal received                                         │
│     │                                                   │
│     ▼                                                   │
│ Cancel context (propagates to all operations)           │
│     │                                                   │
│     ├── During stop: Let current container finish       │
│     │                                                   │
│     ├── During worker: Send SIGTERM to worker           │
│     │                                                   │
│     └── During start: Let current container finish      │
│                                                         │
│ Always attempt to start stack before exiting            │
└─────────────────────────────────────────────────────────┘
```

### Implementation

```go
func (e *Executor) Run(ctx context.Context, job Job) error {
    // Setup signal handling
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigCh
        log.Info("Received shutdown signal, cleaning up...")
        cancel()
    }()

    // Execute job with cleanup guarantee
    err := e.executeJob(ctx, job)

    // ALWAYS try to restart stack, even on interrupt
    if stackWasStopped {
        startCtx, startCancel := context.WithTimeout(context.Background(), DefaultStartTimeout)
        defer startCancel()

        if startErr := e.composeController.StartStack(startCtx, job.Stack); startErr != nil {
            log.Error("Failed to restart stack after interrupt: %v", startErr)
        }
    }

    return err
}
```

---

## 6. Error Types and Exit Codes

### Error Hierarchy

```go
// Port-level errors (internal/ports/compose.go)
type StopError struct {
    Container string
    Cause     error
}

type StartError struct {
    Container string
    Cause     error
}

type TimeoutError struct {
    Operation string  // "stop", "start", "worker"
    Duration  time.Duration
}

// Application-level exit codes (internal/cmd/exitcodes.go)
const (
    ExitSuccess          = 0
    ExitWorkerFailed     = 1
    ExitStopFailed       = 2
    ExitStartFailed      = 3
    ExitTimeout          = 4
    ExitImageNotFound    = 5
    ExitConfigError      = 6
)
```

### User-Facing Messages

| Scenario | Message |
|----------|---------|
| Success | `✓ Job 'daily-backup' completed successfully` |
| Worker failed | `✗ Worker exited with code 1. Stack was restarted. Check logs for details.` |
| Stop timeout | `✗ Timeout stopping container 'db'. Stack may be in inconsistent state.` |
| Image missing | `✗ Worker image 'restic:missing' not found. Stack was NOT stopped.` |
| Start failed | `✗ Failed to start container 'db': port already in use` |

---

## 7. Labels Summary

### New Labels for M3

| Label | Type | Default | Description |
|-------|------|---------|-------------|
| `bosun.backup.stop-timeout` | duration | `30s` | Timeout for stopping each container |
| `bosun.backup.start-timeout` | duration | `30s` | Timeout for starting each container |

### Existing Labels (from #110)

| Label | Type | Default | Description |
|-------|------|---------|-------------|
| `bosun.backup.timeout` | duration | `1h` | Worker execution timeout |
| `bosun.backup.worker-image` | string | required | Worker container image |
| `bosun.backup.enabled` | bool | `false` | Enable backup for this stack |

---

## 8. M3 Limitations (Documented)

1. **No automatic retry**: Failed operations are not retried
2. **No health check waiting**: Stack "started" means containers running, not healthy
3. **No rollback**: Cannot restore previous container state
4. **No concurrent job protection**: Multiple jobs on same stack may conflict
5. **No log persistence**: Logs displayed only, not stored

These limitations are acceptable for M3 MVP and documented in user-facing docs.

---

## Research Sources

### Codebase Analysis
- **Watchtower**: `waitForStopOrTimeout`, 10s default, label-configurable lifecycle timeouts
- **Portainer**: Docker client default timeouts, 1h for remote operations

### Industry Standards
- Docker default stop timeout: 10s
- Docker Compose `stop_grace_period`: 10s
- AWS ECS stop timeout: 30s
- Kubernetes termination grace period: 30s

---

## Related Issues

- #117 - This research (CLOSED)
- #114 - JobExecutor port interface
- #121 - Executor service implementation

---

*This decision defines failure handling for M3 MVP. More sophisticated retry/rollback logic deferred to future milestones.*
