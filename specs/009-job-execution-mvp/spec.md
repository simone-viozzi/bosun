# Feature Specification: Job Execution MVP (Milestone 3)

**Feature Branch**: `009-job-execution-mvp`
**Created**: 2025-12-28
**Status**: Draft
**Input**: User description: "Milestone 3: Job Execution MVP - execute jobs by stopping compose stacks, running worker containers with volumes, and restarting stacks. CLI command: bosun job run with --dry-run flag."
**GitHub Issue**: #85

## Overview

This milestone delivers the first real execution capability to Bosun. Given an `ExecutionPlan` (produced by M2's planner), Bosun will orchestrate the actual job execution:

1. **Stop** the target Compose stack (gracefully)
2. **Run** a worker container with the required volumes attached
3. **Restart** the Compose stack

This is "where Bosun does real work" — prior milestones only discovered, validated, and planned.

## GitHub Sub-Issues

### Research (must complete first)
- #109 - Research: Compose Control Strategy (Docker API vs CLI)
- #110 - Research: Worker Container Architecture (Signals, Base Images, Examples)
- #117 - Research: Compose Control Failure Handling (Timeouts, Rollback)

### Port Definitions
- #115 - Task: Define ComposeController port interface
- #116 - Task: Define WorkerRunner port interface
- #114 - Task: Define JobExecutor port interface

### Adapter Implementation
- #118 - Task: Implement ComposeController adapter
- #119 - Task: Implement WorkerRunner adapter

### Application Layer
- #121 - Task: Implement Executor service

### CLI & Integration
- #122 - Task: Implement `bosun job run` CLI command
- #123 - Task: M3 Integration tests for job execution
- #120 - Task: M3 Basic docs update

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Execute a Backup Job (Priority: P1)

As a system administrator, I want to run a backup job that safely stops my application stack, runs a backup worker container, and restarts the stack, so that I can perform consistent backups without manual intervention.

**Why this priority**: This is the core value proposition of Bosun — actually executing backup jobs. Everything else builds on this.

**Independent Test**: Can be fully tested by running `bosun job run <job-name>` against a test Compose stack with a simple worker that touches a file in an attached volume.

**Acceptance Scenarios**:

1. **Given** a running Compose stack "myapp" with a job "daily-backup" defined via labels, **When** I run `bosun job run daily-backup`, **Then** the stack is stopped, the worker runs with volumes attached, and the stack is restarted
2. **Given** a job with multiple target volumes, **When** the job executes, **Then** all specified volumes are mounted in the worker container with correct read/write modes
3. **Given** a worker container that completes successfully (exit code 0), **When** the job finishes, **Then** the CLI reports success and the stack is running again
4. **Given** a worker container that fails (exit code non-zero), **When** the job finishes, **Then** the CLI reports failure with the exit code and logs, and the stack is still restarted

---

### User Story 2 - Preview Job Execution (Dry Run) (Priority: P1)

As a system administrator, I want to preview what a job execution will do without actually performing it, so that I can verify the plan before any containers are stopped.

**Why this priority**: Users need confidence before granting Bosun permission to stop production services. Dry-run is essential for trust.

**Independent Test**: Can be tested by running `bosun job run <job-name> --dry-run` and verifying the output shows the planned steps without executing them.

**Acceptance Scenarios**:

1. **Given** a job "daily-backup", **When** I run `bosun job run daily-backup --dry-run`, **Then** I see the ordered steps (stop stack → run worker → start stack) without any actual execution
2. **Given** dry-run mode, **When** the command completes, **Then** all containers in the stack are still running (no side effects)
3. **Given** dry-run mode with `--format=json`, **When** the command completes, **Then** output is valid JSON describing the execution plan

---

### User Story 3 - Capture Worker Logs (Priority: P2)

As a system administrator, I want to see the logs from the worker container after job execution, so that I can debug failures or verify what the backup actually did.

**Why this priority**: Essential for debugging but not blocking core execution. Users can initially check logs manually via Docker.

**Independent Test**: Can be tested by running a job with a worker that outputs to stdout/stderr and verifying logs are captured.

**Acceptance Scenarios**:

1. **Given** a job execution completes (success or failure), **When** the CLI finishes, **Then** worker stdout/stderr is displayed
2. **Given** a long-running worker, **When** logs are streaming, **Then** I see real-time output (not buffered until completion)
3. **Given** `--quiet` flag, **When** the job runs, **Then** logs are suppressed (only final status shown)

---

### User Story 4 - Handle Execution Timeouts (Priority: P2)

As a system administrator, I want job execution to respect timeouts, so that a hung worker or stuck Compose operation doesn't block indefinitely.

**Why this priority**: Production safety feature. Without timeouts, a single hung job can cause cascading issues.

**Independent Test**: Can be tested by running a job with a worker that sleeps longer than the timeout and verifying it's terminated.

**Acceptance Scenarios**:

1. **Given** a job with a worker that hangs, **When** the timeout is reached, **Then** the worker is terminated and the stack is restarted
2. **Given** a Compose `down` operation that hangs, **When** the timeout is reached, **Then** the operation is aborted with an error
3. **Given** `--timeout` flag override, **When** specified, **Then** that timeout is used instead of the default

---

### User Story 5 - Graceful Stack Stop/Start (Priority: P2)

As a system administrator, I want the stack stop/start to respect container health checks and dependency order, so that my application shuts down and starts up cleanly.

**Why this priority**: Production stability. Incorrect ordering can corrupt data or cause connection errors.

**Independent Test**: Can be tested with a Compose stack that has `depends_on` relationships and health checks, verifying correct order.

**Acceptance Scenarios**:

1. **Given** a stack with `depends_on` relationships, **When** the stack is stopped, **Then** containers are stopped in reverse dependency order
2. ~~**Given** a stack with health checks, **When** the stack is started, **Then** Bosun waits for health checks to pass before reporting success~~ *(DEFERRED to M6+: M3 starts containers without waiting for health)*
3. **Given** a container with a long graceful shutdown, **When** stopping, **Then** SIGTERM is sent first, with SIGKILL only after grace period

---

### Edge Cases

- What happens when the target stack is already stopped?
  - Job should proceed with worker execution and skip the stop step (or warn)
- What happens when the target stack doesn't exist?
  - Error before execution starts: "Stack 'foo' not found"
- What happens when the worker image doesn't exist or can't be pulled?
  - Error with clear message; stack should NOT be stopped if pull fails
- What happens when volumes specified in the job don't exist?
  - Error before execution; don't proceed with partial volume list
- What happens when Docker daemon is unreachable?
  - Clear error message; retry logic out of scope for M3
- What happens when two jobs target the same stack and run concurrently?
  - Out of scope for M3 (no locking/queuing yet); warn in docs
- What happens when the stack fails to restart after worker completion?
  - Report error, leave stack in whatever state it's in, surface error clearly
- What happens when the user interrupts (Ctrl+C) during execution?
  - Graceful handling: attempt to restart stack before exiting (best effort)

## Requirements *(mandatory)*

### Functional Requirements

#### Compose Control (GitHub #115, #118)

- **FR-001**: System MUST be able to stop all containers in a Compose stack (identified by project name)
- **FR-002**: System MUST be able to start all containers in a Compose stack
- **FR-003**: System MUST respect container dependency order during stop/start operations *(DECIDED #109: Use Docker API with labels to determine topology; stop in reverse order, start in forward order)*
- **FR-004**: ~~System MUST wait for health checks during stack startup~~ *(DEFERRED: M3 will NOT wait for health checks - just start containers; health check waiting deferred to M6+)*
- **FR-005**: System MUST support configurable timeouts for stop/start operations *(DECIDED #117: 30s default for stop/start, configurable via `bosun.job.stop-timeout` and `bosun.job.start-timeout` labels, or `--stop-timeout`/`--start-timeout` CLI flags)*

#### Worker Execution (GitHub #116, #119)

- **FR-006**: System MUST create and run a worker container from a specified image
- **FR-007**: System MUST attach specified volumes to the worker container with correct mount modes (ro/rw)
- **FR-008**: System MUST capture worker container exit code and propagate it as job success/failure
- **FR-009**: System MUST capture worker container stdout/stderr logs
- **FR-010**: System MUST pass job metadata to worker via environment variables *(DECIDED #110: Pass `BOSUN_JOB_NAME`, `BOSUN_RUN_ID`, `BOSUN_STACK`, `BOSUN_DRY_RUN` - no BOSUN_VOLUMES as workers should be pre-configured for their mount paths)*
- **FR-011**: System MUST support worker execution timeout with configurable duration *(DECIDED #110: 1h default via `bosun.job.timeout`, SIGTERM → 10s grace → SIGKILL)*
- **FR-012**: System MUST clean up worker container after execution (remove container) *(DECIDED #110: Always remove on success; keep on failure if `--keep-failed` flag)*

#### Orchestration (GitHub #114, #121)

- **FR-013**: System MUST execute job steps in order: stop stack → run worker → start stack
- **FR-014**: System MUST restart the stack even if worker fails (unless configured otherwise) *(DECIDED #117: Always restart by default; `--keep-stopped` CLI flag to skip restart)*
- **FR-015**: System MUST report overall job status (success only if all steps succeed)
- **FR-016**: System MUST support dry-run mode that shows planned actions without execution

#### CLI (GitHub #122)

- **FR-017**: System MUST provide `bosun job run <job-name>` command
- **FR-018**: System MUST support `--dry-run` flag
- **FR-019**: System MUST support `--timeout` flag to override default timeout
- **FR-020**: System MUST support `--format` flag (text, json) for output
- **FR-021**: System MUST support `--quiet` flag to suppress logs

#### Error Handling

- **FR-022**: System MUST validate job exists before execution
- **FR-023**: System MUST validate worker image is accessible before stopping stack *(DECIDED #117: Use ImageInspect to check local cache, then pull if needed - fail fast before stopping stack)*
- **FR-024**: System MUST handle Ctrl+C gracefully (attempt stack restart before exit)
- **FR-025**: System MUST provide clear error messages with actionable guidance

### Key Entities

- **Job**: A named task definition discovered from labels (from M2), containing target stack, worker image, volume attachments, schedule
- **ExecutionPlan**: The ordered steps to execute a job (from M2's planner)
- **JobRun**: A single execution instance of a job, with start time, end time, status, logs
- **ComposeStack**: A group of containers identified by `com.docker.compose.project` label
- **WorkerContainer**: A temporary container created to perform the actual backup/task work

## Research Dependencies *(blocking)*

The following research questions MUST be resolved before implementation:

### Research #109: Compose Control Strategy

**Question**: Should Bosun control Compose stacks via direct Docker API or by shelling out to `docker compose` CLI?

**Impacts**: FR-001, FR-002, FR-003, FR-004, #115, #118

**Options**:
- **A) Docker API**: Full control, no CLI dependency, must replicate Compose's ordering/health logic
- **B) docker compose CLI**: Proven, handles all edge cases, but CLI dependency and text parsing

**Research Plan** (from GitHub #109):
- Research how Compose determines container start order
- Investigate edge cases where API-only approach might fail
- Examine what metadata Compose stores beyond labels
- Review how other tools handle this (Portainer, Dockge)
- Compare performance: API calls vs subprocess

**✅ RESOLVED** - See `.serena/memories/m3_compose_control_decision.md`

**Decision**: Use Docker API directly with label-based discovery and topological sort for dependency ordering. Not using Compose CLI or library. Key rationale: avoids subprocess spawning, full control over execution, cleaner error handling. Dependencies determined from `com.docker.compose.depends_on` labels.

---

### Research #110: Worker Container Architecture

**Question**: What is the minimum contract between Bosun and worker containers?

**Impacts**: FR-010, FR-011, FR-012, #116, #119

**Sub-questions**:
- What environment variables should Bosun inject?
- Should Bosun send timeout warnings before hard kill?
- Should failed workers be kept for debugging?
- Should Bosun provide base images or purely BYOI (Bring Your Own Image)?

**Research Plan** (from GitHub #110):
- How can workers send signals back to Bosun?
- Evaluate options: Unix signals, HTTP callbacks, file-based, stdout parsing
- What notifications are useful? (progress, warnings, completion status)
- Design example workers: pg_dump, MySQL dump, Restic backup, Redis RDB

**✅ RESOLVED** - See `.serena/memories/m3_worker_contract.md`

**Decision**: BYOI (Bring Your Own Image) approach. Workers receive `BOSUN_*` environment variables and must respect SIGTERM for graceful shutdown. No callback mechanism in M3 - workers are fire-and-forget with exit code determining success. Timeout via SIGTERM → 10s → SIGKILL.

---

### Research #117: Compose Failure Handling

**Question**: How should Bosun handle failures during Compose stop/start operations?

**Impacts**: FR-005, FR-014, FR-023, #118, #121

**Sub-questions**:
- What are reasonable default timeouts for down/up operations?
- Should timeouts be configurable per-job (via labels)?
- What signal escalation? (SIGTERM → wait → SIGKILL?)
- If `up` fails after worker, what state do we leave system in?
- Should we pre-validate/pull the worker image before stopping the stack?

**Research Plan** (from GitHub #117):
- `down` hangs: timeout, force kill, or abort?
- `down` fails: retry? abort job? leave stack running?
- `up` fails after worker completed: retry? alert? manual intervention?
- Rollback strategy: If worker fails, should we still `up` the stack?

**✅ RESOLVED** - See `.serena/memories/m3_failure_handling.md`

**Decision**:
- **Timeouts**: 30s default for stop/start, 1h for worker. Configurable via labels.
- **Stop fails**: Abort job, do NOT proceed to worker. Stack may be in inconsistent state.
- **Worker fails**: Still restart stack (availability > backup success). `--keep-stopped` flag to override.
- **Start fails**: Report error, exit non-zero. No automatic retry in M3.
- **Pre-validation**: ImageInspect before stopping stack - fail fast if worker image unavailable.

---

### Additional Clarifications Needed

**Orchestration Behavior** *(RESOLVED)*:
- ✅ **Worker failure**: Stack always restarted by default. `--keep-stopped` flag to skip restart.
- ✅ **Maintenance mode**: Use `--keep-stopped` flag when stack should stay down after worker.

**Logging & Observability** *(PARTIALLY RESOLVED)*:
- ✅ **Log persistence**: M3 displays logs only (stdout/stderr to terminal). Log persistence deferred to M5.
- ✅ **Log streaming**: Real-time streaming via Docker attach (not collect-at-end).

**Concurrency & Locking** *(DEFERRED)*:
- ✅ **M3 scope**: No locking in M3. User responsible for not running concurrent jobs on same stack. Full locking deferred to M4.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: User can execute `bosun job run <job>` and see a Compose stack stop, worker run, and stack restart within 5 minutes for a simple test case
- **SC-002**: Dry-run mode produces accurate output that matches actual execution steps
- **SC-003**: Worker exit code 0 results in job success; non-zero results in job failure with exit code reported
- **SC-004**: Worker logs are visible to user during or after execution
- **SC-005**: Hung workers are terminated after timeout (default or specified)
- **SC-006**: Integration tests pass with real Docker Compose stacks (#123)
- **SC-007**: Ctrl+C during execution results in stack being restarted (best effort)

## Out of Scope (Deferred)

- **Scheduling/Cron**: M4 (Scheduling Engine)
- **Parallel job execution**: M4
- **Job queuing/locking**: M4
- **Metrics/tracing**: M5 (Observability)
- **HTTP API**: Future milestone
- **Retry/backoff logic**: Beyond simple retry, deferred to future
- **Notifications**: Future milestone

## Assumptions

- M2 (Job Model & Planning) is complete and provides `ExecutionPlan` structs
- Docker daemon is accessible via standard socket/API
- User has permissions to stop/start containers and create new containers
- Compose stacks are identified by `com.docker.compose.project` label (standard Compose behavior)
- Worker images are available (local or pullable from configured registries)

## Dependencies

- **Blocked by**: M2 (Job Model & Planning) ✅ Complete
- **Blocks**: M4 (Scheduling Engine), M5 (Observability)
