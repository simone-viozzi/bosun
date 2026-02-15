# Research: Scheduling Engine & Concurrency Architecture

**Date**: 2026-02-15
**Status**: ✅ Complete (consolidated from #108)
**Related Issue**: [#108](https://github.com/simone-viozzi/bosun/issues/108) - Research: Job Concurrency Strategy

## Summary

This document consolidates research findings from #108 regarding job concurrency, scheduling architecture, and overlap handling. The research established a **three-layer concurrency model** for safe, predictable job execution.

## Three-Layer Concurrency Model

Based on #108 research conclusions, Bosun M4 implements three orthogonal concurrency controls:

### Layer 1: Per-Job Overlap Policy (per-job basis)

**Purpose**: Control what happens when a job's next scheduled time arrives while the previous instance is still running.

**Policies**:
- **`queue`**: Delay the next run until the current one completes (uses `cron.DelayIfStillRunning`)
- **`skip`**: Drop the next run if current is active (uses `cron.SkipIfStillRunning`)
- **`cancel-and-restart`**: Stop current run, start fresh [DEFERRED to #176]

**Implementation**: Provided by `robfig/cron/v3` library via job wrappers.

**Default**: `queue` (safe by default — ensures all scheduled runs eventually execute)

**Configuration**: `bosun.job.overlap-policy` label (new in M4)

### Layer 2: Global Semaphore (system-wide parallelism)

**Purpose**: Limit total number of jobs running concurrently across all stacks to control resource usage.

**Mechanism**: `golang.org/x/sync/semaphore.Weighted` with configurable N.

**Default**: N=1 (sequential execution — safe, predictable, avoids resource exhaustion)

**Configuration**: `--parallelism` flag on `bosun daemon` command

**Behavior**:
- Before executing any job: `sem.Acquire(ctx, 1)` (blocks if N jobs are running)
- After job completes: `sem.Release(1)`

**Use cases**:
- N=1: Serial execution (no concurrent jobs)
- N=3: Run up to 3 jobs simultaneously
- N=10: Higher parallelism for large infrastructures

### Layer 3: Per-Stack Mutex (automatic, always-on)

**Purpose**: Prevent concurrent jobs from executing on the same stack to avoid conflicts (e.g., two jobs trying to stop/start the same containers).

**Mechanism**: `StackLockManager` with map of stack names to `sync.Mutex` (or channels).

**Behavior**:
- Before executing job: `lockMgr.Lock(stackName)` (blocks if another job holds lock)
- After job completes: `defer lockMgr.Unlock(stackName)`

**Automatic**: Users don't configure this — it's always enforced for safety.

**Use cases**:
- Job A (targets stack `web`) and Job B (targets stack `db`) can run concurrently
- Job A (targets stack `web`) and Job C (also targets stack `web`) CANNOT run concurrently

## Execution Flow

```
Scheduler fires job at scheduled time
    ↓
Check overlap policy (Layer 1)
    - queue: enqueue and return
    - skip: check if running, if yes skip
    ↓
Acquire global semaphore (Layer 2)
    - Blocks until slot available
    ↓
Acquire per-stack lock(s) (Layer 3)
    - Sort target stacks alphabetically (deadlock prevention)
    - Lock each stack in sorted order
    - Blocks until all stacks are free
    ↓
Execute job (via Executor.Execute)
    ↓
Update failure counter
    - On success: reset consecutiveFailures to 0
    - On failure: increment; if ≥3 → circuit-break (auto-disable)
    ↓
Release per-stack lock(s) (defer, reverse order)
    ↓
Release global semaphore (defer)
```

## Cron Scheduling Library

**Choice**: `robfig/cron/v3` (already in go.mod)

**Rationale**:
- Standard Go cron library with excellent API
- Built-in overlap policy wrappers (`DelayIfStillRunning`, `SkipIfStillRunning`)
- Supports both 5-field (minute-level) and 6-field (second-level) cron expressions
- Handles goroutine lifecycle, timezone support, next-run calculation
- 13k+ GitHub stars, mature, well-tested

**API**:
```go
c := cron.New()
c.AddFunc("0 3 * * *", func() { /* job logic */ })
c.Start()
defer c.Stop()
```

**Overlap wrappers**:
```go
// Queue policy
cron.New(cron.WithChain(cron.DelayIfStillRunning(logger)))

// Skip policy
cron.New(cron.WithChain(cron.SkipIfStillRunning(logger)))
```

## Config Refresh Strategy

**Goal**: Detect new/changed/removed jobs without daemon restart.

**Mechanism**: Periodic polling loop in scheduler (default 5 minutes).

**Algorithm**:
```
func (s *Scheduler) refreshLoop(ctx context.Context) {
    ticker := time.NewTicker(s.refreshInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.refreshJobs()
        case <-ctx.Done():
            return
        }
    }
}

func (s *Scheduler) refreshJobs() {
    // 1. Re-discover jobs from Docker labels
    currentJobs := s.services.DiscoverJobs()

    // 2. Diff against registered jobs
    added := currentJobs - registeredJobs
    removed := registeredJobs - currentJobs
    changed := jobs with different schedule/policy

    // 3. Update scheduler
    for job in added {
        s.AddJob(job)
        s.events.Emit(JobAdded, job)
    }
    for job in changed {
        s.RemoveJob(job.Name)
        s.AddJob(job)
        s.events.Emit(JobChanged, job)
    }
    for job in removed {
        s.RemoveJob(job.Name)
        s.events.Emit(JobRemoved, job)
    }
}
```

**Notes**:
- Uses same discovery path as CLI commands (label source → snapshot → discover)
- Removed jobs: current run completes, but no future runs scheduled
- Changed jobs: treated as remove + add (atomic from scheduler perspective)
- Disabled jobs: not registered (or removed if previously registered)

## Graceful Shutdown

**Requirements** (from #86):
- Single signal (SIGTERM/SIGINT): graceful shutdown (wait for running jobs)
- Double signal: immediate exit (cancel running jobs)

**Implementation**:
```go
func (d *Daemon) Run(ctx context.Context) error {
    // Setup signal handling
    ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Start scheduler
    go s.scheduler.Start(ctx)

    // Wait for signal
    <-ctx.Done()

    // Graceful shutdown
    log.Info("Shutdown initiated, waiting for jobs to complete...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    if err := s.scheduler.Stop(shutdownCtx); err != nil {
        log.Error("Graceful shutdown failed, forcing exit", "error", err)
        return err
    }

    log.Info("Shutdown complete")
    return nil
}
```

**Double-signal**: Handled by cron library + context cancellation propagating to worker containers.

## Event System Design

**Purpose**: Decouple scheduler from observability concerns (logging, metrics, alerts).

**Port Interface** (`internal/ports/executor.go`):
```go
type EventEmitter interface {
    EmitJobScheduled(jobName string, nextRun time.Time)
    EmitJobStarted(jobName string, runID string)
    EmitJobCompleted(jobName string, runID string, result string)
    EmitJobFailed(jobName string, runID string, err error)
    EmitJobSkipped(jobName string, reason string)
}
```

**Adapter** (`internal/adapters/events/log_emitter.go`):
- Implements `EventEmitter` by writing structured logs
- Used by default in M4
- Can be replaced with Prometheus/Datadog adapters in M5+

**Usage in Scheduler**:
```go
s.events.EmitJobScheduled(job.Name, nextRun)
// ... execution logic ...
s.events.EmitJobCompleted(job.Name, runID, "success")
```

## Job Status Tracking

**Requirements** (for `bosun job list`):
- Current status: idle / running / completed / failed
- Last run time + result
- Next scheduled time
- Overlap policy

**Implementation**:
```go
type JobStatus struct {
    JobName      string
    Status       string // "idle" | "running" | "completed" | "failed"
    Schedule     string // cron expression
    OverlapPolicy string // "queue" | "skip"
    LastRunTime   *time.Time
    LastResult    string // "success" | "error: <msg>"
    NextRunTime   time.Time
    CurrentRunID  string // UUID if running
}

type Scheduler struct {
    statusMap sync.Map // map[string]*JobStatus, thread-safe
    // ...
}

func (s *Scheduler) UpdateStatus(jobName string, status JobStatus) {
    s.statusMap.Store(jobName, &status)
}

func (s *Scheduler) ListJobs() []JobStatus {
    var jobs []JobStatus
    s.statusMap.Range(func(key, value interface{}) bool {
        jobs = append(jobs, *value.(*JobStatus))
        return true
    })
    return jobs
}
```

**Thread safety**: Use `sync.Map` or `sync.RWMutex` to protect status updates during concurrent execution.

## App Bootstrap / Service Factory (#168)

**Problem**: CLI commands directly instantiate adapters (violation of hexagonal architecture, tracked in #141).

**Solution**: Create app-level service factory that wires dependencies.

**Structure**:
```
internal/app/app.go:

type Services struct {
    // Ports (interfaces)
    Planner   ports.JobPlanner
    Discoverer ports.JobDiscoverer
    Executor  ports.JobExecutor
    Scheduler *scheduler.Scheduler

    // Shared clients
    DockerClient *docker.Client
}

func Bootstrap(ctx context.Context, opts BootstrapOptions) (*Services, error) {
    // 1. Create Docker client
    dockerClient, err := docker.NewClientFromEnv()

    // 2. Instantiate adapters
    labelSource := dockerlabels.New(dockerClient)
    composeCtrl := compose.New(dockerClient)
    workerRunner := worker.New(dockerClient)

    // 3. Instantiate app services
    planner := planner.New(labelSource)
    executor := executor.New(planner, composeCtrl, workerRunner, dockerClient)

    // 4. Create scheduler (if daemon mode)
    if opts.DaemonMode {
        scheduler := scheduler.New(executor, opts.RefreshInterval)
        return &Services{
            Planner: planner,
            Executor: executor,
            Scheduler: scheduler,
        }, nil
    }

    return &Services{Planner: planner, Executor: executor}, nil
}
```

**Usage in CLI**:
```go
// internal/cmd/daemon.go
func runDaemon(cmd *cobra.Command, args []string) error {
    services, err := app.Bootstrap(ctx, app.BootstrapOptions{DaemonMode: true})
    if err != nil {
        return err
    }

    return services.Scheduler.Start(ctx)
}
```

**Benefits**:
- CLI layer only imports `internal/app`
- Adapters are implementation details hidden behind app layer
- Resolves #141 permanently
- Makes dependency injection explicit and testable

## Alternative Approaches Considered

### 1. Distributed Locking (Rejected)

**Approach**: Use Redis/etcd for cluster-wide job locks.

**Rejection Rationale**:
- Out of scope for M4 (single-host deployment)
- Adds operational complexity (requires external service)
- Kubernetes CronJobs are the right tool for multi-host scheduling
- Per-stack mutex + global semaphore are sufficient for single-host

### 2. DAG-based Job Dependencies (Rejected)

**Approach**: Allow jobs to depend on other jobs (job A → job B).

**Rejection Rationale**:
- Adds significant complexity (cycle detection, dependency resolution)
- Use case is better served by external orchestrators (Airflow, Temporal)
- Label-based config is not ideal for complex workflows
- Can be added in vNext if demand exists

### 3. Always-on Cancel-and-Restart (Rejected for M4)

**Approach**: Implement all 3 overlap policies in M4.

**Rejection Rationale**:
- `queue` and `skip` cover 90% of use cases
- Cancel-and-restart requires race-safe container termination logic
- Complexity outweighs benefit for M4 scope
- Deferred to #176 as P2 enhancement

### 4. File-based Config Refresh (Rejected)

**Approach**: Watch a YAML config file for changes instead of polling Docker labels.

**Rejection Rationale**:
- Bosun's architecture is label-driven (config lives in Docker labels)
- Introduces config drift (labels vs file)
- File watching has cross-platform quirks (inotify vs polling)
- Can be added as optional feature in vNext if needed (tracked in #63)

## Clarification-Derived Decisions

*Resolved during spec clarification session (2026-02-15).*

### Multi-Stack Lock Ordering (Deadlock Prevention)

**Decision**: When a job targets multiple stacks (`TargetStacks []string`), acquire locks in sorted (alphabetical) order.

**Rationale**: Two jobs targeting overlapping stack sets in different order (e.g., Job A: `[web, db]`, Job B: `[db, web]`) could deadlock with unordered acquisition. Sorted ordering enforces globally consistent lock hierarchy, making deadlocks impossible without requiring timeout-based detection.

**Implementation**:
```go
func (m *StackLockManager) LockAll(ctx context.Context, stacks []string) error {
    sorted := make([]string, len(stacks))
    copy(sorted, stacks)
    sort.Strings(sorted)

    for _, stack := range sorted {
        if err := m.Lock(ctx, stack); err != nil {
            // Release already-acquired locks on failure
            m.UnlockAll(acquired)
            return err
        }
    }
    return nil
}
```

**Alternatives Considered**: Lock-with-timeout (rejected: adds complexity, obscures root cause); lock acquisition ordering by hash (rejected: sorted is simpler and deterministic).

### Circuit-Breaker: Failure Handling Strategy

**Decision**: Continue scheduling after failure (retry on next cron tick), emit failure event, auto-disable after 3 consecutive failures.

**Rationale**: Jobs usually fail transiently (Docker daemon hiccup, network issue). Retrying on the next tick catches most transients. However, repeatedly failing jobs waste resources and may mask systemic issues, so the circuit-breaker auto-disables them after a threshold.

**Implementation**:
```go
type scheduledJobEntry struct {
    job                jobs.Job
    entryID            cron.EntryID
    consecutiveFailures int  // reset to 0 on success, incremented on failure
}

func (s *Scheduler) executeJob(ctx context.Context, entry *scheduledJobEntry) {
    result, err := s.executor.Execute(ctx, entry.job.Name, opts)
    if err != nil {
        entry.consecutiveFailures++
        s.events.EmitJobFailed(ctx, entry.job.Name, runID, err, duration)

        if entry.consecutiveFailures >= 3 {
            s.disableJob(ctx, entry.job.Name)
            s.events.EmitJobCircuitBroken(ctx, entry.job.Name, entry.consecutiveFailures)
        }
    } else {
        entry.consecutiveFailures = 0
        s.events.EmitJobCompleted(ctx, entry.job.Name, runID, duration)
    }
}
```

**Auto-disable semantics**: Circuit-broken jobs are NOT re-enabled by config refresh. Manual intervention (re-setting `bosun.job.enabled=true`) is required. This prevents flapping.

**Failure counter on restart**: Reset to 0 (in-memory, lost on restart). Daemon restart acts as a natural circuit-breaker reset. Persistent failure state is deferred with #177. The `JobStateStore` port (M4: `InMemoryStateStore`) is the extension point — #177 swaps in a durable adapter to persist `ConsecutiveFailures` across restarts.

### Job Identity Rules

**Decision**: `bosun.job.name` is the unique identity key across all containers.

**Rationale**: Jobs need a stable identity for change detection during config refresh. The name label is already required by the schema and provides a natural, human-readable key.

**Behavior**:
- Same name across different containers → same job (first-seen definition wins)
- Duplicate names detected during discovery → emit warning event, use first-seen
- Name is used for: scheduler registration, status tracking, lock tracking, event emission
- Change detection during config refresh uses name as the join key

## Open Questions

### Q1: Should global semaphore be configurable per-job instead of system-wide?

**Status**: ❌ No, system-wide is sufficient for M4.

**Reasoning**:
- Per-job quotas add complexity (need quota manager)
- System-wide limit is easier to reason about
- Can be added later if use case emerges
- Most users want simple "run N jobs at once" control

### Q2: Should we support @reboot, @daily, @hourly cron syntax?

**Status**: ✅ Yes, supported by `robfig/cron/v3` automatically.

**Reasoning**:
- No additional implementation needed
- Improves UX for common schedules
- Standard cron feature

### Q3: Should config refresh interval be configurable?

**Status**: 🔄 Hardcoded to 5min in M4, can add flag later if needed.

**Reasoning**:
- 5 minutes is reasonable default
- Avoids premature configuration complexity
- Can add `--refresh-interval` flag in vNext if requested
- Most users won't need sub-5min refresh

### Q4: Should daemon emit Prometheus metrics?

**Status**: ❌ Deferred to M5 (Observability milestone).

**Reasoning**:
- Event system (M4) provides foundation for metrics
- Prometheus adapter can be added in M5 without changing scheduler
- Log-based events are sufficient for M4 validation

## References

- **#108**: Research: Job Concurrency Strategy (primary research issue)
- **#141**: Bug: CLI commands directly import adapters
- **#28**: Feature: Event-based triggers (vNext)
- **#63**: Feature: Config hot-reload via file watch (vNext)
- **#176**: Feature: Cancel-and-restart overlap policy (deferred P2)
- **#177**: Feature: Persistent scheduling / catch-up runs (deferred, `JobStateStore` port ready)

## Research Artifacts

- [robfig/cron documentation](https://pkg.go.dev/github.com/robfig/cron/v3)
- [golang.org/x/sync/semaphore documentation](https://pkg.go.dev/golang.org/x/sync/semaphore)
- #108 comment with full concurrency architecture diagram
