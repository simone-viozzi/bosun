# Tasks: Milestone 3.5 - Post-M3 Cleanup & Bug Fixes

**Input**: Design documents from `/specs/010-m35-cleanup/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: No new tests explicitly requested. Existing tests will be updated as part of implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5, US6)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No new project setup needed — this is a cleanup milestone modifying existing code.

- [x] T001 Verify branch `010-m35-cleanup` is checked out and up-to-date with main
- [x] T002 Run `make build` to confirm current state compiles
- [x] T003 Run `make test` to establish baseline test status

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core type changes that other tasks depend on.

**⚠️ CRITICAL**: US2 and US4 share dependencies on types in `internal/domain/jobs/`. Complete these first.

- [x] T004 Add `Warnings` field to `ValidationErrors` struct in `internal/config/loader/errors.go`
- [x] T005 Add `LoadOptions` struct with `Strict` field in `internal/config/loader/loader.go`
- [x] T006 Add `AddUnknownKeyWarning` method to `ValidationErrors` in `internal/config/loader/errors.go`

**Checkpoint**: Foundation types ready — user story implementation can begin.

- [x] T006a 🔒 COMMIT: `feat(loader): add ValidationErrors.Warnings and LoadOptions foundation`

---

## Phase 3: User Story 1 - Validate Config Without False Positives (Priority: P0) 🎯 MVP

**Goal**: Config validation succeeds on valid compose files, warns (not errors) on unknown `bosun.*` keys.

**Independent Test**: `bin/bosun config validate --project joblabels` passes with warnings, not errors.

### Implementation for User Story 1

- [x] T007 [US1] Modify `FromLabels` in `internal/config/loader/loader.go` to accept `LoadOptions` parameter
- [x] T008 [US1] Change unknown key handling: call `AddUnknownKeyWarning` instead of `AddUnknownKey` when `!opts.Strict`
- [x] T009 [US1] Add `--strict` flag to `config validate` command in `internal/cmd/validate.go`
- [x] T010 [US1] Update CLI output to display warnings separately from errors in `internal/cmd/validate.go`
- [x] T011 [P] [US1] Fix test compose file: remove `bosun.other` from `internal/testutil/compose/joblabels-compose.yaml:46`
- [x] T012 [P] [US1] Fix test compose file: change `bosun.network` to `bosun.network.priority` in `internal/testutil/compose/joblabels-compose.yaml:68`
- [x] T013 [US1] Update loader unit tests in `internal/config/loader/loader_test.go` for lenient mode
- [x] T014 [US1] Update loader unit tests in `internal/config/loader/loader_test.go` for strict mode

**Checkpoint**: `bosun config validate --project joblabels` passes. US1 independently testable.

- [x] T014a 🔒 COMMIT: `feat(config): lenient unknown-key handling with --strict override (#139)`

---

## Phase 4: User Story 2 - See Complete Execution Plan (Priority: P0)

**Goal**: `bosun plan show` displays ALL execution steps including container restart.

**Independent Test**: `bin/bosun plan show daily-backup --project joblabels --format json` includes `start_containers` step.

### Implementation for User Story 2

- [x] T015 [US2] Add `start_containers` step generation in `internal/app/planner/planner.go` (remove TODO comment at line ~84)
- [x] T016 [US2] Create `generateStartDescription` helper function in `internal/app/planner/planner.go`
- [x] T017 [US2] Update `Plan()` to append `StepTypeStartContainers` step after `StepTypeRunWorker`
- [x] T018 [US2] Update planner unit tests in `internal/app/planner/planner_test.go` to expect 3 steps
- [x] T019 [US2] Update integration test assertions in `integration/plan_test.go` for new step

**Checkpoint**: `bosun plan show <job>` shows stop → worker → start. US2 independently testable.

- [x] T019a 🔒 COMMIT: `feat(planner): add start_containers step to execution plan (#142)`

---

## Phase 5: User Story 3 - Clean Generic Terminology (Priority: P1)

**Goal**: Remove "backup" terminology from production code; use generic "job" terminology.

**Independent Test**: `grep -r "BackupEnabled" internal/ | grep -v _test.go` returns empty.

### Implementation for User Story 3

- [x] T020 [P] [US3] Rename `BackupEnabled` → `Enabled` in `internal/config/schema/config_v1.go:30`
- [x] T021 [P] [US3] Update label key `bosun.volume.backupEnabled` → `bosun.volume.enabled` in `internal/config/schema/config_v1.go`
- [x] T022 [P] [US3] Update doc comments in `internal/config/schema/job_labels.go` (remove "backup" references)
- [x] T023 [P] [US3] Update package doc in `internal/domain/jobs/run.go` (change "backup job execution" → "job execution")
- [x] T024 [P] [US3] Update package doc in `internal/adapters/docker/worker/doc.go` (change "backup worker" → "worker")
- [x] T025 [P] [US3] Update CLI description in `internal/cmd/plan_list.go` (change "backup jobs" → "jobs")
- [x] T026 [P] [US3] Update CLI description in `internal/cmd/plan_show.go` (change "backup job" → "job")
- [x] T027 [P] [US3] Update CLI description in `internal/cmd/job_run.go` (change "backup job" → "job")
- [x] T028 [P] [US3] Update package doc in `internal/ports/doc.go` (change "backup job definitions" → "job definitions")
- [x] T029 [P] [US3] Update documentation in `docs/config.md` for renamed label
- [x] T030 [US3] Update config schema tests in `internal/config/schema/config_v1_test.go` for renamed field
- [x] T031 [US3] Update any golden files or test fixtures referencing `BackupEnabled`

**Checkpoint**: No "backup" terminology in production code. US3 independently testable.

- [x] T031a 🔒 COMMIT: `refactor(schema): rename BackupEnabled to Enabled, remove backup terminology (#140)`

---

## Phase 6: User Story 4 - Execute Jobs via Plan Interpreter (Priority: P1)

**Goal**: Executor interprets execution plan steps instead of hardcoding the flow.

**Independent Test**: Unit tests pass; executor iterates `plan.Steps`.

### Implementation for User Story 4

- [x] T032 [US4] Update `JobExecutor` interface in `internal/ports/executor.go` to use `jobs.Job` parameter
- [x] T033 [US4] Remove unused `discoverer` parameter from `Executor.New()` in `internal/app/executor/executor.go`
- [x] T034 [US4] Rename `ExecuteJob` → `Execute` in `internal/app/executor/executor.go`
- [x] T035 [US4] Remove stub `Execute(ctx, jobName)` method that returns "not implemented"
- [x] T036 [US4] Add `executeStep(ctx, step, opts)` helper method in `internal/app/executor/executor.go`
- [x] T037 [US4] Refactor `Execute` to iterate `plan.Steps` and call `executeStep` for each
- [x] T038 [US4] Implement step handler for `StepTypeStopContainers` in `executeStep`
- [x] T039 [US4] Implement step handler for `StepTypeRunWorker` in `executeStep`
- [x] T040 [US4] Implement step handler for `StepTypeStartContainers` in `executeStep`
- [x] T041 [US4] Update `StepResult` recording to capture per-step outcomes
- [x] T042 [US4] Update CLI callers in `internal/cmd/job_run.go` for new interface signature
- [x] T043 [US4] Update executor unit tests in `internal/app/executor/executor_test.go`
- [x] T044 [US4] Update integration tests in `integration/job_execution_test.go`

**Checkpoint**: Executor interprets plan steps. DryRun matches Execute. US4 independently testable.

- [x] T044a 🔒 COMMIT: `feat(executor): plan-driven step interpreter with JobExecutor interface (#143)`

---

## Phase 7: User Story 5 - Clean CLI Error Output (Priority: P3)

**Goal**: CLI errors appear only once, not duplicated.

**Independent Test**: `bin/bosun invalidcommand 2>&1 | grep -c "unknown command"` returns 1.

### Implementation for User Story 5

- [x] T045 [US5] Identify duplicate error printing source in `internal/cmd/root.go` or `cmd/bosun/main.go`
- [x] T046 [US5] Fix duplicate error handling — ensure Cobra's `SilenceErrors` or error propagation is correct
- [x] T047 [US5] Verify fix with manual testing of invalid commands

**Checkpoint**: Errors appear once. US5 independently testable.

- [x] T047a 🔒 COMMIT: `fix(cli): deduplicate error output (#133)`

---

## Phase 8: User Story 6 - Learn Job Execution from README (Priority: P3)

**Goal**: README documents basic job execution with working examples.

**Independent Test**: Follow README instructions successfully.

### Implementation for User Story 6

- [x] T048 [P] [US6] Add "Running Jobs" section to `README.md` after existing sections
- [x] T049 [P] [US6] Document `bosun job run <job>` usage with flags in `README.md`
- [x] T050 [P] [US6] Document `--dry-run` flag with example output in `README.md`
- [x] T051 [P] [US6] Add example compose file snippet showing job labels in `README.md`
- [x] T052 [US6] Verify README examples work with test compose files

**Checkpoint**: README has working "Running Jobs" section. US6 independently testable.

- [x] T052a 🔒 COMMIT: `docs(readme): add Running Jobs section with examples (#120)` (already complete)

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T053 Run `make build` to verify compilation
- [x] T054 Run `make test` to verify unit tests pass
- [x] T055 Run `make it` to verify integration tests pass
- [x] T056 Run `make lint` to verify linting passes
- [x] T057 Run quickstart.md validation commands to verify all success criteria
- [x] T058 Update `docs/config.md` if any additional label documentation needed (no changes needed)

---

## Phase 10: PR #151 Review Fixes

**Purpose**: Address review comments from PR #151

**Source**: Review comments in `pr151-review.md` and `wip_smell_milestone3.md` (smell #23 NOT FIXED)

### Critical: Executor Step Interpreter (Smell #23)

- [x] T059 [US4] Implement step interpreter pattern in `internal/app/executor/executor.go`:
  - Add `stepExecutionContext` struct to hold shared state across steps
  - Add `executeStep(ctx, step, execCtx)` method with switch on `step.Type`
  - Add `executeStopStep`, `executeWorkerStep`, `executeStartStep` helper methods
  - Refactor `Execute` to iterate `plan.Steps` and call `executeStep` for each
  - Remove hardcoded stop→worker→start sequence

### Planner Logic Cleanup

- [x] T060 [US2] Fix duplicated useCompose logic in `internal/app/planner/planner.go`:
  - Extract `useCompose` decision once before Step 1
  - Reuse for both stop and start steps
  - Update TODO comment to clarify correct compose stop logic

### Comment Fixes (Per Reviewer Suggestions)

- [x] T061 [P] Fix `internal/cmd/job_run.go:156` - Remove `(FR-024)` from signal handler comment
- [x] T062 [P] Add TODO in `internal/cmd/job_run.go` near `printDryRunText` noting layering concern
- [x] T063 [P] Fix `internal/cmd/validate.go:284` - Change comment to "Print config label errors (entity-grouped)"
- [x] T064 [P] Fix `integration/validate_test.go:69` - Simplify comment per reviewer suggestion
- [x] T065 [P] Fix `internal/config/loader/errors.go:94` - Make error messages consistent (remove mode mentions)

### Final Verification

- [x] T066 Update `wip_smell_milestone3.md` to mark smell #23 as FIXED
- [x] T067 Run `make build && make test && make lint` to verify all fixes

**Checkpoint**: All PR review comments addressed. Ready for re-review.

- [x] T067a 🔒 COMMIT: `fix(executor): implement step interpreter pattern, address PR review comments`

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
     ↓
Phase 2 (Foundational) ← BLOCKS all user stories
     ↓
┌────┴────┐
│  US1-6  │ (can proceed in priority order or parallel)
└────┬────┘
     ↓
Phase 9 (Polish)
```

### User Story Dependencies

| Story | Priority | Dependencies | Can Parallel With |
|-------|----------|--------------|-------------------|
| US1 | P0 | Phase 2 | US2 (different files) |
| US2 | P0 | Phase 2 | US1 (different files) |
| US3 | P1 | Phase 2 | US1, US2, US5, US6 (different files) |
| US4 | P1 | Phase 2, US2 (needs start step) | US3, US5, US6 |
| US5 | P3 | Phase 2 | All others (different files) |
| US6 | P3 | Phase 2 | All others (different files) |

### Within Each User Story

- Implementation before test updates
- Core changes before CLI integration
- Unit tests before integration tests

### Parallel Opportunities

```bash
# After Phase 2, launch P0 stories together:
# US1: T007-T014 (config loader)
# US2: T015-T019 (planner)

# US3 tasks are highly parallelizable (all [P]):
# T020-T031 can all run in parallel (different files)

# US6 tasks are highly parallelizable (all [P]):
# T048-T051 can all run in parallel (README sections)
```

---

## Implementation Strategy

### MVP First (P0 Stories Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 (config validation)
4. Complete Phase 4: US2 (execution plan)
5. **STOP and VALIDATE**: Both P0 stories should work
6. Run quickstart.md P0 validation commands

### Incremental Delivery

1. P0 complete → Basic M3.5 functionality working
2. Add US3 (terminology) → Cleaner codebase
3. Add US4 (executor) → Plan-driven execution
4. Add US5 (CLI errors) → Better UX
5. Add US6 (docs) → Onboarding ready

---

## Task Summary

| Phase | Tasks | Parallelizable | Commit |
|-------|-------|----------------|--------|
| Setup | T001-T003 | 0 | — |
| Foundational | T004-T006a | 0 | T006a |
| US1 (P0) | T007-T014a | 2 (T011, T012) | T014a |
| US2 (P0) | T015-T019a | 0 | T019a |
| US3 (P1) | T020-T031a | 10 (T020-T029) | T031a |
| US4 (P1) | T032-T044a | 0 | T044a |
| US5 (P3) | T045-T047a | 0 | T047a |
| US6 (P3) | T048-T052a | 4 (T048-T051) | T052a |
| Polish | T053-T058 | 0 | — |
| **Total** | **65 tasks** | **16 parallelizable** | **7 commits** |

---

## Notes

- Tasks T020-T029 (US3) can all run in parallel — they modify different files
- US4 depends on US2 being complete (needs `start_containers` step in plan)
- Test updates follow implementation in each story
- Breaking changes to `bosun.volume.enabled` label are acceptable (pre-1.0)
