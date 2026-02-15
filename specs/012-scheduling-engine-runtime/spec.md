# Feature Specification: Scheduling Engine & Runtime

**Feature Branch**: `012-scheduling-engine-runtime`
**Created**: 2026-02-15
**Status**: Draft
**Related Issue**: [#86](https://github.com/simone-viozzi/bosun/issues/86) (M4: Scheduling Engine & Runtime)
**Prerequisites**: [#108](https://github.com/simone-viozzi/bosun/issues/108) (Research: Job Concurrency Strategy) ✅, [#163](https://github.com/simone-viozzi/bosun/issues/163) (M3.75: Critical Bug Fixes) ✅

## Clarifications

### Session 2026-02-15

- Q: How should per-stack mutex locking work when a job targets multiple stacks (`TargetStacks []string`)? → A: Lock all targeted stacks in sorted (alphabetical) order before execution; release all on completion. Sorted acquisition prevents deadlocks when two jobs target overlapping stack sets in different order.
- Q: When a job execution fails, what should the scheduler do with subsequent scheduled runs? → A: Continue scheduling normally (retry on next cron tick), emit failure notification event via EventEmitter, and auto-disable the job after 3 consecutive failures (circuit-breaker). A successful run resets the failure counter.
- Q: How is job identity determined for config refresh change detection? → A: `bosun.job.name` is the unique identity key. Two jobs with the same name are considered the same job, even across different containers. Duplicate names detected during discovery should emit a warning event and use the first-seen definition.
- Q: Should there be a separate `Notifier` port for failure alerts, or should `EventEmitter` handle everything? → A: Single `EventEmitter` port handles all events (lifecycle + alerts). Adapter implementations choose which events to escalate. Implementation of notification adapters (Slack, email, etc.) is not part of this spec.
- Q: What happens to the consecutive failure counter when the daemon restarts? → A: Reset to 0 on daemon restart (in-memory counter is lost, all jobs start fresh). Daemon restart acts as a natural circuit-breaker reset. Persisting failure state to disk is deferred.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Schedule Jobs with Cron Expressions (Priority: P1)

As a DevOps engineer, I want to define job schedules using standard cron expressions in Docker labels so that routine backup and maintenance tasks run automatically without manual intervention.

**Why this priority**: Core functionality — without scheduling, this is not a scheduling engine. This enables the primary use case of automated job orchestration.

**Independent Test**: Define a job with `bosun.job.schedule=*/5 * * * *` and verify it executes every 5 minutes via daemon logs. Delivers immediate value as automated execution.

**Acceptance Scenarios**:

1. **Given** a job configured with `bosun.job.schedule=0 3 * * *`, **When** the daemon discovers and registers the job, **Then** the job executes daily at 03:00.
2. **Given** a job with `bosun.job.schedule=*/15 * * * *`, **When** time advances, **Then** the job executes every 15 minutes.
3. **Given** a job with an invalid cron expression, **When** the scheduler attempts to parse it, **Then** an error event is emitted and the job is skipped.
4. **Given** multiple jobs with different schedules, **When** the scheduler runs, **Then** each job executes according to its individual schedule.

---

### User Story 2 - Control Job Overlap with Policies (Priority: P1)

As a DevOps engineer, I want to control what happens when a job is still running when its next scheduled execution time arrives, using overlap policies (`queue` or `skip`), so that I can prevent resource conflicts and ensure safe execution.

**Why this priority**: Critical for production reliability. Without overlap control, long-running jobs can cause resource exhaustion or data corruption.

**Independent Test**: Create a job with 2-minute schedule and 5-minute execution time. Set `bosun.job.overlap-policy=skip` and verify subsequent runs are skipped until first completes.

**Acceptance Scenarios**:

1. **Given** a job with `overlap-policy=queue` that is still running, **When** the next scheduled time arrives, **Then** the new execution is queued and starts after the current run completes.
2. **Given** a job with `overlap-policy=skip` that is still running, **When** the next scheduled time arrives, **Then** the new execution is skipped with an event emitted.
3. **Given** a job with no explicit overlap policy, **When** overlap occurs, **Then** the default behavior is `queue` (safe by default).
4. **Given** a job completes before its next schedule, **When** the scheduled time arrives, **Then** the job executes normally regardless of overlap policy.

---

### User Story 3 - Run Bosun as a Long-Lived Daemon (Priority: P1)

As a system administrator, I want to run `bosun daemon` as a long-lived background process that continuously monitors schedules and executes jobs, so that I can deploy Bosun as a systemd service or Docker container that runs indefinitely.

**Why this priority**: Essential infrastructure. Users need a reliable daemon process for production deployment. This is the runtime that makes scheduling work.

**Independent Test**: Start `bosun daemon`, verify it stays running, executes scheduled jobs, responds to signals, and can be gracefully stopped with SIGTERM.

**Acceptance Scenarios**:

1. **Given** jobs are configured with schedules, **When** `bosun daemon` starts, **Then** all enabled jobs are discovered and registered with the scheduler.
2. **Given** the daemon is running, **When** a SIGTERM or SIGINT signal is received, **Then** the daemon initiates graceful shutdown (waits for running jobs to complete, then exits).
3. **Given** the daemon is running, **When** a double-signal (two SIGTERM/SIGINT) is received, **Then** the daemon cancels running jobs immediately and exits.
4. **Given** the daemon has been running for hours, **When** checked, **Then** it remains stable without memory leaks or goroutine leaks.

---

### User Story 4 - Automatically Refresh Configuration (Priority: P1)

As a DevOps engineer, I want the daemon to periodically refresh its configuration from Docker labels so that I can add, modify, or remove jobs without restarting the daemon.

**Why this priority**: Operational convenience and uptime. Avoids service disruption for configuration changes. Makes Bosun feel more dynamic and cloud-native.

**Independent Test**: Start daemon with 1 job, add a new job via Docker Compose label, wait for refresh interval, verify new job is discovered and scheduled.

**Acceptance Scenarios**:

1. **Given** the daemon is running with job A, **When** a new job B is added to Docker labels and config refresh occurs, **Then** job B is discovered and registered with the scheduler.
2. **Given** the daemon is running with job A, **When** job A's schedule is changed in Docker labels and refresh occurs, **Then** job A is removed and re-registered with the new schedule.
3. **Given** the daemon is running with job A, **When** job A is removed from Docker labels and refresh occurs, **Then** job A is unregistered (current run completes if active).
4. **Given** a job is disabled (`bosun.job.enabled=false`), **When** config refresh occurs, **Then** the job is not registered or is removed if previously registered.

---

### User Story 5 - Enforce Concurrency Limits (Per-Stack and Global) (Priority: P1)

As a DevOps engineer, I want automatic per-stack locking to prevent concurrent jobs on the same stack, and a configurable global semaphore to limit total parallel jobs, so that I can control resource usage and prevent conflicts.

**Why this priority**: Safety and predictability. Prevents race conditions (per-stack) and resource exhaustion (global limit). This is the foundation of the three-layer concurrency model from #108 research.

**Independent Test**: Configure 2 jobs targeting the same stack (A and B), and 1 job targeting a different stack (C). Set global limit to 1. Start all three. Verify only 1 runs at a time, and A→B are serialized.

**Acceptance Scenarios**:

1. **Given** two jobs (A and B) targeting the same stack, **When** both are scheduled simultaneously, **Then** only one executes at a time (per-stack mutex enforced automatically).
2. **Given** global semaphore is set to N=1, **When** multiple jobs are ready to execute, **Then** only 1 job runs at a time across all stacks.
3. **Given** global semaphore is set to N=3, **When** 5 jobs are ready to execute, **Then** at most 3 jobs run concurrently.
4. **Given** a job is blocked waiting for per-stack lock or global semaphore, **When** resources become available, **Then** the job starts execution and emits a "job started" event.

---

### User Story 6 - List Currently Scheduled Jobs (Priority: P2)

As a DevOps engineer, I want to run `bosun job list` to see all scheduled jobs, their status (running/idle), last run time, next run time, and overlap policy, so that I can monitor and debug scheduling behavior.

**Why this priority**: Important for observability but not critical for core scheduling functionality. Can be added after the scheduler is working.

**Independent Test**: Start daemon with 3 jobs, run `bosun job list`, verify output shows all 3 jobs with accurate status and timing information.

**Acceptance Scenarios**:

1. **Given** the daemon is running with multiple jobs, **When** `bosun job list` is executed, **Then** all jobs are listed with their names, schedules, last run time, next run time, and status.
2. **Given** a job is currently executing, **When** `bosun job list` is executed, **Then** the job's status shows as "running" with start time.
3. **Given** a job completed 5 minutes ago, **When** `bosun job list` is executed, **Then** the job shows last run time, result (success/failure), and next scheduled time.
4. **Given** no jobs are configured, **When** `bosun job list` is executed, **Then** output indicates no jobs are scheduled.

---

### User Story 7 - Disable Jobs Without Removing Them (Priority: P2)

As a DevOps engineer, I want to temporarily disable a job by setting `bosun.job.enabled=false` so that I can pause automation without removing configuration, and re-enable it later.

**Why this priority**: Useful operational feature but not essential for initial deployment. Can be implemented after core scheduling works.

**Independent Test**: Set job's `enabled=false` label, verify daemon does not schedule it. Change to `enabled=true`, wait for config refresh, verify job is scheduled.

**Acceptance Scenarios**:

1. **Given** a job with `bosun.job.enabled=false`, **When** the daemon discovers jobs, **Then** the job is not registered with the scheduler.
2. **Given** a running job is changed to `enabled=false`, **When** config refresh occurs, **Then** the job is removed from the scheduler (current run completes).
3. **Given** a disabled job is changed to `enabled=true`, **When** config refresh occurs, **Then** the job is registered and scheduled normally.
4. **Given** a job has no explicit `enabled` label, **When** the daemon discovers it, **Then** it defaults to `enabled=true`.

---

### User Story 8 - Cancel-and-Restart Overlap Policy (Priority: P3 - Deferred)

As a DevOps engineer, I want the option to use `overlap-policy=cancel-and-restart` so that when a new scheduled time arrives, the currently running instance is terminated and a fresh execution begins.

**Why this priority**: Advanced feature with implementation complexity. Deferred to [#176](https://github.com/simone-viozzi/bosun/issues/176) — `queue` and `skip` are sufficient for M4.

**Independent Test**: [Deferred to #176]

**Acceptance Scenarios**: [Deferred to #176]

---

### Edge Cases

- What happens when a job's schedule is changed while it's executing?
  - **Answer**: Current run completes with old schedule. Job is re-registered with new schedule. Next run uses new schedule.
- What happens when Docker daemon becomes unavailable during scheduled execution?
  - **Answer**: Job fails with timeout error. Scheduler continues running. Next scheduled execution retries.
- What happens when config refresh discovers 100 new jobs simultaneously?
  - **Answer**: All jobs are registered. Actual execution is throttled by global semaphore (default N=1).
- What happens if two jobs have overlapping schedules but no stack conflicts?
  - **Answer**: Both execute concurrently (up to global semaphore limit). Per-stack mutex only applies to stack conflicts.
- What happens when daemon is killed (SIGKILL) while jobs are running?
  - **Answer**: Running worker containers keep executing. On restart, daemon re-discovers jobs and resumes scheduling. No orphan cleanup in M4.
- What happens when the daemon restarts and a weekly job missed its schedule during downtime?
  - **Answer**: **Known M4 limitation.** The cron library is forward-looking only — on startup it calculates the next run from "now" with no awareness of missed runs. A weekly job restarted on Monday won't run until next Sunday. Persistent scheduling (storing `last_run_at` per job and firing catch-up runs on startup, similar to systemd `Persistent=true`) is deferred to high-priority follow-up (M4.5 or early M5).

## Requirements *(mandatory)*

### Functional Requirements

#### Scheduling Engine

- **FR-001**: System MUST parse standard cron expressions (5-field and 6-field with seconds)
- **FR-002**: System MUST register discovered jobs with their cron schedules using `robfig/cron/v3`
- **FR-003**: System MUST trigger job execution at scheduled times via the `JobExecutor` interface
- **FR-004**: System MUST support cron expressions with minute-level granularity (minimum)
- **FR-005**: System MUST handle invalid cron expressions by logging errors and skipping job registration

#### Overlap Policies

- **FR-006**: System MUST support `overlap-policy=queue` (delay next run until current completes)
- **FR-007**: System MUST support `overlap-policy=skip` (drop next run if current is active)
- **FR-008**: System MUST default to `overlap-policy=queue` when not specified
- **FR-009**: System MUST emit events when jobs are skipped due to overlap

#### Failure Handling & Circuit-Breaker

- **FR-037**: System MUST continue scheduling a job normally after a failure (next run fires at its cron time)
- **FR-038**: System MUST emit a failure notification event via `EventEmitter` when a job execution fails
- **FR-039**: System MUST track consecutive failure count per job (reset to 0 on success)
- **FR-040**: System MUST auto-disable a job after 3 consecutive failures (circuit-breaker) and emit a `JobCircuitBroken` event
- **FR-041**: An auto-disabled job MUST NOT be re-enabled by config refresh; manual intervention (re-setting `bosun.job.enabled=true`) is required

#### Concurrency Control

- **FR-010**: System MUST enforce per-stack mutex automatically (jobs targeting same stack never run concurrently). For multi-stack jobs, all targeted stacks MUST be locked in sorted (alphabetical) order before execution to prevent deadlocks, and released on completion.
- **FR-011**: System MUST enforce global semaphore limit (default N=1, configurable via CLI flag `--parallelism`)
- **FR-012**: System MUST track which stacks are currently locked and release locks after execution
- **FR-013**: System MUST prevent deadlocks via sorted lock acquisition order for multi-stack jobs (no timeouts needed)

#### Daemon Lifecycle

- **FR-014**: System MUST start as a long-lived daemon process via `bosun daemon` command
- **FR-015**: System MUST discover all enabled jobs on startup and register them with the scheduler
- **FR-016**: System MUST handle SIGTERM/SIGINT for graceful shutdown
- **FR-017**: System MUST wait for running jobs to complete during graceful shutdown (with timeout)
- **FR-018**: System MUST support double-signal (immediate exit) that cancels running jobs
- **FR-019**: System MUST log daemon lifecycle events (started, shutdown initiated, shutdown complete)

#### Config Refresh

- **FR-020**: System MUST periodically refresh configuration from Docker labels (default interval: 5 minutes)
- **FR-021**: System MUST detect new jobs and register them with the scheduler
- **FR-022**: System MUST detect removed jobs and unregister them (current run completes)
- **FR-023**: System MUST detect changed job schedules/policies and re-register jobs
- **FR-024**: System MUST respect `bosun.job.enabled=false` and not schedule disabled jobs
- **FR-025**: System MUST emit events for job additions, removals, and changes

#### Job Status & Observability

- **FR-026**: System MUST track job status (idle, running, completed, failed) in memory
- **FR-027**: System MUST record last run time, last result, and next scheduled time for each job
- **FR-028**: System MUST expose job status via `bosun job list` command
- **FR-029**: System MUST emit events for job lifecycle (scheduled, started, completed, failed, skipped)
- **FR-030**: Events MUST be handled by pluggable `EventEmitter` port interface

#### Architecture & Dependencies

- **FR-031**: System MUST use hexagonal architecture with ports and adapters
- **FR-032**: CLI commands MUST NOT import adapters directly (use app-level service factory from #168)
- **FR-033**: System MUST define `EventEmitter` port interface for event notifications
- **FR-034**: System MUST implement log-based `EventEmitter` adapter as default
- **FR-035**: System MUST use `robfig/cron/v3` for cron parsing and scheduling
- **FR-036**: System MUST use `golang.org/x/sync/semaphore` for global concurrency control

### Key Entities *(include if feature involves data)*

- **ScheduledJob**: Represents a job registered with the scheduler
  - Attributes: JobName (unique identity key), Schedule (cron), OverlapPolicy, Enabled flag, ConsecutiveFailures (int, for circuit-breaker)
  - Identity: `bosun.job.name` is the unique key — same name across different containers refers to the same job
  - Relationships: References `jobs.Job` from domain, wrapped by cron library

- **JobStatus**: Represents current state of a scheduled job
  - Attributes: JobName, Status (idle/running/completed/failed), LastRunTime, LastResult, NextRunTime, OverlapPolicy
  - Used by: `bosun job list` command, event emission

- **StackLock**: Represents a per-stack mutex
  - Attributes: StackName, Locked (bool), LockedBy (job name), LockedAt (timestamp)
  - Behavior: Acquired before execution, released after completion (via defer)

- **GlobalSemaphore**: Represents system-wide concurrency limit
  - Attributes: MaxParallelism (N), CurrentlyRunning (count)
  - Behavior: Acquire slot before execution, release after completion

- **Event**: Represents a lifecycle event for observability
  - Attributes: EventType (scheduled/started/completed/failed/skipped/circuit-broken), JobName, Timestamp, Metadata
  - Handled by: `EventEmitter` port implementations (single port for lifecycle + alerts; adapters choose what to escalate)
  - Note: Only the log adapter is implemented in M4; notification adapters (Slack, email) are out of scope

- **OverlapPolicy**: Enum defining overlap behavior
  - Values: `queue`, `skip`, `cancel-and-restart` (deferred)
  - Used by: Scheduler to wrap jobs with appropriate cron library decorators

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Daemon runs continuously for 72+ hours without restarts during normal operation
- **SC-002**: Jobs execute on schedule with <10 second jitter from cron target time (under normal load)
- **SC-003**: Config refresh detects and registers new jobs within 5 minutes (default refresh interval)
- **SC-004**: Graceful shutdown completes within 60 seconds when no long-running jobs are active
- **SC-005**: Per-stack mutex prevents concurrent execution on the same stack 100% of the time
- **SC-006**: Global semaphore correctly limits parallelism (N=1: serial execution, N=3: max 3 concurrent jobs)
- **SC-007**: `bosun job list` returns results within 1 second for typical workloads (<100 jobs)
- **SC-008**: Overlap policies work correctly (queue: runs are serialized, skip: excess runs are dropped)
- **SC-009**: Daemon memory usage remains stable (<50MB growth over 24 hours for typical workloads)
- **SC-010**: Integration tests pass consistently with real Docker daemon (race detector enabled)

### Architecture Outcomes

- **SC-011**: No direct adapter imports in `internal/cmd/` package (verified via `grep -r "internal/adapters" internal/cmd/` = 0 results)
- **SC-012**: App bootstrap service factory resolves #141 (CLI→adapter coupling) permanently
- **SC-013**: All port interfaces are properly defined and used by app layer
- **SC-014**: Event system is extensible (can add new EventEmitter implementations without changing scheduler)

## Dependencies *(optional)*

### Prerequisites Complete

- ✅ [#108](https://github.com/simone-viozzi/bosun/issues/108) - Research: Job Concurrency Strategy (three-layer model design)
- ✅ [#163](https://github.com/simone-viozzi/bosun/issues/163) - M3.75: Critical Bug Fixes

### Required Libraries (Already in go.mod)

- `robfig/cron/v3 v3.0.1` - Cron parsing and scheduling with overlap policy wrappers
- `golang.org/x/sync v0.19.0` - Semaphore for global concurrency control

### Blocks Future Work

- #87 (M5: Observability & Robustness) - Requires stable daemon + event emission infrastructure

## Assumptions *(optional)*

- **Docker availability**: Daemon assumes Docker socket is accessible during config refresh
- **Label-based config**: All job configuration comes from Docker labels (no external config files in M4)
- **Single-host deployment**: No distributed locking or multi-host scheduling in M4
- **In-memory state**: Job status, scheduler state, and failure counters are ephemeral (no persistence in M4). **Known limitation**: missed runs during downtime are not detected or caught up (see Edge Cases).
- **Sequential default**: Global semaphore defaults to N=1 (serial execution) for safety
- **No orphan cleanup**: Daemon does not track or clean up orphaned worker containers on startup (deferred to vNext)
- **Config refresh granularity**: 5-minute default refresh is acceptable for M4 (can be made configurable later)

## Out of Scope *(optional)*

### Explicitly Deferred

- **Cancel-and-restart overlap policy** → [#176](https://github.com/simone-viozzi/bosun/issues/176) (P2, deferred)
- **Event-based triggers** → [#28](https://github.com/simone-viozzi/bosun/issues/28) (vNext)
- **Config hot-reload via file watch** → [#63](https://github.com/simone-viozzi/bosun/issues/63) (vNext)
- **HTTP API for daemon control** → M5+ (CLI-only for v1)
- **Prometheus metrics** → M5+ (observability milestone)
- **Persistent job history** → M5+ (currently log-only)
- **Notification adapters** (Slack, email, webhook) → M5+ (EventEmitter port is ready; only log adapter in M4)
- **Persistent scheduling / catch-up runs** → [#177](https://github.com/simone-viozzi/bosun/issues/177), M4.5 or early M5 (**high priority**). Store `last_run_at` per job to disk; on startup, fire catch-up runs for jobs whose schedule was missed during downtime (systemd `Persistent=true` style). Required for weekly/monthly jobs to be reliable across daemon restarts.
- **Distributed locking** → vNext (multi-host support)
- **Health checks before job execution** → vNext (wait for service readiness)

### Never Planned

- Multi-host / cluster scheduling (use K8s CronJobs instead)
- Job dependencies / DAG execution (use external orchestrator)
- Event-based triggers from external systems (use webhooks to call `bosun job run`)
