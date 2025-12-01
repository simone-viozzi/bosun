# Tasks: Milestone 2.5 – Polish Job Model, Config Schema & Test Suite

**Feature Branch**: `007-milestone-2-5-polish`
**Created**: 2025-11-30
**Status**: Ready for Implementation

---

## Phase 1: Setup

**Goal**: Establish foundational infrastructure before user story implementation.

- [x] T001 Create exit codes constants in `internal/cmd/exitcodes.go`
- [x] T002 Add `StackFilter` field to `Selector` struct in `internal/ports/labels.go`
- [x] T003 [P] Add unit test for empty filter behavior in `internal/ports/labels_test.go`

**Checkpoint**: ✅ Exit codes defined, Selector extended with StackFilter.

---

## Phase 2: Foundational – Docker API Filtering

**Goal**: Implement Docker API label filtering, blocking prerequisite for all filtering user stories.

- [x] T004 Implement Docker API filtering for `ProjectFilter` in `internal/adapters/dockerlabels/source.go`
- [x] T005 Implement Docker API filtering for `StackFilter` in `internal/adapters/dockerlabels/source.go`
- [x] T006 [P] Add unit tests for Docker API filtering in `internal/adapters/dockerlabels/source_test.go`

**Checkpoint**: ✅ Docker adapter filters containers by project/stack labels server-side.

---

## Phase 3: User Story 1 – Validation Unification (P1)

**Story Goal**: Single source of truth for job validation rules.

**Independent Test**: Run `bosun plan list` and `bosun config validate` on the same compose stack – both report identical validation errors.

### Implementation Tasks

- [x] T007 [US1] Add helper functions to extract label keys from struct tags in `internal/config/schema/job_labels.go`
- [x] T008 [US1] Export `DefaultSchedule()`, `DefaultWorkerImage()`, `DefaultMountMode()` functions in `internal/config/schema/job_labels.go`
- [x] T009 [P] [US1] Add unit tests for schema helper functions in `internal/config/schema/job_labels_test.go`
- [x] T010 [US1] Refactor `internal/adapters/joblabels/discoverer.go` to use schema package constants
- [x] T011 [US1] Refactor `internal/config/loader/job_validation.go` to use schema package constants
- [x] T012 [US1] Implement case-insensitive mount mode validation with lowercase normalization in `internal/config/schema/job_labels.go`
- [x] T013 [P] [US1] Remove duplicate constants from `internal/domain/jobs/types.go` (keep imports to schema)
- [x] T014 [US1] Verify both validation paths produce identical errors for same input

**Checkpoint**: All validation logic unified, mount mode case-insensitive.

---

## Phase 4: User Story 2 – Project/Stack Filtering (P1)

**Story Goal**: Isolate CLI operations to specific Docker Compose projects.

**Independent Test**: Deploy two compose stacks, use `bosun plan list --project stack-a`, verify only stack-a jobs appear.

### Implementation Tasks

- [x] T015 [US2] Add `--project` flag to `bosun plan list` in `internal/cmd/plan_list.go`
- [x] T016 [US2] Add `--stack` flag to `bosun plan list` in `internal/cmd/plan_list.go`
- [x] T017 [P] [US2] Add `--project` and `--stack` flags to `bosun plan show` in `internal/cmd/plan_show.go`
- [x] T018 [P] [US2] Add `--project` and `--stack` flags to `bosun labels snapshot` in `internal/cmd/snapshot.go`
- [x] T019 [US2] Implement "No jobs found" stderr message when filter matches nothing
- [x] T020 [US2] Update integration tests to use `--project` filtering in `integration/joblabels_test.go`
- [x] T021 [P] [US2] Update integration tests to use `--project` filtering in `integration/plan_test.go`

**Checkpoint**: All plan/snapshot commands support project/stack filtering, tests isolated.

---

## Phase 5: User Story 3 – CLI Branding (P2)

**Story Goal**: CLI help describes Bosun as a backup job orchestrator.

**Independent Test**: Run `bosun --help`, verify description mentions backup/orchestration.

### Implementation Tasks

- [x] T022 [US3] Update root command Short/Long descriptions in `internal/cmd/root.go`
- [x] T023 [P] [US3] Update `plan` subcommand description in `internal/cmd/plan.go`
- [x] T024 [P] [US3] Update `config` subcommand description in `internal/cmd/config.go`
- [x] T025 [P] [US3] Update `labels` subcommand description in `internal/cmd/labels.go`

**Checkpoint**: All command help text aligned with README branding.

---

## Phase 6: User Story 4 – Error Messages (P2)

**Story Goal**: Consistent, actionable error messages across all commands.

**Independent Test**: Trigger various errors, verify consistent format with context and suggestions.

### Implementation Tasks

- [x] T026 [US4] Update commands to use `ExitRuntimeError` (1) and `ExitValidationError` (2) from `internal/cmd/exitcodes.go`
- [x] T027 [US4] Improve "job not found" error in `plan show` to suggest `bosun plan list` in `internal/cmd/plan_show.go`
- [x] T028 [P] [US4] Update `config validate` to return exit code 2 for validation failures in `internal/cmd/validate.go`
- [x] T029 [US4] Standardize error message format across commands (Error: context: message)

**Checkpoint**: All exit codes standardized, error messages actionable.

---

## Phase 7: User Story 5 – Test Redundancy (P3)

**Story Goal**: Clear separation between unit and integration test responsibilities.

**Independent Test**: Review test inventory, verify no scenario duplicated between layers.

### Implementation Tasks

- [x] T030 [US5] Create testing philosophy doc in `internal/testutil/doc.go`
- [x] T031 [US5] Review `integration/joblabels_test.go` and remove redundant edge case tests
- [x] T032 [P] [US5] Add header comments to compose fixtures in `internal/testutil/compose/`
- [x] T033 [US5] Review `integration/plan_test.go` and keep only happy path E2E tests

**Checkpoint**: Test responsibilities documented, redundant tests removed.

---

## Phase 8: User Story 6 – Documentation (P3)

**Story Goal**: README and generated docs accurately reflect current behavior.

**Independent Test**: Follow README examples, verify they work as documented.

### Implementation Tasks

- [x] T034 [US6] Run `go generate ./internal/config/schema/...` and verify `docs/config.md` includes job labels
- [x] T035 [P] [US6] Review and update README examples if stale
- [x] T036 [US6] Convert remaining TODOs in `TODO.md` to GitHub issues or mark complete

**Checkpoint**: Docs accurate, TODOs triaged.

---

## Phase 9: Polish & Cross-Cutting

**Goal**: Final verification and cleanup.

- [x] T037 Run `make test` and verify all unit tests pass
- [x] T038 Run `make it` and verify all integration tests pass
- [x] T039 Manual verification: `bosun plan list --project` works end-to-end
- [x] T040 Manual verification: `bosun --help` shows job orchestrator branding

**Checkpoint**: All tests green, manual verification complete.

---

## Dependencies

```
T001 ──────────────────────────────────────────────────▶ T026, T027, T028, T029
T002 ──────────────────────────────────────────────────▶ T004, T005
T004, T005 ────────────────────────────────────────────▶ T015, T016, T017, T018
T007, T008 ────────────────────────────────────────────▶ T010, T011, T013
T015, T016, T017, T018 ────────────────────────────────▶ T020, T021
T030 ──────────────────────────────────────────────────▶ T031, T033
```

**Key Dependency Notes**:
- **Phase 2 blocks Phase 4**: Docker API filtering must work before CLI flags can use it
- **Phase 3 (US1) is independent**: Validation unification doesn't depend on filtering
- **Phase 5, 6 depend on Phase 1**: CLI branding and error messages need exit codes defined
- **Phase 7, 8 are mostly independent**: Documentation and test cleanup can proceed in parallel

---

## Parallel Execution Examples

### Within Phase 3 (US1):
```
T009 ─┬─ T010 ────────▶ T014
      │
T013 ─┘
```

### Within Phase 4 (US2):
```
T015 ────▶ T016 ────▶ T019
T017 ─┬─▶ T020
T018 ─┘   T021
```

### Within Phase 5 (US3):
```
T022 ─┬─▶ All subcommands in parallel
T023 ─┤
T024 ─┤
T025 ─┘
```

### Across Phases:
```
Phase 1 + Phase 3 (US1) can run in parallel
Phase 4 (US2) requires Phase 2 complete
Phase 5 (US3), Phase 6 (US4) require Phase 1 complete
Phase 7 (US5), Phase 8 (US6) can start anytime
```

---

## Implementation Strategy

### MVP Scope
- **Minimum Viable**: Phase 1 + Phase 2 + Phase 4 (US2 - Project Filtering)
  - Exit codes defined
  - Docker filtering works
  - `--project` flag on all commands
  - Integration tests isolated

This MVP unblocks reliable CI and enables all other polish work.

### Incremental Delivery Order
1. **Sprint 1**: Phases 1-2, then US2 (filtering) – unblocks CI
2. **Sprint 2**: US1 (validation) – reduces tech debt
3. **Sprint 3**: US3, US4 (branding, errors) – improves UX
4. **Sprint 4**: US5, US6 (tests, docs) – polish

### Risk Mitigation
- **Docker API changes**: Use well-documented filter API, test with Docker 24+
- **Schema refactor scope creep**: Focus only on label key/default extraction
- **Test removal**: Only remove tests that are 100% covered by unit tests

---

## Summary

| Metric | Value |
|--------|-------|
| **Total Tasks** | 40 |
| **Setup/Foundational** | 6 tasks |
| **US1 (Validation)** | 8 tasks |
| **US2 (Filtering)** | 7 tasks |
| **US3 (Branding)** | 4 tasks |
| **US4 (Errors)** | 4 tasks |
| **US5 (Tests)** | 4 tasks |
| **US6 (Docs)** | 3 tasks |
| **Polish** | 4 tasks |
| **Parallel Opportunities** | 16 tasks marked [P] |

### Independent Test Criteria per Story
- **US1**: Both validation paths produce identical errors
- **US2**: `--project` filter returns only matching jobs
- **US3**: Help text mentions backup/orchestration
- **US4**: Errors include context and suggestions
- **US5**: No duplicated test scenarios
- **US6**: README examples work

### Format Validation
✅ All 40 tasks follow checklist format: `- [ ] [TaskID] [P?] [Story?] Description with file path`
