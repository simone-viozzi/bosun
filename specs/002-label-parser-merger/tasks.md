# Tasks: Label Parser and Source Merger

**Input**: Design documents from `/specs/002-label-parser-merger/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create package structure and shared types

- [X] T001 Create loader package structure with `internal/config/loader/doc.go`
- [X] T002 [P] Create merge package structure with `internal/config/merge/doc.go`
- [X] T003 [P] Implement ValidationError and ValidationErrors types in `internal/config/loader/errors.go`
- [X] T004 [P] Implement error tests in `internal/config/loader/errors_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core parsing infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Implement label filtering function `filterBosunLabels(labels map[string]string) map[string]string` in `internal/config/loader/loader.go`
- [X] T006 Implement type parsing functions (parseBool, parseInt, parseDuration, parseSize, parseEnum, parseList) in `internal/config/loader/parse.go`
- [X] T007 [P] Implement parse function unit tests in `internal/config/loader/parse_test.go`
- [X] T008 Implement reflection-based field setter `setField(cfg *ConfigV1, fieldSpec FieldSpec, value any) error` in `internal/config/loader/loader.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Parse Valid Labels into Typed Config (Priority: P1) 🎯 MVP

**Goal**: Parse Docker labels into strongly-typed ConfigV1 struct with all 7 types working

**Independent Test**: Run unit tests with valid labels for each type, verify correct typed values

### Implementation for User Story 1

- [X] T009 [US1] Implement `FromLabels(spec Spec, labels map[string]string, scope Scope) (ConfigV1, error)` core function in `internal/config/loader/loader.go`
- [X] T010 [US1] Add tests for string type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_StringType)
- [X] T011 [P] [US1] Add tests for bool type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_BoolType)
- [X] T012 [P] [US1] Add tests for int type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_IntType)
- [X] T013 [P] [US1] Add tests for duration type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_DurationType)
- [X] T014 [P] [US1] Add tests for size type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_SizeType)
- [X] T015 [P] [US1] Add tests for enum type parsing in `internal/config/loader/loader_test.go` (TestFromLabels_EnumType)
- [X] T016 [P] [US1] Add tests for list type parsing (CSV and JSON) in `internal/config/loader/loader_test.go` (TestFromLabels_ListType)

**Checkpoint**: User Story 1 complete - all 7 types parse correctly

---

## Phase 4: User Story 2 - Fail Fast on Unknown Keys (Priority: P1)

**Goal**: Reject unknown `bosun.*` keys with clear error messages

**Independent Test**: Provide unknown label key, verify error message contains the key name

### Implementation for User Story 2

- [X] T017 [US2] Implement unknown key detection in `FromLabels()` that collects unknown keys into ValidationErrors in `internal/config/loader/loader.go`
- [X] T018 [US2] Add test for single unknown key in `internal/config/loader/loader_test.go` (TestFromLabels_UnknownKey)
- [X] T019 [P] [US2] Add test for multiple unknown keys in `internal/config/loader/loader_test.go` (TestFromLabels_MultipleUnknownKeys)
- [X] T020 [P] [US2] Add test verifying all unknown keys are reported (not just first) in `internal/config/loader/loader_test.go` (TestFromLabels_AllErrorsReported)

**Checkpoint**: User Story 2 complete - unknown keys fail with clear messages

---

## Phase 5: User Story 3 - Validate Label Scopes (Priority: P1)

**Goal**: Reject labels applied to wrong entity types (e.g., container label on volume)

**Independent Test**: Apply container-scoped label to volume scope, verify scope mismatch error

### Implementation for User Story 3

- [X] T021 [US3] Implement scope validation logic in `FromLabels()` checking `fieldSpec.Scope` against entity scope in `internal/config/loader/loader.go`
- [X] T022 [US3] Add test for scope mismatch (container label on volume) in `internal/config/loader/loader_test.go` (TestFromLabels_ScopeMismatch)
- [X] T023 [P] [US3] Add test for global scope allowed on any entity in `internal/config/loader/loader_test.go` (TestFromLabels_GlobalScopeAllowed)
- [X] T024 [P] [US3] Add test for matching scope (network label on network) in `internal/config/loader/loader_test.go` (TestFromLabels_MatchingScope)

**Checkpoint**: User Story 3 complete - scope mismatches are rejected

---

## Phase 6: User Story 4 - Validate Type Parsing (Priority: P2)

**Goal**: Provide clear error messages for type parsing failures

**Independent Test**: Provide invalid duration/bool/size/enum values, verify specific error messages

### Implementation for User Story 4

- [X] T025 [US4] Enhance parse functions to return descriptive errors with key context in `internal/config/loader/parse.go`
- [X] T026 [US4] Add test for invalid duration error message in `internal/config/loader/loader_test.go` (TestFromLabels_InvalidDuration)
- [X] T027 [P] [US4] Add test for invalid bool error message in `internal/config/loader/loader_test.go` (TestFromLabels_InvalidBool)
- [X] T028 [P] [US4] Add test for invalid size error message in `internal/config/loader/loader_test.go` (TestFromLabels_InvalidSize)
- [X] T029 [P] [US4] Add test for invalid enum error message (should list valid values) in `internal/config/loader/loader_test.go` (TestFromLabels_InvalidEnum)
- [X] T030 [P] [US4] Add test for required field missing in `internal/config/loader/loader_test.go` (TestFromLabels_RequiredMissing)

**Checkpoint**: User Story 4 complete - type errors have clear messages

---

## Phase 7: User Story 5 - Merge Config from Multiple Sources (Priority: P2)

**Goal**: Merge configs with precedence: defaults < file < env < labels

**Independent Test**: Provide conflicting values in defaults and labels, verify labels win

### Implementation for User Story 5

- [X] T031 [US5] Create Options struct with EnableEnv field in `internal/config/merge/merge.go`
- [X] T032 [US5] Implement `Merge(spec Spec, defaults ConfigV1, file, env, labels *ConfigV1, opts Options) (ConfigV1, error)` in `internal/config/merge/merge.go`
- [X] T033 [US5] Implement `mergeLayer(base, override ConfigV1) ConfigV1` helper using reflection in `internal/config/merge/merge.go`
- [X] T034 [US5] Implement `isZeroValue(v reflect.Value) bool` helper in `internal/config/merge/merge.go`
- [X] T035 [US5] Add test for defaults only (no overrides) in `internal/config/merge/merge_test.go` (TestMerge_DefaultsOnly)
- [X] T036 [P] [US5] Add test for labels override defaults in `internal/config/merge/merge_test.go` (TestMerge_LabelsOverrideDefaults)
- [X] T037 [P] [US5] Add test for file overrides defaults in `internal/config/merge/merge_test.go` (TestMerge_FileOverridesDefaults)
- [X] T038 [P] [US5] Add test for labels override file in `internal/config/merge/merge_test.go` (TestMerge_LabelsOverrideFile)
- [X] T039 [P] [US5] Add test for nil layers handling in `internal/config/merge/merge_test.go` (TestMerge_NilLayers)
- [X] T040 [US5] Add test for deterministic output in `internal/config/merge/merge_test.go` (TestMerge_Deterministic)

**Checkpoint**: User Story 5 complete - config merging works with correct precedence

---

## Phase 8: User Story 6 - Optional Environment Variable Layer (Priority: P3)

**Goal**: Support optional env layer controlled by feature flag

**Independent Test**: Enable env flag, verify env values override file but not labels

### Implementation for User Story 6

- [X] T041 [US6] Implement env layer logic in Merge() respecting Options.EnableEnv flag in `internal/config/merge/merge.go`
- [X] T042 [US6] Add test for env layer disabled (ignored) in `internal/config/merge/merge_test.go` (TestMerge_EnvDisabled)
- [X] T043 [P] [US6] Add test for env layer enabled (overrides file) in `internal/config/merge/merge_test.go` (TestMerge_EnvEnabled)
- [X] T044 [P] [US6] Add test for env does not override labels in `internal/config/merge/merge_test.go` (TestMerge_EnvDoesNotOverrideLabels)

**Checkpoint**: User Story 6 complete - env layer is optional and works correctly

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup and documentation

- [X] T045 [P] Add package documentation to `internal/config/loader/doc.go`
- [X] T046 [P] Add package documentation to `internal/config/merge/doc.go`
- [X] T047 Run `go test -cover ./internal/config/...` and verify >80% coverage
- [X] T048 Run `go vet ./internal/config/...` and fix any issues
- [X] T049 Validate quickstart.md examples compile and work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-8)**: All depend on Foundational phase completion
  - US1, US2, US3 are all P1 - should be done sequentially (shared loader.go)
  - US4 can start after US1-3 (builds on their foundation)
  - US5, US6 are independent (merge package) - can run in parallel with US4
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundational → US1 (no story deps)
- **User Story 2 (P1)**: Foundational → US1 → US2 (extends FromLabels)
- **User Story 3 (P1)**: Foundational → US1 → US3 (extends FromLabels)
- **User Story 4 (P2)**: Foundational → US1-3 → US4 (better error messages)
- **User Story 5 (P2)**: Foundational → US5 (independent, merge package)
- **User Story 6 (P3)**: US5 → US6 (extends Merge)

### Parallel Opportunities Per Phase

**Phase 1 (Setup)**:
- T001, T002, T003, T004 can all run in parallel

**Phase 2 (Foundational)**:
- T007 can run in parallel with T005, T006, T008

**Phase 3-6 (Loader Stories)**:
- Tests within each phase marked [P] can run in parallel
- Different test functions don't conflict

**Phase 7-8 (Merge Stories)**:
- Tests marked [P] can run in parallel
- Can run entirely in parallel with Phase 6 (different package)

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: US1 - Basic parsing works
4. Complete Phase 4: US2 - Unknown keys rejected
5. Complete Phase 5: US3 - Scope validation works
6. **STOP and VALIDATE**: Core loader is complete and testable

### Full Feature (Add US4-6)

7. Complete Phase 6: US4 - Better error messages
8. Complete Phase 7: US5 - Config merging
9. Complete Phase 8: US6 - Env layer (optional)
10. Complete Phase 9: Polish

### Task Count Summary

| Phase | Tasks | Parallel Tasks |
|-------|-------|----------------|
| Phase 1: Setup | 4 | 3 |
| Phase 2: Foundational | 4 | 1 |
| Phase 3: US1 | 8 | 6 |
| Phase 4: US2 | 4 | 2 |
| Phase 5: US3 | 4 | 2 |
| Phase 6: US4 | 6 | 4 |
| Phase 7: US5 | 10 | 4 |
| Phase 8: US6 | 4 | 2 |
| Phase 9: Polish | 5 | 2 |
| **Total** | **49** | **26** |
