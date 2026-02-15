# Tasks: Scheduling Engine & Runtime

**Input**: Design documents from `/specs/012-scheduling-engine-runtime/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Included — Constitution Principle III (Test-First Development) mandates unit and integration tests.

**Organization**: Tasks grouped by user story (US1–US7) for independent implementation and testing. US8 (cancel-and-restart) deferred to #176.

**Issue Mapping**:

| Issue | Scope | Phase |
|-------|-------|-------|
| [#168](https://github.com/simone-viozzi/bosun/issues/168) | App Bootstrap / Service Factory | Foundational |
| [#169](https://github.com/simone-viozzi/bosun/issues/169) | Domain Types (OverlapPolicy, Schedule, Enabled) | Foundational |
| [#170](https://github.com/simone-viozzi/bosun/issues/170) | EventEmitter Port + LogEmitter Adapter | Foundational |
| [#171](https://github.com/simone-viozzi/bosun/issues/171) | Scheduler Core (cron integration) | US1, US2 |
| [#172](https://github.com/simone-viozzi/bosun/issues/172) | Config Refresh (periodic re-discovery) | US4 |
| [#173](https://github.com/simone-viozzi/bosun/issues/173) | Concurrency Control (Stack Locks + Global Semaphore) | US5 |
| [#174](https://github.com/simone-viozzi/bosun/issues/174) | Daemon Command (`bosun daemon`) | US3 |
| [#175](https://github.com/simone-viozzi/bosun/issues/175) | Job List Command (`bosun job list`) | US6 |
| [#176](https://github.com/simone-viozzi/bosun/issues/176) | Cancel-and-Restart Policy | DEFERRED |
| [#177](https://github.com/simone-viozzi/bosun/issues/177) | Persistent Scheduling (port in M4, adapter later) | Foundational (port only) |

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Exact file paths included in every task description

---

### T1 - App Bootstrap & Service Factory (#168)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create package directories and register new configuration labels

- [X] T001 Create new package directories: internal/app/scheduler/, internal/app/concurrency/, internal/adapters/events/, internal/adapters/state/
- [X] T002 [P] Register schedule, overlap-policy, enabled labels in internal/config/schema/job_labels.go (#169)
- [X] T003 [P] Add daemon and job sub-command stubs to internal/cmd/root.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types, port interfaces, base adapters, and app bootstrap — everything user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Domain Types (#169)

- [X] T004 [P] Implement OverlapPolicy type, constants, and ValidateOverlapPolicy() in internal/domain/jobs/overlap.go
- [X] T005 [P] Extend Job struct with Schedule, OverlapPolicy, Enabled fields in internal/domain/jobs/types.go
- [X] T006 [P] Implement JobStatus and RunStatus types in internal/domain/jobs/status.go

### Port Interfaces (#170, #177)

- [X] T007 [P] Define EventEmitter port interface (10 methods per contracts/event_emitter.go) in internal/ports/events.go
- [X] T008 [P] Define JobStateStore port interface, JobState type, and ErrJobStateNotFound sentinel in internal/ports/state.go

### Adapters (#170, #177)

- [X] T009 [P] Implement LogEmitter adapter (slog-based, all 10 EventEmitter methods) in internal/adapters/events/log_emitter.go
- [X] T010 [P] Implement InMemoryStateStore adapter (sync.Map-backed, 4 CRUD methods) in internal/adapters/state/memory.go

### App Bootstrap (#168)

- [X] T011 Implement Services struct and Bootstrap() factory resolving #141 in internal/app/app.go (depends on T004–T010)

### Foundation Tests

- [X] T012 [P] Unit test OverlapPolicy validation (valid, invalid, deferred cancel-and-restart) in internal/domain/jobs/overlap_test.go
- [X] T013 [P] Unit test LogEmitter — verify all event methods emit structured slog output in internal/adapters/events/log_emitter_test.go
- [X] T014 [P] Unit test InMemoryStateStore — SaveJobState, LoadJobState, ListJobStates, DeleteJobState, ErrJobStateNotFound in internal/adapters/state/memory_test.go
- [X] T015 [P] Unit test Bootstrap — verify Services fields are populated correctly in internal/app/app_test.go

**Checkpoint**: Foundation ready — domain types, ports, adapters, and bootstrap wired. User story implementation can begin.

---

## Phase 3: User Story 1 — Schedule Jobs with Cron Expressions (Priority: P1) 🎯 MVP

**Goal**: Jobs with `bosun.job.schedule` labels are automatically parsed via `robfig/cron/v3` and executed at cron-scheduled times through the existing `JobExecutor`

**Independent Test**: Define a job with `bosun.job.schedule=*/5 * * * *`, start the scheduler, verify it triggers execution every 5 minutes via daemon logs

**GitHub Issues**: #171

**FRs**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-037, FR-038, FR-039, FR-040, FR-041

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T016 [P] [US1] Unit test Scheduler.AddJob — valid cron registered, invalid cron returns error in internal/app/scheduler/scheduler_test.go
- [X] T017 [P] [US1] Unit test Scheduler.executeJob — mock executor success updates status and resets failures in internal/app/scheduler/scheduler_test.go
- [X] T018 [P] [US1] Unit test circuit-breaker — 3 consecutive failures auto-disables job, emits JobCircuitBroken in internal/app/scheduler/scheduler_test.go

### Implementation for User Story 1

- [X] T019 [US1] Implement Scheduler struct with New() constructor (cron.Cron, executor, events, stateStore, stackLocks, globalSem, statusMap, entries) in internal/app/scheduler/scheduler.go
- [X] T020 [US1] Implement AddJob — parse cron expression, register with cron.Cron, emit JobScheduled event, track in entries map in internal/app/scheduler/scheduler.go
- [X] T021 [US1] Implement RemoveJob — cron.Remove(entryID), clean up entries map, emit JobRemoved event in internal/app/scheduler/scheduler.go
- [X] T022 [US1] Implement executeJob — acquire global semaphore, call Executor.Execute, update statusMap, emit Started/Completed/Failed events in internal/app/scheduler/scheduler.go
- [X] T023 [US1] Implement circuit-breaker in executeJob — track consecutiveFailures per entry, auto-disable at 3, emit JobCircuitBroken, reset counter on success in internal/app/scheduler/scheduler.go
- [X] T024 [US1] Persist job state after each execution via stateStore.SaveJobState() in internal/app/scheduler/scheduler.go
- [X] T025 [US1] Implement Start — start cron.Cron, block on context cancellation in internal/app/scheduler/scheduler.go
- [X] T026 [US1] Implement Stop — stop cron.Cron, wait for running jobs via WaitGroup in internal/app/scheduler/scheduler.go
- [X] T027 [US1] Implement ListJobs — iterate statusMap, return []JobStatus with NextRunTime from cron entries in internal/app/scheduler/scheduler.go

**Checkpoint**: Core scheduler works — jobs are registered via cron, executed through JobExecutor, status is tracked, circuit-breaker auto-disables failing jobs. This is the MVP.

---

## Phase 4: User Story 2 — Control Job Overlap with Policies (Priority: P1)

**Goal**: Jobs respect `bosun.job.overlap-policy` label — `queue` serializes overlapping runs, `skip` drops them

**Independent Test**: Create a job with 2-minute schedule and 5-minute execution time. Set `overlap-policy=skip`, verify subsequent runs are skipped until first completes.

**GitHub Issues**: #171

**FRs**: FR-006, FR-007, FR-008, FR-009

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T028 [P] [US2] Unit test queue overlap — second invocation blocks until first completes in internal/app/scheduler/scheduler_test.go
- [X] T029 [P] [US2] Unit test skip overlap — second invocation is dropped with JobSkipped event in internal/app/scheduler/scheduler_test.go
- [X] T030 [P] [US2] Unit test default overlap policy is queue when OverlapPolicy field is empty in internal/app/scheduler/scheduler_test.go

### Implementation for User Story 2

- [X] T031 [US2] Wrap cron job with cron.DelayIfStillRunning when overlap-policy=queue in AddJob in internal/app/scheduler/scheduler.go
- [X] T032 [US2] Wrap cron job with cron.SkipIfStillRunning when overlap-policy=skip in AddJob in internal/app/scheduler/scheduler.go
- [X] T033 [US2] Emit JobSkipped event when skip policy drops a scheduled run in internal/app/scheduler/scheduler.go
- [X] T034 [US2] Default to OverlapPolicyQueue in AddJob when policy is empty string in internal/app/scheduler/scheduler.go

**Checkpoint**: Overlap policies work — `queue` jobs serialize, `skip` jobs drop excess runs with event notification.

---

## Phase 5: User Story 5 — Enforce Concurrency Limits (Priority: P1)

**Goal**: Three-layer concurrency model operational: per-stack mutexes prevent concurrent jobs on the same Compose stack, global semaphore caps total parallelism

**Independent Test**: Configure 2 jobs on same stack (A, B) and 1 on different stack (C). Set --parallelism=1. Start all three. Verify only 1 runs at a time and A/B are serialized.

**GitHub Issues**: #173

**FRs**: FR-010, FR-011, FR-012, FR-013

### Tests for User Story 5

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T035 [P] [US5] Unit test StackLockManager.Lock/Unlock — mutual exclusion on same stack in internal/app/concurrency/stack_lock_test.go
- [X] T036 [P] [US5] Unit test StackLockManager.LockAll — sorted alphabetical acquisition prevents deadlocks in internal/app/concurrency/stack_lock_test.go
- [X] T037 [P] [US5] Unit test StackLockManager.LockAll — context cancellation releases already-acquired locks in internal/app/concurrency/stack_lock_test.go
- [X] T038 [P] [US5] Unit test global semaphore — N=1 enforces serial execution, N=3 allows 3 concurrent in internal/app/scheduler/scheduler_test.go

### Implementation for User Story 5

- [X] T039 [US5] Implement StackLockManager struct with Lock/Unlock using sync.Map of mutexes in internal/app/concurrency/stack_lock.go
- [X] T040 [US5] Implement LockAll/UnlockAll with sorted (alphabetical) acquisition and rollback on error in internal/app/concurrency/stack_lock.go
- [X] T041 [US5] Implement IsLocked for observability/debugging in internal/app/concurrency/stack_lock.go
- [X] T042 [US5] Integrate StackLockManager into Scheduler.executeJob — LockAll before Execute, UnlockAll via defer in internal/app/scheduler/scheduler.go
- [X] T043 [US5] Wire global semaphore (semaphore.Weighted) into Scheduler via BootstrapOptions.Parallelism in internal/app/app.go

**Checkpoint**: Three-layer concurrency model complete — per-job overlap (Layer 1), global semaphore (Layer 2), per-stack mutex (Layer 3) all enforced.

---

## Phase 6: User Story 3 — Run Bosun as a Long-Lived Daemon (Priority: P1)

**Goal**: `bosun daemon` starts a reliable long-lived process that runs the scheduler, handles OS signals, and shuts down gracefully

**Independent Test**: Start `bosun daemon`, verify it stays running, executes scheduled jobs, responds to SIGTERM with graceful shutdown.

**GitHub Issues**: #174

**FRs**: FR-014, FR-015, FR-016, FR-017, FR-018, FR-019

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T044 [P] [US3] Unit test daemon signal handling — SIGTERM triggers graceful Scheduler.Stop in internal/cmd/daemon_test.go
- [X] T045 [P] [US3] Unit test double-signal — second SIGTERM/SIGINT cancels context for immediate exit in internal/cmd/daemon_test.go

### Implementation for User Story 3

- [X] T046 [US3] Implement daemon Cobra command with --parallelism and --refresh-interval flags in internal/cmd/daemon.go
- [X] T047 [US3] Implement daemon run function — Bootstrap → discover jobs → Scheduler.Start → block on signals in internal/cmd/daemon.go
- [X] T048 [US3] Implement signal handler — first SIGTERM/SIGINT calls Scheduler.Stop with 60s timeout in internal/cmd/daemon.go
- [X] T049 [US3] Implement double-signal handler — second SIGTERM/SIGINT cancels root context for immediate exit in internal/cmd/daemon.go
- [X] T050 [US3] Log daemon lifecycle events (started, shutdown initiated, shutdown complete) via slog in internal/cmd/daemon.go
- [X] T051 [US3] Define daemon exit codes (0=clean shutdown, 1=error) in internal/cmd/exitcodes.go
- [X] T052 [US3] Register daemon command under root in internal/cmd/root.go

**Checkpoint**: `bosun daemon` runs as a stable long-lived process. Graceful shutdown on SIGTERM, force exit on double-signal.

---

## Phase 7: User Story 4 — Automatically Refresh Configuration (Priority: P1)

**Goal**: Daemon periodically re-discovers jobs from Docker labels and detects additions, removals, and schedule/policy changes without restart

**Independent Test**: Start daemon with 1 job. Add a new job via Docker Compose label. Wait for refresh interval. Verify new job is discovered and scheduled.

**GitHub Issues**: #172

**FRs**: FR-020, FR-021, FR-022, FR-023, FR-024, FR-025, FR-041

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T053 [P] [US4] Unit test diff logic — detect new jobs added since last refresh in internal/app/scheduler/refresh_test.go
- [X] T054 [P] [US4] Unit test diff logic — detect removed jobs and changed schedules/policies in internal/app/scheduler/refresh_test.go
- [X] T055 [P] [US4] Unit test circuit-broken jobs are NOT re-enabled by config refresh (FR-041) in internal/app/scheduler/refresh_test.go

### Implementation for User Story 4

- [X] T056 [US4] Implement diff function — compare discovered jobs vs registered entries by job name in internal/app/scheduler/refresh.go
- [X] T057 [US4] Handle new jobs in diff — call AddJob, emit JobAdded event in internal/app/scheduler/refresh.go
- [X] T058 [US4] Handle removed jobs in diff — call RemoveJob (current run completes), emit JobRemoved event in internal/app/scheduler/refresh.go
- [X] T059 [US4] Handle changed jobs in diff — RemoveJob + AddJob, emit JobChanged event with old/new schedule in internal/app/scheduler/refresh.go
- [X] T060 [US4] Preserve circuit-broken state across refresh — auto-disabled jobs are NOT re-enabled by refresh (FR-041) in internal/app/scheduler/refresh.go
- [X] T061 [US4] Implement refresh loop — time.Ticker goroutine calling diff, stop on context cancel in internal/app/scheduler/refresh.go
- [X] T062 [US4] Integrate refresh loop into Scheduler.Start (start ticker goroutine) in internal/app/scheduler/scheduler.go

**Checkpoint**: Config hot-reload works — add, modify, or remove jobs by changing Docker labels. No daemon restart needed.

---

## Phase 8: User Story 6 — List Currently Scheduled Jobs (Priority: P2)

**Goal**: `bosun job list` displays all scheduled jobs with status (idle/running), last run time, next run time, and overlap policy

**Independent Test**: Start daemon with 3 jobs, run `bosun job list`, verify output shows all 3 with accurate status and timing info.

**GitHub Issues**: #175

**FRs**: FR-026, FR-027, FR-028, FR-029, FR-030

### Tests for User Story 6

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T063 [P] [US6] Unit test text table output formatting (columns, alignment) in internal/cmd/job_list_test.go
- [X] T064 [P] [US6] Unit test JSON and YAML output formats in internal/cmd/job_list_test.go

### Implementation for User Story 6

- [X] T065 [US6] Implement job list Cobra command with --format text|json|yaml flag in internal/cmd/job_list.go
- [X] T066 [US6] Implement text table renderer — columns: Name, Schedule, Status, Last Run, Next Run, Overlap Policy in internal/cmd/job_list.go
- [X] T067 [US6] Implement JSON and YAML output formats using encoding/json and gopkg.in/yaml.v3 in internal/cmd/job_list.go
- [X] T068 [US6] Register job list sub-command under job parent command in internal/cmd/root.go

**Checkpoint**: `bosun job list` provides full job visibility with text/JSON/YAML output. SC-007: response within 1 second for <100 jobs.

---

## Phase 9: User Story 7 — Disable Jobs Without Removing Them (Priority: P2)

**Goal**: Jobs can be paused via `bosun.job.enabled=false` label and re-enabled without removing or changing other configuration

**Independent Test**: Set job's `enabled=false` label, verify daemon does not schedule it. Change to `enabled=true`, wait for config refresh, verify job resumes scheduling.

**GitHub Issues**: #169, #172

**FRs**: FR-024

### Tests for User Story 7

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T069 [P] [US7] Unit test disabled job (Enabled=false) is skipped during initial discovery in internal/app/scheduler/scheduler_test.go
- [X] T070 [P] [US7] Unit test enabled→disabled transition on config refresh removes job in internal/app/scheduler/refresh_test.go
- [X] T071 [P] [US7] Unit test disabled→enabled transition on config refresh registers job in internal/app/scheduler/refresh_test.go

### Implementation for User Story 7

- [X] T072 [US7] Filter disabled jobs (Enabled=false) during initial job discovery in Scheduler.Start in internal/app/scheduler/scheduler.go
- [X] T073 [US7] Handle enabled→disabled transition in refresh diff — RemoveJob for newly-disabled jobs in internal/app/scheduler/refresh.go
- [X] T074 [US7] Handle disabled→enabled transition in refresh diff — AddJob for newly-enabled jobs in internal/app/scheduler/refresh.go
- [X] T075 [US7] Default Enabled=true when label is absent (existing behavior preserved) in internal/domain/jobs/types.go

**Checkpoint**: Jobs can be toggled via `bosun.job.enabled` label. Disabled jobs are not scheduled; re-enabling restores scheduling on next refresh.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Executor integration, integration tests, stability validation, documentation

### Executor Integration (#170)

- [X] T076 Inject EventEmitter into existing JobExecutor to emit lifecycle events during execution in internal/app/executor/executor.go

### Integration Tests

- [X] T077 [P] Integration test: scheduler registers and fires jobs with real cron (accelerated intervals) in integration/scheduling_test.go
- [X] T078 [P] Integration test: per-stack mutex serializes jobs targeting same stack under concurrent load in integration/concurrency_test.go
- [X] T079 [P] Integration test: global semaphore limits parallelism (N=1 serial, N=3 allows 3) in integration/concurrency_test.go
- [X] T080 [P] Integration test: daemon lifecycle — start, signal handling, graceful shutdown in integration/daemon_test.go
- [X] T081 Integration test: end-to-end with real Docker daemon and testcontainers-go fixtures in integration/scheduling_test.go

### Stability & Quality

- [X] T082 [P] Run race detector on scheduler tests: go test -race ./internal/app/scheduler/...
- [X] T083 [P] Run race detector on concurrency tests: go test -race ./internal/app/concurrency/...
- [X] T084 Run golangci-lint on all new packages (scheduler, concurrency, events, state, cmd)
- [X] T085 Verify SC-011: grep -r "internal/adapters" internal/cmd/ returns 0 results (no adapter imports in CLI)

### Documentation

- [X] T086 [P] Update docs/config.md with new schedule, overlap-policy, enabled labels and daemon flags
- [X] T087 Validate quickstart.md code snippets compile against implemented interfaces

**Checkpoint**: All integration tests pass with race detector. Code quality validated. Documentation updated. M4 complete.

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup ────────────────────────────┐
                                            ↓
Phase 2: Foundational ─────────────────────┐  (Domain, Ports, Adapters, Bootstrap)
                                            ↓  ⚠️ BLOCKS ALL USER STORIES
                                            ↓
Phase 3: US1 (Scheduling Core) ──────────> MVP checkpoint
                                            ↓
               ┌────────────────────────────┤
               ↓                            ↓
Phase 4: US2 (Overlap) ──┐   Phase 5: US5 (Concurrency) ──┐
                           ↓                                 ↓
                           └───────────┬─────────────────────┘
                                       ↓
                          Phase 6: US3 (Daemon)
                                       ↓
                          Phase 7: US4 (Config Refresh)
                                       ↓
               ┌───────────────────────┤
               ↓                       ↓
Phase 8: US6 (Job List)   Phase 9: US7 (Enable/Disable)
               ↓                       ↓
               └───────────┬───────────┘
                           ↓
                  Phase 10: Polish ──────> ✅ M4 DONE
```

### User Story Dependencies

| Story | Depends On | Can Parallel With |
|-------|-----------|-------------------|
| US1 (P1) | Foundational | — (must be first) |
| US2 (P1) | US1 | US5 |
| US5 (P1) | US1 | US2 |
| US3 (P1) | US1, US2, US5 | — |
| US4 (P1) | US3 | — |
| US6 (P2) | US1 | US7 |
| US7 (P2) | US4 | US6 |

### Within Each User Story

1. Tests written FIRST — must FAIL before implementation (Constitution Principle III)
2. Core struct/types before methods
3. Methods before integration points
4. Story checkpoint validates independent functionality

### Parallel Opportunities

| Phase | Parallel Tasks | Notes |
|-------|---------------|-------|
| Phase 1 | T002, T003 | Different files |
| Phase 2 | T004–T010 (all 7) | All different files, no interdependencies |
| Phase 2 | T012–T015 (all 4) | Test files, different packages |
| Phase 3 | T016–T018 (3 tests) | Same test file but independent test functions |
| Phase 4 | T028–T030 (3 tests) | Independent test functions |
| Phase 5 | T035–T038 (4 tests) | Two different test files |
| Phase 4 + 5 | Entire phases | Different packages (scheduler vs concurrency) |
| Phase 8 + 9 | Entire phases | Different concerns, no shared code |
| Phase 10 | T077–T080, T082–T083 | Independent test files |

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together (test-first):
Task: T016 "Unit test Scheduler.AddJob" in internal/app/scheduler/scheduler_test.go
Task: T017 "Unit test Scheduler.executeJob" in internal/app/scheduler/scheduler_test.go
Task: T018 "Unit test circuit-breaker" in internal/app/scheduler/scheduler_test.go

# Then implement sequentially (same file, interdependent):
Task: T019 "Scheduler struct with New()" → T020 "AddJob" → T021 "RemoveJob"
Task: T022 "executeJob" → T023 "circuit-breaker" → T024 "SaveJobState"
Task: T025 "Start" → T026 "Stop" → T027 "ListJobs"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks everything)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Jobs scheduled via cron, executed through JobExecutor, circuit-breaker works
5. Deploy/demo if ready — core scheduling is functional

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. **US1** → Core scheduling → **MVP!** (jobs run on cron schedule)
3. **US2** → Overlap policies → *safer execution* (queue/skip prevent conflicts)
4. **US5** → Concurrency controls → *production-grade* (three-layer model from #108)
5. **US3** → Daemon command → *deployable as systemd service*
6. **US4** → Config refresh → *dynamic reconfiguration without restart*
7. **US6 + US7** → Job list + enable/disable → *full M4 feature set*
8. **Polish** → Integration tests, race detector, docs → **M4 complete**

Each story adds value without breaking previous stories.

### Success Criteria Coverage

| SC | Covered By |
|----|-----------|
| SC-001 (72h stable daemon) | T080, T081 |
| SC-002 (<10s jitter) | T077 |
| SC-003 (5min config refresh) | T053–T054, T081 |
| SC-004 (60s graceful shutdown) | T044, T048 |
| SC-005 (per-stack mutex 100%) | T035, T078 |
| SC-006 (global semaphore) | T038, T079 |
| SC-007 (<1s job list) | T063–T064 |
| SC-008 (overlap policies) | T028–T030, T077 |
| SC-009 (<50MB memory growth) | T081 |
| SC-010 (integration tests + race) | T077–T083 |
| SC-011 (no adapter imports in cmd/) | T085 |
| SC-012 (Bootstrap resolves #141) | T011, T015 |
| SC-013 (port interfaces defined) | T007, T008 |
| SC-014 (extensible events) | T009, T013, T076 |

---

## Notes

- **[P]** tasks = different files, no dependencies on incomplete tasks
- **[Story]** label maps task to specific user story for traceability
- Each user story is independently completable and testable at its checkpoint
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- US8 (cancel-and-restart) deferred to [#176](https://github.com/simone-viozzi/bosun/issues/176)
- Persistent scheduling adapter deferred to [#177](https://github.com/simone-viozzi/bosun/issues/177) — only the `JobStateStore` port is in M4
