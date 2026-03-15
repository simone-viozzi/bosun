# Data Model: Scheduling Engine & Runtime

**Date**: 2026-02-15
**Status**: Draft
**Related**: [spec.md](spec.md), [research.md](research.md)

## Overview

This document defines the key domain types, port interfaces, and data structures for M4 scheduling engine. Follows hexagonal architecture with clear separation between domain, ports, and adapters.

## Domain Types

### Domain: OverlapPolicy (NEW)

**Location**: `internal/domain/jobs/types.go`

```go
// OverlapPolicy defines behavior when job is scheduled while previous run is active
type OverlapPolicy string

const (
    // OverlapPolicyQueue delays next run until current completes
    OverlapPolicyQueue OverlapPolicy = "queue"

    // OverlapPolicySkip drops next run if current is active
    OverlapPolicySkip OverlapPolicy = "skip"

    // OverlapPolicyCancelRestart stops current run and starts fresh
    // DEFERRED: Not implemented in M4, tracked in #176
    OverlapPolicyCancelRestart OverlapPolicy = "cancel-and-restart"
)

// DefaultOverlapPolicy is used when not specified
const DefaultOverlapPolicy = OverlapPolicyQueue

// ValidateOverlapPolicy returns error if policy is invalid
func ValidateOverlapPolicy(policy OverlapPolicy) error {
    switch policy {
    case OverlapPolicyQueue, OverlapPolicySkip:
        return nil
    case OverlapPolicyCancelRestart:
        return fmt.Errorf("cancel-and-restart policy not implemented (tracked in #176)")
    default:
        return fmt.Errorf("invalid overlap policy: %s (must be 'queue' or 'skip')", policy)
    }
}
```

### Domain: Job (EXTENDED)

**Location**: `internal/domain/jobs/types.go`

```go
// Job represents a scheduled backup/maintenance job (from M3)
type Job struct {
    Name             string
    Schedule         string           // NEW: cron expression (e.g., "0 3 * * *")
    OverlapPolicy    OverlapPolicy    // NEW: how to handle overlapping runs
    Enabled          bool             // NEW: whether job is active

    // Existing fields from M3
    TargetContainers []ContainerFilter
    TargetStacks     []string
    WorkerImage      string
    AttachVolumes    []VolumeMount
    SourceContainers []string
}
```

**New fields**:
- `Schedule`: Cron expression for automatic execution
- `OverlapPolicy`: Controls concurrent execution behavior
- `Enabled`: Allows pausing jobs without removing configuration

**Identity rule**: `bosun.job.name` is the unique identity key across all containers. Two jobs with the same name are considered the same job — duplicate names detected during discovery emit a warning event and use the first-seen definition.

### Domain: JobStatus (NEW)

**Location**: `internal/domain/jobs/status.go`

```go
// JobStatus represents current state of a scheduled job
type JobStatus struct {
    JobName       string
    Status        RunStatus     // idle / running / completed / failed
    Schedule      string        // cron expression
    OverlapPolicy OverlapPolicy
    LastRunTime   *time.Time    // nil if never run
    LastResult    string        // "success" or "error: <message>"
    NextRunTime   time.Time     // calculated by cron library
    CurrentRunID  string        // UUID if running, empty otherwise
}

// RunStatus represents execution state
type RunStatus string

const (
    RunStatusIdle      RunStatus = "idle"      // not currently executing
    RunStatusRunning   RunStatus = "running"   // currently executing
    RunStatusCompleted RunStatus = "completed" // last run succeeded
    RunStatusFailed    RunStatus = "failed"    // last run failed
)
```

**Usage**:
- Tracked by Scheduler in memory
- Returned by `Scheduler.ListJobs()` for `bosun job list` command
- Updated during job lifecycle

## Port Interfaces

### Port: EventEmitter (NEW)

**Location**: `internal/ports/executor.go`

```go
// EventEmitter emits job lifecycle events for observability
type EventEmitter interface {
    // EmitJobScheduled is called when job is registered with scheduler
    EmitJobScheduled(ctx context.Context, jobName string, schedule string, nextRun time.Time)

    // EmitJobStarted is called when job execution begins
    EmitJobStarted(ctx context.Context, jobName string, runID string)

    // EmitJobCompleted is called when job finishes successfully
    EmitJobCompleted(ctx context.Context, jobName string, runID string, duration time.Duration)

    // EmitJobFailed is called when job execution fails
    EmitJobFailed(ctx context.Context, jobName string, runID string, err error, duration time.Duration)

    // EmitJobSkipped is called when scheduled run is skipped (overlap policy)
    EmitJobSkipped(ctx context.Context, jobName string, reason string)

    // EmitJobAdded is called when config refresh discovers new job
    EmitJobAdded(ctx context.Context, jobName string)

    // EmitJobRemoved is called when config refresh detects removed job
    EmitJobRemoved(ctx context.Context, jobName string)

    // EmitJobChanged is called when job schedule/policy changes
    EmitJobChanged(ctx context.Context, jobName string, oldSchedule, newSchedule string)

    // EmitJobCircuitBroken is called when a job is auto-disabled after consecutive failures
    EmitJobCircuitBroken(ctx context.Context, jobName string, consecutiveFailures int)

    // EmitJobDuplicateName is called when duplicate job names detected during discovery
    EmitJobDuplicateName(ctx context.Context, jobName string, containerID string)
}
```

**Implementations**:
- `LogEmitter` (M4): Writes structured logs
- `PrometheusEmitter` (M5+): Emits metrics
- `DatadogEmitter` (vNext): Sends traces

### Port: JobStateStore (NEW — persistence-ready)

**Location**: `internal/ports/state.go`

```go
// JobStateStore persists per-job state across daemon restarts.
// M4 ships with InMemoryStateStore (no durability).
// #177 adds a durable adapter (BoltDB / SQLite) as a drop-in replacement.
type JobStateStore interface {
    // SaveJobState persists state for a single job.
    // Called after every execution completes (success or failure).
    SaveJobState(ctx context.Context, state JobState) error

    // LoadJobState retrieves persisted state for a single job.
    // Returns ErrJobStateNotFound if no state exists.
    LoadJobState(ctx context.Context, jobName string) (JobState, error)

    // ListJobStates returns all persisted job states.
    // Used during startup for catch-up reconciliation (#177).
    ListJobStates(ctx context.Context) ([]JobState, error)

    // DeleteJobState removes persisted state for a job.
    // Called when a job is permanently removed during config refresh.
    DeleteJobState(ctx context.Context, jobName string) error
}

// JobState is the subset of per-job state that survives daemon restarts.
type JobState struct {
    JobName             string
    LastRunAt           *time.Time // nil if never run
    LastResult          string     // "success" or "error: <message>"
    ConsecutiveFailures int        // circuit-breaker counter
}
```

**Implementations**:
- `InMemoryStateStore` (M4): Uses `sync.Map`, zero durability — state lost on restart
- `BoltStateStore` (#177): BoltDB-based, file-local durability for catch-up runs

**Design rationale**: Extracting `consecutiveFailures` and `LastRunAt` into a port interface means #177 (persistent scheduling) becomes a drop-in adapter swap — no scheduler refactoring required.

### Port: JobExecutor (EXISTING from M3)

**Location**: `internal/ports/executor.go`

```go
// JobExecutor executes a job's backup/maintenance workflow
type JobExecutor interface {
    // Execute runs a job: discover → validate → stop → backup → start
    Execute(ctx context.Context, jobName string, opts ExecuteOptions) (*jobs.ExecutionResult, error)
}

type ExecuteOptions struct {
    DryRun         bool
    Timeout        time.Duration
    KeepOnFailure  bool
}
```

**Usage in M4**:
- Called by Scheduler when job is triggered by cron
- No changes needed from M3 implementation
- Concurrency controls (stack lock + semaphore) wrap calls to `Execute`

## App Layer Types

### App: Scheduler (NEW)

**Location**: `internal/app/scheduler/scheduler.go`

```go
// Scheduler manages cron-based job scheduling with config refresh
type Scheduler struct {
    cron      *cron.Cron           // robfig/cron scheduler
    executor  ports.JobExecutor    // for executing jobs
    events    ports.EventEmitter   // for lifecycle events
    discoverer ports.JobDiscoverer // for config refresh
    stateStore ports.JobStateStore // for persisting per-job state (M4: in-memory)

    stackLocks  *StackLockManager  // per-stack mutex
    globalSem   *semaphore.Weighted // global concurrency limit

    statusMap   sync.Map           // map[string]*JobStatus
    entries     sync.Map           // map[string]*scheduledJobEntry for tracking

    refreshInterval time.Duration
    cancel          context.CancelFunc
    wg              sync.WaitGroup
}

// scheduledJobEntry tracks per-job state within the scheduler
type scheduledJobEntry struct {
    job                 jobs.Job
    entryID             cron.EntryID  // cron library entry ID for removal
    consecutiveFailures int           // reset to 0 on success; auto-disable at 3 (circuit-breaker)
}

// SchedulerOptions configures scheduler behavior
type SchedulerOptions struct {
    RefreshInterval time.Duration // how often to refresh config (default 5min)
    Parallelism     int          // global concurrency limit (default 1)
}

// New creates a new Scheduler with dependencies
func New(
    executor ports.JobExecutor,
    discoverer ports.JobDiscoverer,
    events ports.EventEmitter,
    stateStore ports.JobStateStore,
    opts SchedulerOptions,
) *Scheduler

// Start begins scheduling and config refresh loop
func (s *Scheduler) Start(ctx context.Context) error

// Stop gracefully shuts down scheduler (waits for running jobs)
func (s *Scheduler) Stop(ctx context.Context) error

// AddJob registers a job with cron scheduler
func (s *Scheduler) AddJob(ctx context.Context, job jobs.Job) error

// RemoveJob unregisters a job from scheduler
func (s *Scheduler) RemoveJob(ctx context.Context, jobName string) error

// ListJobs returns current status of all scheduled jobs
func (s *Scheduler) ListJobs() []jobs.JobStatus
```

### App: StackLockManager (NEW)

**Location**: `internal/app/concurrency/stack_lock.go`

```go
// StackLockManager provides per-stack mutex for preventing concurrent
// execution on the same Compose stack
type StackLockManager struct {
    locks sync.Map // map[string]*sync.Mutex
}

// NewStackLockManager creates a new lock manager
func NewStackLockManager() *StackLockManager

// Lock acquires exclusive lock for a single stack
// Blocks until lock is available or context is cancelled
func (m *StackLockManager) Lock(ctx context.Context, stackName string) error

// LockAll acquires locks for multiple stacks in sorted (alphabetical) order
// Prevents deadlocks when two jobs target overlapping stack sets
// On failure, releases any already-acquired locks
func (m *StackLockManager) LockAll(ctx context.Context, stacks []string) error

// Unlock releases lock for a single stack
// Safe to call multiple times (idempotent)
func (m *StackLockManager) Unlock(stackName string)

// UnlockAll releases locks for multiple stacks
func (m *StackLockManager) UnlockAll(stacks []string)

// IsLocked returns whether stack is currently locked
func (m *StackLockManager) IsLocked(stackName string) bool
```

**Implementation note**: Uses `sync.Map` to avoid lock contention on map access.

### App: Services (EXTENDED)

**Location**: `internal/app/app.go`

```go
// Services provides dependency-injected application services
type Services struct {
    // Existing from M3
    Planner    ports.JobPlanner
    Discoverer ports.JobDiscoverer
    Executor   ports.JobExecutor

    // NEW in M4
    Scheduler     *scheduler.Scheduler
    EventEmitter  ports.EventEmitter
    StateStore    ports.JobStateStore
    StackLocks    *concurrency.StackLockManager
    GlobalSem     *semaphore.Weighted

    // Shared infrastructure
    DockerClient *docker.Client
}

// BootstrapOptions configures service initialization
type BootstrapOptions struct {
    DaemonMode      bool          // whether to initialize scheduler
    Parallelism     int           // global concurrency limit (default 1)
    RefreshInterval time.Duration // config refresh interval (default 5min)
}

// Bootstrap creates and wires all application services
func Bootstrap(ctx context.Context, opts BootstrapOptions) (*Services, error)
```

**Purpose**: Resolves #141 by centralizing adapter instantiation.

## Configuration Schema Extensions

### Config: Job Labels (EXTENDED)

**Location**: `internal/config/schema/job_labels.go`

```go
// JobLabelsV1 defines labels for job configuration
type JobLabelsV1 struct {
    // Existing from M3
    Name    string   `bosun:"key=bosun.job.name,scope=container,type=string,required=true,doc='Job name'"`
    Enabled *bool    `bosun:"key=bosun.job.enabled,scope=container,type=bool,default=true,doc='Whether job is enabled'"`

    // NEW in M4
    Schedule      string        `bosun:"key=bosun.job.schedule,scope=container,type=string,required=true,doc='Cron schedule (e.g., 0 3 * * *)'"`
    OverlapPolicy OverlapPolicy `bosun:"key=bosun.job.overlap-policy,scope=container,type=string,default=queue,enum=queue|skip,doc='Behavior when job overlaps (queue or skip)'"`

    // Existing execution config
    WorkerImage      string   `bosun:"key=bosun.job.worker-image,scope=container,type=string,required=true,doc='Docker image for worker'"`
    TargetStacks     []string `bosun:"key=bosun.job.target-stacks,scope=container,type=[]string,doc='Compose stacks to target'"`
    AttachVolumes    []string `bosun:"key=bosun.job.attach-volumes,scope=container,type=[]string,doc='Volumes to attach'"`
    SourceContainers []string `bosun:"key=bosun.job.source-containers,scope=container,type=[]string,doc='Containers to source volumes from'"`
}
```

**New labels**:
- `bosun.job.schedule`: Required for scheduled execution
- `bosun.job.overlap-policy`: Optional, defaults to `queue`
- `bosun.job.enabled`: Optional, defaults to `true` (moved from vNext to M4)

## Adapter Implementations

### Adapter: InMemoryStateStore (NEW)

**Location**: `internal/adapters/state/memory.go`

```go
// InMemoryStateStore implements JobStateStore with no durability.
// State is lost on daemon restart. This is the M4 default.
// #177 replaces this with a durable adapter (BoltDB).
type InMemoryStateStore struct {
    states sync.Map // map[string]*ports.JobState
}

func NewInMemoryStateStore() *InMemoryStateStore

func (s *InMemoryStateStore) SaveJobState(ctx context.Context, state ports.JobState) error {
    s.states.Store(state.JobName, &state)
    return nil
}

func (s *InMemoryStateStore) LoadJobState(ctx context.Context, jobName string) (ports.JobState, error) {
    val, ok := s.states.Load(jobName)
    if !ok {
        return ports.JobState{}, ports.ErrJobStateNotFound
    }
    return *val.(*ports.JobState), nil
}
```

**Purpose**: Default M4 implementation. Same `sync.Map` approach already planned, but behind a port interface so #177 swaps in a durable adapter without touching the scheduler.

### Adapter: LogEmitter (NEW)

**Location**: `internal/adapters/events/log_emitter.go`

```go
// LogEmitter implements EventEmitter by writing structured logs
type LogEmitter struct {
    logger *slog.Logger
}

// NewLogEmitter creates a new log-based event emitter
func NewLogEmitter(logger *slog.Logger) *LogEmitter

// Implements all EventEmitter methods by calling logger.Info() with structured attrs
func (e *LogEmitter) EmitJobScheduled(ctx context.Context, jobName string, schedule string, nextRun time.Time) {
    e.logger.InfoContext(ctx, "Job scheduled",
        "event", "job.scheduled",
        "job_name", jobName,
        "schedule", schedule,
        "next_run", nextRun,
    )
}
```

**Purpose**: Default M4 implementation, provides observability via logs.

## CLI Command Types

### CLI: DaemonOptions (NEW)

**Location**: `internal/cmd/daemon.go`

```go
// DaemonOptions configures daemon command behavior
type DaemonOptions struct {
    Parallelism     int           // --parallelism flag
    RefreshInterval time.Duration // --refresh-interval flag (optional, default 5min)
}
```

### CLI: JobListOptions (NEW)

**Location**: `internal/cmd/job_list.go`

```go
// JobListOptions configures job list command
type JobListOptions struct {
    Format string // --format flag: text (default) | json | yaml
    Watch  bool   // --watch flag: continuously update (optional for M4)
}
```

## Data Flow Diagrams

### Config Refresh Flow

```
Timer fires (every 5 min)
    ↓
Discoverer.DiscoverJobs()
    ↓
    ├─ LabelSource.GetLabels()
    ├─ Snapshot creation
    └─ Job parsing
    ↓
Diff: currentJobs vs registeredJobs
    ↓
    ├─ New jobs → Scheduler.AddJob()
    │   ├─ Parse cron schedule
    │   ├─ Wrap with overlap policy
    │   ├─ Register with cron.Cron
    │   └─ EventEmitter.EmitJobAdded()
    │
    ├─ Removed jobs → Scheduler.RemoveJob()
    │   ├─ cron.Cron.Remove(entryID)
    │   └─ EventEmitter.EmitJobRemoved()
    │
    └─ Changed jobs → Remove + Add
        └─ EventEmitter.EmitJobChanged()
```

### Job Execution Flow

```
Cron triggers job at scheduled time
    ↓
Overlap policy check (cron library)
    ├─ queue: enqueue and proceed
    └─ skip: check if running, if yes return
    ↓
Acquire global semaphore
    globalSem.Acquire(ctx, 1) [blocks if limit reached]
    ↓
Acquire per-stack lock(s)
    stackLocks.LockAll(ctx, sortedStacks) [sorted alphabetical, blocks if any stack busy]
    ↓
Update status: running
    statusMap.Store(jobName, &JobStatus{Status: "running", ...})
    ↓
EventEmitter.EmitJobStarted()
    ↓
Executor.Execute(ctx, jobName, opts)
    ↓
    ├─ Success → consecutiveFailures = 0
    │            EventEmitter.EmitJobCompleted()
    └─ Failure → consecutiveFailures++
                 EventEmitter.EmitJobFailed()
                 if consecutiveFailures >= 3 → EventEmitter.EmitJobCircuitBroken()
                                                auto-disable job
    ↓
Persist job state
    stateStore.SaveJobState(ctx, JobState{LastRunAt, LastResult, ConsecutiveFailures})
    ↓
Update status: completed / failed
    statusMap.Store(jobName, &JobStatus{Status: "completed", ...})
    ↓
Release per-stack lock(s) (defer)
    stackLocks.UnlockAll(sortedStacks)
    ↓
Release global semaphore (defer)
    globalSem.Release(1)
```

### Job List Flow

```
User runs: bosun job list
    ↓
CLI connects to daemon's Services
    ↓
Scheduler.ListJobs()
    ↓
    ├─ Read statusMap (sync.Map.Range)
    └─ Format JobStatus entries
    ↓
Render output (text / json / yaml)
    ↓
Print to stdout
```

## Testing Strategy

### Unit Tests

**Scheduler**:
- `TestScheduler_AddJob`: Verify job is registered with cron
- `TestScheduler_RemoveJob`: Verify job is unregistered
- `TestScheduler_ConfigRefresh`: Mock discoverer, verify diff logic
- `TestScheduler_OverlapPolicies`: Verify queue/skip behavior with mock executor

**StackLockManager**:
- `TestStackLock_ConcurrentAccess`: Verify mutual exclusion per stack
- `TestStackLock_DifferentStacks`: Verify parallel execution on different stacks
- `TestStackLock_ContextCancellation`: Verify lock respects context

**InMemoryStateStore**:
- `TestInMemoryStateStore_SaveLoad`: Round-trip save and load
- `TestInMemoryStateStore_NotFound`: Verify ErrJobStateNotFound
- `TestInMemoryStateStore_ListAll`: Verify ListJobStates returns all entries
- `TestInMemoryStateStore_Delete`: Verify DeleteJobState removes entry

**EventEmitter**:
- `TestLogEmitter_Events`: Verify all event methods write logs with correct attrs

### Integration Tests

**Location**: `integration/scheduling_test.go`

```go
// TestScheduler_RealCron tests scheduling with actual cron library
func TestScheduler_RealCron(t *testing.T)

// TestScheduler_ConfigRefresh_Docker tests refresh with real Docker
func TestScheduler_ConfigRefresh_Docker(t *testing.T)

// TestConcurrency_StackLocks tests per-stack mutex with real executor
func TestConcurrency_StackLocks(t *testing.T)

// TestConcurrency_GlobalSemaphore tests parallelism limit
func TestConcurrency_GlobalSemaphore(t *testing.T)

// TestDaemon_GracefulShutdown tests signal handling
func TestDaemon_GracefulShutdown(t *testing.T)
```

**Requirements**:
- Real Docker daemon
- `testcontainers-go` for Compose fixture
- Race detector enabled
- Accelerated timers (short cron intervals like `*/2 * * * * *`)

## File Organization

```
internal/
├── domain/jobs/
│   ├── types.go         [EXTENDED] Job with Schedule, OverlapPolicy, Enabled
│   ├── status.go        [NEW] JobStatus, RunStatus
│   └── overlap.go       [NEW] OverlapPolicy type + validation
│
├── ports/
│   ├── executor.go      [EXTENDED] Add EventEmitter interface
│   ├── state.go         [NEW] JobStateStore interface + JobState type
│   ├── planner.go       [EXISTING] JobDiscoverer, JobPlanner
│   └── ...
│
├── app/
│   ├── app.go           [EXTENDED] Services + Bootstrap (resolves #141)
│   ├── scheduler/       [NEW]
│   │   ├── scheduler.go [NEW] Scheduler struct + methods
│   │   └── refresh.go   [NEW] Config refresh loop
│   ├── concurrency/     [NEW]
│   │   └── stack_lock.go [NEW] StackLockManager
│   └── executor/        [EXISTING from M3]
│       └── executor.go  [EXTENDED] Inject EventEmitter
│
├── adapters/
│   ├── events/          [NEW]
│   │   └── log_emitter.go [NEW] LogEmitter adapter
│   ├── state/           [NEW]
│   │   └── memory.go    [NEW] InMemoryStateStore adapter
│   └── ...
│
├── config/schema/
│   └── job_labels.go    [EXTENDED] Add Schedule, OverlapPolicy labels
│
└── cmd/
    ├── daemon.go        [NEW] bosun daemon command
    ├── job_list.go      [NEW] bosun job list command
    └── job_run.go       [EXISTING] No changes needed
```

## Dependencies

### External Libraries

```go
require (
    github.com/robfig/cron/v3 v3.0.1        // Cron scheduling + overlap wrappers
    golang.org/x/sync v0.19.0               // semaphore.Weighted
    github.com/docker/docker v...           // [EXISTING] Docker client
    github.com/spf13/cobra v...             // [EXISTING] CLI framework
)
```

### Internal Dependencies

```
domain/jobs (pure domain types)
    ↑
ports/* (interfaces)
    ↑
    ├─ adapters/* (implementations)
    └─ app/* (orchestration)
        ↑
    cmd/* (CLI entrypoints)
```

No circular dependencies. All arrows point inward (hexagonal architecture).
