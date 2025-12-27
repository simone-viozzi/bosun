# Tasks: Strict golangci-lint Checks

**Input**: Design documents from `/specs/008-strict-golangci-lint/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, quickstart.md ✓

**Tests**: No tests requested - this feature improves existing linting, verification is via `golangci-lint run`.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup needed - this feature modifies existing codebase

*No tasks in this phase - golangci-lint tooling already configured.*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core fixes that MUST be complete before linters can be enabled

**⚠️ CRITICAL**: Linters cannot be enabled until violations are fixed

- [ ] T001 [P] Fix prealloc: pre-allocate `out` slice in `internal/adapters/dockerlabels/source.go:72`
- [ ] T002 [P] Fix prealloc: pre-allocate `out` slice in `internal/adapters/dockerlabels/source.go:113`
- [ ] T003 [P] Fix prealloc: pre-allocate `out` slice in `internal/adapters/dockerlabels/source.go:148`
- [ ] T004 [P] Fix prealloc: pre-allocate `labelConfigs` slice in `internal/cmd/validate.go:213`
- [ ] T005 [P] Fix prealloc: pre-allocate `sections` slice in `internal/tools/configdoc/markdown.go:100`
- [ ] T006 [P] Fix unparam: remove always-nil error return from `setField` in `internal/config/loader/loader.go:26`
- [ ] T007 [P] Fix unparam: remove unused `fieldType` parameter from `setDefaultValue` in `internal/config/schema/defaults.go:93`
- [ ] T008 [P] Fix unparam: remove always-nil error return from `parseTagValue` in `internal/config/schema/tags.go:17`

**Checkpoint**: All prealloc and unparam violations fixed - ready to enable these linters

---

## Phase 3: User Story 1 - Clean Lint Runs for Contributors (Priority: P1) 🎯 MVP

**Goal**: Enable errcheck, prealloc, and unparam linters with zero violations

**Independent Test**: `golangci-lint run --enable errcheck,prealloc,unparam ./...` returns 0 issues

### Implementation for User Story 1

- [x] T009 [US1] Enable `prealloc` linter in `.golangci.yml` (uncomment from enable section)
- [x] T010 [US1] Enable `unparam` linter in `.golangci.yml` (uncomment from enable section)
- [x] T011 [US1] Enable `errcheck` linter in `.golangci.yml` (remove from disable section)
- [x] T012 [US1] Run `golangci-lint run` to verify zero issues from enabled linters

**Checkpoint**: User Story 1 complete - contributors can now rely on prealloc, unparam, errcheck

---

## Phase 4: User Story 2 - Consistent Code Style via Staticcheck (Priority: P2)

**Goal**: Enable all staticcheck ST* rules with zero violations

**Independent Test**: `golangci-lint run ./...` with no `-ST*` exclusions returns 0 issues

### Implementation for User Story 2

- [x] T013 [P] [US2] Create package comment in `internal/adapters/dockerlabels/doc.go` (ST1000)
- [x] T014 [P] [US2] Create package comment in `internal/app/doc.go` (ST1000)
- [x] T015 [P] [US2] Create package comment in `internal/domain/labels/doc.go` (ST1000)
- [x] T016 [P] [US2] Create package comment in `internal/ports/doc.go` (ST1000)
- [x] T017 [P] [US2] Fix receiver name `d` → `s` in `Snapshot` method in `internal/adapters/dockerlabels/source.go` (ST1016)
- [x] T018 [P] [US2] Fix comment format for `LabelInstance` constant in `internal/domain/labels/types.go` (ST1022)
- [x] T019 [US2] Remove ST1000, ST1016, ST1022 exclusions from `.golangci.yml` staticcheck.checks
- [x] T020 [US2] Run `golangci-lint run` to verify zero staticcheck ST* violations

**Checkpoint**: User Story 2 complete - all staticcheck rules enforced

---

## Phase 5: User Story 3 - Test Code Quality via govet (Priority: P3)

**Goal**: Enable unusedwrite analyzer with zero violations in test files

**Independent Test**: `golangci-lint run --build-tags integration ./...` with unusedwrite enabled returns 0 issues

### Implementation for User Story 3

- [x] T021 [P] [US3] Fix unusedwrite: remove/use `TakenAt` field in `internal/domain/labels/types_test.go:45`
- [x] T022 [P] [US3] Fix unusedwrite: remove/use `Prefixes` and `IncludeStopped` fields in `internal/ports/labels_test.go:54-55`
- [x] T023 [P] [US3] Fix unusedwrite: remove/use `Prefixes` and `IncludeStopped` fields in `internal/ports/labels_test.go:71-72`
- [x] T024 [P] [US3] Fix unusedwrite: remove/use `Prefixes` and `IncludeStopped` fields in `internal/ports/labels_test.go:87-88`
- [x] T025 [US3] Enable `unusedwrite` analyzer in `.golangci.yml` govet.disable (remove from disable list)
- [x] T026 [US3] Run `golangci-lint run --build-tags integration` to verify zero unusedwrite violations

**Checkpoint**: User Story 3 complete - test code quality enforced

---

## Phase 6: Polish & Verification

**Purpose**: Final validation and documentation

- [x] T027 Run `golangci-lint config verify` to validate configuration
- [x] T028 Run full lint suite `golangci-lint run ./...` to confirm 0 issues
- [x] T029 Run `make test` to ensure all unit tests still pass
- [x] T030 Run `make it` to ensure all integration tests still pass
- [x] T031 Verify `.golangci.yml` has no disabled linters and no ST* exclusions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: N/A - no setup tasks
- **Phase 2 (Foundational)**: Must complete before any linters enabled - T001-T008
- **Phase 3 (US1)**: Depends on Phase 2 - enables core linters
- **Phase 4 (US2)**: Can start in parallel with US1 after Phase 2, but T019 must wait for T013-T018
- **Phase 5 (US3)**: Can start in parallel with US1/US2 after Phase 2, but T025 must wait for T021-T024
- **Phase 6 (Polish)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires T001-T008 (Foundational) - MVP increment
- **User Story 2 (P2)**: Independent of US1 - can proceed in parallel
- **User Story 3 (P3)**: Independent of US1/US2 - can proceed in parallel

### Parallel Opportunities

**Phase 2 (All parallel - different files)**:
```
T001, T002, T003  → internal/adapters/dockerlabels/source.go (same file, but independent lines)
T004              → internal/cmd/validate.go
T005              → internal/tools/configdoc/markdown.go
T006              → internal/config/loader/loader.go
T007              → internal/config/schema/defaults.go
T008              → internal/config/schema/tags.go
```

**User Story 2 (Parallel doc.go creation)**:
```
T013, T014, T015, T016  → 4 different doc.go files
T017                    → internal/adapters/dockerlabels/source.go
T018                    → internal/domain/labels/types.go
```

**User Story 3 (Parallel test fixes)**:
```
T021  → internal/domain/labels/types_test.go
T022, T023, T024  → internal/ports/labels_test.go (same file, independent fixes)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001-T008)
2. Complete Phase 3: User Story 1 (T009-T012)
3. **STOP and VALIDATE**: `golangci-lint run --enable errcheck,prealloc,unparam ./...`
4. Can deploy - core linting enabled

### Incremental Delivery

1. Complete Foundational → prealloc/unparam fixes done
2. Add User Story 1 → errcheck, prealloc, unparam enabled (MVP!)
3. Add User Story 2 → staticcheck ST* rules enabled
4. Add User Story 3 → unusedwrite enabled
5. Polish → Final verification

### Single Developer Strategy

Execute in order:
1. T001-T008 (Foundational - parallel within phase)
2. T009-T012 (US1 - sequential)
3. T013-T020 (US2 - T13-T18 parallel, then T19-T20)
4. T021-T026 (US3 - T21-T24 parallel, then T25-T26)
5. T027-T031 (Polish - sequential)

---

## Summary

| Phase | Tasks | Parallel? | Description |
|-------|-------|-----------|-------------|
| Setup | 0 | - | No setup needed |
| Foundational | 8 | ✅ | prealloc + unparam fixes |
| User Story 1 | 4 | ❌ | Enable core linters |
| User Story 2 | 8 | ✅ (T13-T18) | Staticcheck ST* fixes + enable |
| User Story 3 | 6 | ✅ (T21-T24) | unusedwrite fixes + enable |
| Polish | 5 | ❌ | Final verification |
| **Total** | **31** | | |

---

## Notes

- T001-T003 modify same file but independent functions - can be done in one commit
- T022-T024 modify same file but independent test cases - can be done in one commit
- Commit after each logical group (e.g., "fix prealloc violations")
- Run `golangci-lint run` after each phase to catch regressions early
- Keep `shadow` and `fieldalignment` disabled (out of scope per assumptions)
