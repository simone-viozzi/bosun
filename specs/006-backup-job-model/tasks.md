# Tasks: Job Model & Planning

**Input**: Design documents from `/specs/006-backup-job-model/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Unit tests included for domain and planner. Integration tests included per testing_structure memory.

**Organization**: Tasks grouped by user story (P1, P2) to enable independent implementation and testing.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3, US4)
- Exact file paths included in descriptions

## Path Conventions

Following hexagonal architecture from plan.md:
- Domain: `internal/domain/jobs/`
- Ports: `internal/ports/`
- Adapters: `internal/adapters/joblabels/`
- Application: `internal/app/planner/`
- CLI: `internal/cmd/`
- Config: `internal/config/schema/`
- Integration tests: `integration/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add dependency and create package scaffolding

- [x] T001 Add `github.com/robfig/cron/v3` dependency via `go get github.com/robfig/cron/v3`
- [x] T002 [P] Create `internal/domain/jobs/` package directory with `doc.go`
- [x] T003 [P] Create `internal/adapters/joblabels/` package directory with `doc.go`
- [x] T004 [P] Create `internal/app/planner/` package directory with `doc.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types and port interfaces that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Define `Job` struct in `internal/domain/jobs/types.go` (fields: Name, Schedule, TargetContainers, TargetStacks, WorkerImage, AttachVolumes, SourceContainers)
- [x] T006 Define `VolumeAttachment` struct in `internal/domain/jobs/types.go` (fields: Name, MountPath, Mode)
- [x] T007 Define `ExecutionPlan` struct in `internal/domain/jobs/types.go` (fields: JobName, Steps, CreatedAt)
- [x] T008 Define `PlanStep` struct and `StepType` constants in `internal/domain/jobs/types.go` (StepTypeStopContainers, StepTypeRunWorker, StepTypeStartContainers)
- [x] T009 Define `Stack` struct in `internal/domain/jobs/types.go` (fields: Name, Containers, Volumes, Networks)
- [x] T010 Define `ContainerDependency` struct in `internal/domain/jobs/types.go` for parsing `com.docker.compose.depends_on`
- [x] T011 Add JSON struct tags to all domain types in `internal/domain/jobs/types.go` for serialization
- [x] T012 [P] Define `JobDiscoverer` interface in `internal/ports/planner.go` with `DiscoverJobs(ctx, snapshot) ([]Job, []ValidationError, error)`
- [x] T013 [P] Define `JobPlanner` interface in `internal/ports/planner.go` with `Plan(ctx, job) (ExecutionPlan, error)`
- [x] T014 [P] Define `ValidationError` struct in `internal/ports/planner.go` (EntityKind, EntityID, EntityName, Field, Message)
- [x] T015 [P] Define error sentinels in `internal/ports/planner.go` (ErrOrphanedDependents, ErrInvalidSchedule, ErrMissingJobName, ErrConflictingJobField)
- [x] T016 Write unit tests for domain type instantiation in `internal/domain/jobs/types_test.go`

**Checkpoint**: Foundation ready - domain types and interfaces defined

---

## Phase 3: User Story 1 - Define Jobs via Labels (Priority: P1) 🎯 MVP

**Goal**: Users can define jobs using `bosun.job.*` labels on containers and volumes; Bosun discovers and parses them correctly

**Independent Test**: Deploy Docker Compose stack with job labels → run discoverer → verify job definitions

### Implementation for User Story 1

- [x] T017 [US1] Define job label schema constants in `internal/config/schema/job_labels.go` (bosun.job.enabled, bosun.job.name, bosun.job.schedule, bosun.job.worker.image, bosun.job.attach)
- [x] T018 [US1] Implement `JobLabelConfig` and `JobVolumeConfig` structs with bosun tags in `internal/config/schema/job_labels.go`
- [x] T019 [US1] Add `JobSpec()` function in `internal/config/schema/job_labels.go` to return parsed `Spec` for job labels
- [x] T020 [US1] Write unit tests for job label schema parsing in `internal/config/schema/job_labels_test.go`
- [x] T021 [US1] Implement `NewDiscoverer()` factory in `internal/adapters/joblabels/discoverer.go`
- [x] T022 [US1] Implement `DiscoverJobs()` method: filter containers with `bosun.job.enabled=true` in `internal/adapters/joblabels/discoverer.go`
- [x] T023 [US1] Implement job merging by `bosun.job.name` (multiple containers → one job) in `internal/adapters/joblabels/discoverer.go`
- [x] T024 [US1] Implement cron expression validation using `robfig/cron/v3` parser in `internal/adapters/joblabels/discoverer.go`
- [x] T025 [US1] Implement default value application (schedule, worker image) in `internal/adapters/joblabels/discoverer.go`
- [x] T026 [US1] Implement volume attachment discovery from `bosun.job.attach` labels in `internal/adapters/joblabels/discoverer.go`
- [x] T027 [US1] Implement volume mode parsing (`ro`/`rw`, default `ro`) in `internal/adapters/joblabels/discoverer.go`
- [x] T028 [US1] Implement stack resolution (bosun.stack > com.docker.compose.project) in `internal/adapters/joblabels/discoverer.go`
- [x] T029 [US1] Implement validation error collection (type errors, missing name, conflicting values) in `internal/adapters/joblabels/discoverer.go`
- [x] T030 [US1] Write unit tests for discoverer with mocked snapshot in `internal/adapters/joblabels/discoverer_test.go`
- [x] T031 [US1] Create test Docker Compose file with job labels in `internal/testutil/compose/joblabels-compose.yaml`
- [x] T032 [US1] Write integration test for job discovery in `integration/joblabels_test.go`

**Checkpoint**: User Story 1 complete - jobs can be defined via labels and discovered

---

## Phase 4: User Story 2 - View Discovered Jobs (Priority: P1) 🎯 MVP

**Goal**: Users can run `bosun plan list` to see all discovered jobs

**Independent Test**: Run `bosun plan list` → verify output shows jobs in text/json/yaml formats

### Implementation for User Story 2

- [x] T033 [US2] Create `plan` command group in `internal/cmd/plan.go` with `NewPlanCmd()`
- [x] T034 [US2] Register `plan` command in `internal/cmd/root.go`
- [x] T035 [US2] Implement `bosun plan list` command in `internal/cmd/plan_list.go`
- [x] T036 [US2] Add `--format` flag (text/json/yaml) to plan list in `internal/cmd/plan_list.go`
- [x] T037 [US2] Add `--stopped` flag to include stopped containers in `internal/cmd/plan_list.go`
- [x] T038 [US2] Add `--stack` filter flag in `internal/cmd/plan_list.go`
- [x] T039 [US2] Implement text output renderer for job list (table format) in `internal/cmd/plan_list.go`
- [x] T040 [US2] Implement JSON output renderer for job list in `internal/cmd/plan_list.go`
- [x] T041 [US2] Implement YAML output renderer for job list in `internal/cmd/plan_list.go`
- [x] T042 [US2] Implement "no jobs discovered" friendly message in `internal/cmd/plan_list.go`
- [x] T043 [US2] Handle Docker unavailable error with exit code 2 in `internal/cmd/plan_list.go`
- [x] T044 [US2] Handle validation errors with exit code 1 in `internal/cmd/plan_list.go`
- [x] T045 [US2] Write integration test for `bosun plan list` command in `integration/plan_test.go`

**Checkpoint**: User Story 2 complete - `bosun plan list` works with all formats

---

## Phase 5: User Story 3 - Preview Execution Plan (Priority: P1) 🎯 MVP

**Goal**: Users can run `bosun plan show <job>` to see the exact steps Bosun would take

**Independent Test**: Run `bosun plan show <job>` → verify output shows ordered steps (stop, run-worker)

### Implementation for User Story 3

- [X] T046 [US3] Implement `NewPlanner()` factory in `internal/app/planner/planner.go`
- [X] T047 [US3] Implement `Plan()` method: generate stop_containers step in `internal/app/planner/planner.go`
- [X] T048 [US3] Implement run_worker step generation with volume mounts in `internal/app/planner/planner.go`
- [X] T049 [US3] Implement deterministic sorting (containers by ID, volumes by name) in `internal/app/planner/planner.go`
- [X] T050 [US3] Implement Compose dependency parsing from `com.docker.compose.depends_on` label in `internal/app/planner/planner.go`
- [X] T051 [US3] Implement orphan dependent validation (fail if stopping would orphan dependents) in `internal/app/planner/planner.go`
- [X] T052 [US3] Implement `useComposeStop` detection (all containers in stack) in `internal/app/planner/planner.go`
- [X] T053 [US3] Implement human-readable step descriptions in `internal/app/planner/planner.go`
- [X] T054 [US3] Write unit tests for planner determinism in `internal/app/planner/planner_test.go`
- [X] T055 [US3] Write unit tests for orphan dependent detection in `internal/app/planner/planner_test.go`
- [X] T056 [US3] Implement `bosun plan show <job-name>` command in `internal/cmd/plan_show.go`
- [X] T057 [US3] Add `--format` flag (text/json/yaml) to plan show in `internal/cmd/plan_show.go`
- [X] T058 [US3] Add `--stopped` flag to include stopped containers in `internal/cmd/plan_show.go`
- [X] T059 [US3] Implement text output renderer for execution plan in `internal/cmd/plan_show.go`
- [X] T060 [US3] Implement JSON output renderer for execution plan in `internal/cmd/plan_show.go`
- [X] T061 [US3] Implement YAML output renderer for execution plan in `internal/cmd/plan_show.go`
- [X] T062 [US3] Implement "job not found" error with available jobs list in `internal/cmd/plan_show.go`
- [X] T063 [US3] Implement orphan dependent error output in `internal/cmd/plan_show.go`
- [X] T064 [US3] Write integration test for `bosun plan show` command in `integration/plan_test.go`

**Checkpoint**: User Story 3 complete - `bosun plan show` generates deterministic plans ✅

---

## Phase 6: User Story 4 - Validate Job Configuration (Priority: P2)

**Goal**: Users can run `bosun config validate` to catch job label errors before running

**Independent Test**: Deploy stack with invalid job labels → run `bosun config validate` → verify error messages

### Implementation for User Story 4

- [X] T065 [US4] Extend existing validation in `internal/cmd/validate.go` to include job labels
- [X] T066 [US4] Add job label validation to loader in `internal/config/loader/loader.go`
- [X] T067 [US4] Implement validation for `bosun.job.enabled` type (boolean) in `internal/config/loader/loader.go`
- [X] T068 [US4] Implement validation for `bosun.job.schedule` (cron expression) in `internal/config/loader/loader.go`
- [X] T069 [US4] Implement validation for missing `bosun.job.name` when enabled=true in `internal/config/loader/loader.go`
- [X] T070 [US4] Implement validation for conflicting job field values in `internal/config/loader/loader.go`
- [X] T071 [US4] Implement warning for orphaned volume attachment (job doesn't exist) in `internal/config/loader/loader.go`
- [X] T072 [US4] Write unit tests for job label validation in `internal/config/loader/loader_test.go`
- [X] T073 [US4] Create test compose file with invalid job labels in `internal/testutil/compose/joblabels-invalid-compose.yaml`
- [X] T074 [US4] Write integration test for job label validation in `integration/validate_test.go`

**Checkpoint**: User Story 4 complete - `bosun config validate` catches job label errors ✅

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, and final validation

- [x] T075 [P] Update `docs/config.md` with job label documentation
- [x] T076 [P] Run `go generate` to update config schema JSON
- [x] T077 [P] Add job labels section to README.md
- [x] T078 Run all unit tests: `make test`
- [x] T079 Run all integration tests: `make it`
- [x] T080 Run linting: `make lint`
- [x] T081 Validate quickstart.md scenarios work end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (Phase 3): Can start after Phase 2
  - US2 (Phase 4): Depends on US1 (needs discoverer)
  - US3 (Phase 5): Depends on US1 (needs jobs), can parallel with US2
  - US4 (Phase 6): Depends on US1 (needs schema), can parallel with US2/US3
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 2 (Foundational)
        ↓
    Phase 3 (US1: Define Jobs)
        ↓
   ┌────┴────┬──────────┐
   ↓         ↓          ↓
Phase 4   Phase 5   Phase 6
(US2)     (US3)     (US4)
   └────┬────┴──────────┘
        ↓
    Phase 7 (Polish)
```

### Within Each Phase

- Tasks marked [P] can run in parallel
- Schema tasks before discoverer tasks
- Discoverer tasks before CLI tasks
- Planner tasks before plan show CLI tasks
- Unit tests alongside implementation
- Integration tests after implementation

### Parallel Opportunities

```bash
# Phase 1: All setup tasks can run in parallel
T002, T003, T004

# Phase 2: Port definitions can parallel domain types
T012, T013, T014, T015 (after T005-T011)

# After US1 complete: US2, US3, US4 can parallel
Phase 4 || Phase 5 || Phase 6

# Phase 7: Documentation tasks can parallel
T075, T076, T077
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T016)
3. Complete Phase 3: US1 - Define Jobs (T017-T032)
4. Complete Phase 4: US2 - View Jobs (T033-T045)
5. Complete Phase 5: US3 - Preview Plans (T046-T064)
6. **STOP and VALIDATE**: MVP is complete with job discovery, listing, and plan preview
7. Optionally continue to Phase 6: US4 - Validate Config (T065-T074)
8. Complete Phase 7: Polish (T075-T081)

### Task Count Summary

| Phase | Story | Tasks | Parallel |
|-------|-------|-------|----------|
| 1: Setup | — | 4 | 3 |
| 2: Foundational | — | 12 | 4 |
| 3: US1 | Define Jobs | 16 | 2 |
| 4: US2 | View Jobs | 13 | 0 |
| 5: US3 | Preview Plans | 19 | 2 |
| 6: US4 | Validate Config | 10 | 0 |
| 7: Polish | — | 7 | 3 |
| **Total** | | **81** | **14** |

---

## Notes

- All file paths are relative to repository root
- [P] = parallelizable (different files, no dependencies)
- [USn] = belongs to User Story n
- Commit after each task or logical group
- Run `make test` frequently during implementation
- Schema changes require `go generate` to update docs
