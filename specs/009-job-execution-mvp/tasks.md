# Tasks: Job Execution MVP (Milestone 3)

**Input**: Design documents from `/specs/009-job-execution-mvp/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Integration tests included as requested in spec (#123)

**Organization**: Tasks grouped by user story priority (P1, P2) for independent implementation

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- File paths relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Domain types and error definitions needed by all components

- [ ] T001 Create domain run types in internal/domain/jobs/run.go (JobRun, RunStatus, ExecutionResult, StepResult)
- [ ] T002 [P] Create domain error types in internal/domain/jobs/errors.go (StopError, StartError, TimeoutError, WorkerError)
- [ ] T003 [P] Add M3 exit codes to internal/cmd/exitcodes.go (ExitWorkerFailed, ExitStopFailed, ExitStartFailed, ExitTimeout, ExitImageNotFound)
- [ ] T004 [P] Add new timeout labels to internal/config/schema/job_labels.go (bosun.backup.stop-timeout, bosun.backup.start-timeout)
- [ ] T005 [P] Add defaults constants in internal/domain/jobs/defaults.go (DefaultStopTimeout, DefaultStartTimeout, GracePeriod)

---

## Phase 2: Port Interfaces (Parallel - No Dependencies)

**Purpose**: Define contracts that adapters must implement

**⚠️ CRITICAL**: Ports must be complete before adapters can start

- [ ] T006 [P] Define ComposeController port interface in internal/ports/compose.go (#115)
- [ ] T007 [P] Define WorkerRunner port interface in internal/ports/worker.go (#116)
- [ ] T008 [P] Define JobExecutor port interface in internal/ports/executor.go (#114)

**Checkpoint**: All port interfaces defined - adapter implementation can begin

---

## Phase 3: User Story 1 - Execute a Backup Job (Priority: P1) 🎯 MVP

**Goal**: Run `bosun job run <job-name>` to stop stack, run worker, restart stack

**Independent Test**: `bosun job run daily-backup` against test Compose stack with worker that touches file in volume

### Adapters for User Story 1

- [ ] T009 [US1] Create internal/adapters/docker/compose/doc.go with package documentation
- [ ] T010 [US1] Implement topological sort in internal/adapters/docker/compose/topology.go
- [ ] T011 [US1] Implement ComposeController adapter in internal/adapters/docker/compose/controller.go (#118)
  - ListStackContainers using com.docker.compose.project label filter
  - StopStack with reverse dependency order
  - StartStack with forward dependency order
  - IsStackRunning checking all container states
- [ ] T012 [US1] Unit tests for ComposeController in internal/adapters/docker/compose/controller_test.go
- [ ] T013 [P] [US1] Create internal/adapters/docker/worker/doc.go with package documentation
- [ ] T014 [US1] Implement WorkerRunner adapter in internal/adapters/docker/worker/runner.go (#119)
  - Container creation with image, env, mounts
  - Log streaming via Docker ContainerLogs
  - Timeout enforcement (SIGTERM → 10s → SIGKILL)
  - Container cleanup (remove unless KeepOnFailure)
- [ ] T015 [US1] Unit tests for WorkerRunner in internal/adapters/docker/worker/runner_test.go

### Application Layer for User Story 1

- [ ] T016 [US1] Create internal/app/executor/doc.go with package documentation
- [ ] T017 [US1] Implement Executor service in internal/app/executor/executor.go (#121)
  - Constructor with dependency injection (JobDiscoverer, JobPlanner, ComposeController, WorkerRunner)
  - Execute flow: discover → validate image → stop → run → start
  - Guaranteed restart via defer (even on worker failure)
- [ ] T017.1 [US1] Implement signal handler for graceful Ctrl+C (FR-024)
  - Set up signal.Notify for SIGINT/SIGTERM in cmd/job_run.go
  - Create cancellable context, pass to Executor.Execute()
  - On signal: cancel context → Executor aborts worker → restart stack → exit 16 (ExitInterrupted)
  - Ensure stack restart happens even if worker is killed mid-execution
- [ ] T018 [US1] Unit tests for Executor in internal/app/executor/executor_test.go

### CLI for User Story 1

- [ ] T019 [US1] Implement `bosun job run` command in internal/cmd/job_run.go (#122)
  - Cobra command with `run <job-name>` argument
  - Wire up Executor with Docker client and adapters
  - Exit code mapping from errors
  - Text output for execution progress

**Checkpoint**: User Story 1 (P1) complete - can execute backup jobs via CLI

---

## Phase 4: User Story 2 - Preview Job Execution (Dry Run) (Priority: P1) 🎯 MVP

**Goal**: `bosun job run <job-name> --dry-run` shows plan without execution

**Independent Test**: Run with `--dry-run`, verify no containers stopped, output shows steps

### Implementation for User Story 2

- [ ] T020 [US2] Add DryRun method to Executor in internal/app/executor/executor.go
- [ ] T021 [US2] Add --dry-run flag to job run command in internal/cmd/job_run.go
- [ ] T022 [US2] Add --format flag (text, json) to job run command in internal/cmd/job_run.go
- [ ] T023 [US2] Implement JSON output formatter for dry-run results in internal/cmd/job_run.go
- [ ] T024 [US2] Unit test for dry-run mode in internal/app/executor/executor_test.go

**Checkpoint**: User Stories 1 & 2 (P1) complete - core MVP functionality ready

---

## Phase 5: User Story 3 - Capture Worker Logs (Priority: P2)

**Goal**: See worker stdout/stderr during and after job execution

**Independent Test**: Worker outputs to stdout, logs visible in CLI

### Implementation for User Story 3

- [ ] T025 [US3] Add log streaming to WorkerRunner.Run in internal/adapters/docker/worker/runner.go
- [ ] T026 [US3] Add LogWriter field to ExecuteOptions for real-time output in internal/ports/executor.go
- [ ] T027 [US3] Wire log streaming from Executor to CLI stdout in internal/cmd/job_run.go
- [ ] T028 [US3] Add --quiet flag to suppress logs in internal/cmd/job_run.go
- [ ] T029 [US3] Unit test for log capture in internal/adapters/docker/worker/runner_test.go

**Checkpoint**: User Story 3 (P2) complete - worker logs visible

---

## Phase 6: User Story 4 - Handle Execution Timeouts (Priority: P2)

**Goal**: Hung workers/operations terminated after configurable timeout

**Independent Test**: Worker sleeps longer than timeout, gets terminated, stack restarted

### Implementation for User Story 4

- [ ] T030 [US4] Implement timeout context in ComposeController.StopStack in internal/adapters/docker/compose/controller.go
- [ ] T031 [US4] Implement timeout context in ComposeController.StartStack in internal/adapters/docker/compose/controller.go
- [ ] T032 [US4] Add --timeout flag to override worker timeout in internal/cmd/job_run.go
- [ ] T033 [US4] Add --stop-timeout flag to override stop timeout in internal/cmd/job_run.go
- [ ] T034 [US4] Add --start-timeout flag to override start timeout in internal/cmd/job_run.go
- [ ] T035 [US4] Unit test for timeout handling in internal/adapters/docker/worker/runner_test.go

**Checkpoint**: User Story 4 (P2) complete - timeouts enforced

---

## Phase 7: User Story 5 - Graceful Stack Stop/Start (Priority: P2)

**Goal**: Stack stop/start respects dependency order with graceful signals

**Independent Test**: Stack with depends_on, verify containers stop/start in correct order

### Implementation for User Story 5

- [ ] T036 [US5] Parse com.docker.compose.depends_on label in internal/adapters/docker/compose/controller.go
- [ ] T037 [US5] Validate topological sort handles cycles with error in internal/adapters/docker/compose/topology.go
- [ ] T038 [US5] Add cycle detection test in internal/adapters/docker/compose/topology_test.go
- [ ] T039 [US5] Integration test for dependency ordering in integration/job_execution_test.go

**Checkpoint**: User Story 5 (P2) complete - dependency ordering works

---

## Phase 8: Integration Testing & Documentation

**Purpose**: End-to-end tests with real Docker and docs update

### Integration Tests (#123)

- [ ] T040 Create test harness for job execution in internal/testutil/job_harness.go
- [ ] T041 [P] Create test Compose stack fixture in integration/testdata/simple-stack/docker-compose.yml
- [ ] T042 [P] Create test worker image Dockerfile in integration/testdata/test-worker/Dockerfile
- [ ] T043 Integration test: happy path job execution in integration/job_execution_test.go
- [ ] T044 Integration test: worker failure with stack restart in integration/job_execution_test.go
- [ ] T045 Integration test: timeout termination in integration/job_execution_test.go
- [ ] T046 Integration test: dry-run no side effects in integration/job_execution_test.go
- [ ] T047 Integration test: Ctrl+C graceful shutdown in integration/job_execution_test.go

### Documentation (#120)

- [ ] T048 [P] Update README.md with job run command usage
- [ ] T049 [P] Add M3 example to docs/ with sample Compose stack and worker
- [ ] T050 Validate quickstart.md instructions by following them

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup and quality checks

- [ ] T051 [P] Run golangci-lint and fix any issues
- [ ] T052 [P] Add --keep-stopped flag to skip restart after worker in internal/cmd/job_run.go
- [ ] T053 [P] Add --keep-failed flag to preserve worker container on failure in internal/cmd/job_run.go
- [ ] T054 Verify all exit codes are properly mapped in internal/cmd/job_run.go
- [ ] T055 Run full integration test suite with `make it`
- [ ] T056 Update CHANGELOG.md with M3 release notes

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup) ─────────────────────────────────────────────┐
                                                              │
Phase 2 (Ports) ─────────────────────────────────────────────┤
                                                              │
                     ┌────────────────────────────────────────┘
                     ▼
Phase 3 (US1 - Execute Job) ──────────────────────────────────┐
                     │                                         │
                     ▼                                         │
Phase 4 (US2 - Dry Run) ──────────────────────────────────────┤
                     │                                         │  MVP Complete
                     ├─────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┬────────────┐
        ▼            ▼            ▼
Phase 5 (US3)   Phase 6 (US4)   Phase 7 (US5)  [Can Parallel]
        │            │            │
        └────────────┼────────────┘
                     ▼
Phase 8 (Integration Tests) ──────────────────────────────────
                     │
                     ▼
Phase 9 (Polish) ─────────────────────────────────────────────
```

### Issue to Task Mapping

| GitHub Issue | Tasks | Component |
|--------------|-------|-----------|
| #114 | T008 | JobExecutor port |
| #115 | T006 | ComposeController port |
| #116 | T007 | WorkerRunner port |
| #118 | T009-T012 | ComposeController adapter |
| #119 | T013-T015, T025, T029, T035 | WorkerRunner adapter |
| #121 | T016-T018, T020, T024 | Executor service |
| #122 | T019, T021-T023, T027-T028, T032-T034, T052-T54 | CLI command |
| #123 | T040-T047 | Integration tests |
| #120 | T048-T050 | Documentation |

### Parallel Opportunities

**Phase 1** (all parallel):
- T002, T003, T004, T005 can run simultaneously

**Phase 2** (all parallel):
- T006, T007, T008 can run simultaneously

**Phase 3** (limited parallel):
- T009-T012 (compose adapter) sequential
- T013-T015 (worker adapter) can parallel with compose after T013

**Phases 5-7** (fully parallel):
- US3, US4, US5 implementation can all proceed in parallel after Phase 4

**Phase 8** (partial parallel):
- T041, T042 (fixtures) can parallel
- T043-T047 (tests) sequential (shared Docker state)

---

## MVP Scope

**Minimum Viable Product**: Phases 1-4 (Tasks T001-T024)

After MVP:
- User can run `bosun job run daily-backup` to execute a job
- User can preview with `bosun job run daily-backup --dry-run`
- Output available in text and JSON formats

**Post-MVP** (Phases 5-9): Logs, timeouts, dependency ordering, integration tests, docs

---

## Task Count Summary

| Phase | Tasks | User Story |
|-------|-------|------------|
| Setup | 5 | - |
| Ports | 3 | - |
| US1 Execute Job | 11 | P1 🎯 |
| US2 Dry Run | 5 | P1 🎯 |
| US3 Logs | 5 | P2 |
| US4 Timeouts | 6 | P2 |
| US5 Dependencies | 4 | P2 |
| Integration | 11 | - |
| Polish | 6 | - |
| **Total** | **56** | |

**MVP Tasks**: 24 (Phases 1-4)
**Post-MVP Tasks**: 32 (Phases 5-9)
