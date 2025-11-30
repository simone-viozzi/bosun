# Tasks: Config Integration Tests

**Input**: Design documents from `/specs/005-config-integration-tests/`
**Prerequisites**: plan.md ✅, spec.md ✅, checklists/requirements.md ✅

**Status**: ✅ All tasks complete - Feature ready for merge

---

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)

---

## Phase 1: Setup (Shared Infrastructure) ✅ COMPLETE

**Purpose**: Verify existing test infrastructure can be reused

- [x] T001 Verify test harness exists in `internal/testutil/harness.go`
- [x] T002 Verify Docker utilities exist in `internal/testutil/docker.go`
- [x] T003 [P] Verify compose file embedding works in `internal/testutil/compose/`

**Checkpoint**: Existing infrastructure confirmed reusable ✅

---

## Phase 2: Foundational (Blocking Prerequisites) ✅ COMPLETE

**Purpose**: Ensure loader and merger are implemented before integration tests

- [x] T004 Verify loader (#58) is implemented in `internal/config/loader/loader.go`
- [x] T005 [P] Verify merger (#59) is implemented in `internal/config/merge/merge.go`
- [x] T006 [P] Verify CLI validate command exists in `internal/cmd/validate.go`
- [x] T007 Create `validate-valid.yaml` compose file in `internal/testutil/compose/`
- [x] T008 Create `validate-invalid.yaml` compose file in `internal/testutil/compose/`

**Checkpoint**: Foundation ready ✅

---

## Phase 3: User Story 1 - Verify Typed Settings Happy Path (P1) 🎯

**Goal**: Prove all 6 config types parse correctly from Docker labels

**Independent Test**: Run `make it` and verify `Test_Integration_ConfigValidate_ValidConfig` passes

### Implementation for User Story 1

- [x] T009 [US1] Add string type label `bosun.instance` to `validate-valid.yaml`
- [x] T010 [P] [US1] Add bool type labels `bosun.container.autoRestart`, `bosun.volume.backupEnabled` to `validate-valid.yaml`
- [x] T011 [P] [US1] Add int type label `bosun.network.priority` to `validate-valid.yaml`
- [x] T012 [P] [US1] Add duration type labels `bosun.container.stopGracePeriod`, `healthCheckInterval` to `validate-valid.yaml`
- [x] T013 [P] [US1] Add size type label `bosun.volume.maxSize` to `validate-valid.yaml`
- [x] T014 [P] [US1] Add enum type label `bosun.container.logLevel` to `validate-valid.yaml`
- [x] T015 [US1] Implement `Test_Integration_ConfigValidate_ValidConfig` in `integration/validate_test.go`

**Checkpoint**: US1 complete - all 6 types verified ✅

---

## Phase 4: User Story 2 - Verify Unknown Key Rejection (P1)

**Goal**: Prove unknown `bosun.*` keys cause hard failures

**Independent Test**: Run `make it` and verify `Test_Integration_ConfigValidate_InvalidConfig` fails with "unknown key"

### Implementation for User Story 2

- [x] T016 [US2] Add typo label `bosun.container.stopGracPeriod` to `validate-invalid.yaml`
- [x] T017 [US2] Implement `Test_Integration_ConfigValidate_InvalidConfig` in `integration/validate_test.go`
- [x] T018 [US2] Add assertion for "unknown key" in stderr output

**Checkpoint**: US2 complete - unknown keys rejected ✅

---

## Phase 5: User Story 3 - Verify Scope Validation (P2)

**Goal**: Prove scope mismatches are rejected (container label on volume)

**Independent Test**: Run `make it` and verify `Test_Integration_ConfigValidate_ScopeFlag` filters correctly

### Implementation for User Story 3

- [x] T019 [US3] Add container-scoped label to volume in `validate-invalid.yaml`
- [x] T020 [US3] Implement `Test_Integration_ConfigValidate_ScopeFlag` in `integration/validate_test.go`
- [x] T021 [US3] Add assertion that volume errors don't appear when `--scope container`

**Checkpoint**: US3 complete - scope validation works ✅

---

## Phase 6: User Story 4 - Verify Type Validation Errors (P2)

**Goal**: Prove invalid type values are rejected with helpful messages

**Independent Test**: Run `make it` and verify invalid values produce parse errors

### Implementation for User Story 4

- [x] T022 [US4] Add invalid duration `healthCheckInterval=notaduration` to `validate-invalid.yaml`
- [x] T023 [P] [US4] Add invalid enum `logLevel=verbose` to `validate-invalid.yaml`
- [x] T024 [P] [US4] Add invalid bool `autoRestart=maybe` to `validate-invalid.yaml`
- [x] T025 [US4] Verify `Test_Integration_ConfigValidate_InvalidConfig` catches all type errors

**Checkpoint**: US4 complete - type validation works ✅

---

## Phase 7: User Story 5 - Verify Config Merge End-to-End (P2)

**Goal**: Prove full pipeline (discover → parse → merge) works correctly

**Independent Test**: Run `make it` and verify `Test_Integration_ConfigValidate_PrintFlag` outputs valid JSON

### Implementation for User Story 5

- [x] T026 [US5] Implement `Test_Integration_ConfigValidate_PrintFlag` in `integration/validate_test.go`
- [x] T027 [US5] Add JSON unmarshaling assertion for merged config
- [x] T028 [US5] Verify `StopGracePeriod` field exists in output

**Checkpoint**: US5 complete - full pipeline works ✅

---

## Phase 8: Polish & Cross-Cutting Concerns ✅ COMPLETE

**Purpose**: Documentation, cleanup, follow-up tracking

- [x] T029 [P] Update spec.md with clarifications section
- [x] T030 [P] Update checklists/requirements.md with implementation evidence
- [x] T031 [P] Create plan.md with implementation summary
- [x] T032 Create tasks.md (this file)
- [ ] T033 Create GitHub issue for multi-container testing (deferred feature)

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)          → No dependencies
Phase 2 (Foundational)   → Depends on Phase 1
Phase 3-7 (User Stories) → Depend on Phase 2
Phase 8 (Polish)         → Depends on all user stories
```

### User Story Independence

All user stories share compose files but test different aspects:
- **US1**: Happy path with valid labels → `validate-valid.yaml`
- **US2**: Unknown key rejection → `validate-invalid.yaml`
- **US3**: Scope validation → `validate-invalid.yaml` + `--scope` flag
- **US4**: Type parse errors → `validate-invalid.yaml`
- **US5**: Full pipeline → `validate-valid.yaml` + `--print` flag

### Parallel Opportunities

Tasks marked [P] within each phase can run in parallel:
- T009-T014 (type labels) can all be added simultaneously
- T022-T024 (invalid values) can all be added simultaneously
- T029-T031 (documentation) can all be updated simultaneously

---

## Verification Commands

```bash
# Run all integration tests
make it

# Run with verbose output
make itv

# Run specific test
go test -tags=integration -run Test_Integration_ConfigValidate_ValidConfig ./integration/...

# Check test coverage
make test && make it
```

---

## Summary

| Phase | Tasks | Status |
|-------|-------|--------|
| Setup | T001-T003 | ✅ Complete |
| Foundational | T004-T008 | ✅ Complete |
| US1 (P1) | T009-T015 | ✅ Complete |
| US2 (P1) | T016-T018 | ✅ Complete |
| US3 (P2) | T019-T021 | ✅ Complete |
| US4 (P2) | T022-T025 | ✅ Complete |
| US5 (P2) | T026-T028 | ✅ Complete |
| Polish | T029-T033 | 🔶 1 pending |

**Total**: 33 tasks, 32 complete, 1 pending (follow-up issue)
