# Tasks: CLI Config Validate Command

**Input**: Design documents from `/specs/003-cli-config-validate/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/cli.md ✅, quickstart.md ✅

**Tests**: Integration tests included - this is a CLI command requiring Docker interaction.

**Organization**: Tasks grouped by user story. US1 and US2 are both P1 (core validation flow), US3-US5 are enhancements.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- CLI commands: `internal/cmd/`
- Config packages: `internal/config/{schema,loader,merge}/`
- Integration tests: `integration/`

---

## Phase 1: Setup ✅

**Purpose**: Project structure and command registration

- [x] T001 Create config command group in `internal/cmd/config.go`
- [x] T002 Register NewConfigCmd() in `internal/cmd/root.go`
- [x] T003 Create validate command skeleton in `internal/cmd/validate.go`

---

## Phase 2: Foundational (Blocking Prerequisites) ✅

**Purpose**: Core types and infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Define ConfigSource enum (auto/labels/file) in `internal/cmd/validate.go`
- [x] T005 Define ValidateOptions struct with all flags in `internal/cmd/validate.go`
- [x] T006 Define EntityValidationError struct in `internal/cmd/validate.go`
- [x] T007 Define ValidationResult struct in `internal/cmd/validate.go`
- [x] T008 Implement flag parsing (--from, --scope, --print, --stopped) in `internal/cmd/validate.go`
- [x] T009 Implement Docker connection with friendly error handling in `internal/cmd/validate.go`
- [x] T010 Implement snapshot loading using existing dockerlabels adapter in `internal/cmd/validate.go`

**Checkpoint**: Command skeleton with flags works, can connect to Docker ✅

---

## Phase 3: User Story 1 & 2 - Core Validation (Priority: P1) 🎯 MVP ✅

**Goal**: Basic validation that passes on valid config and fails with clear errors on invalid config

**Independent Test**: Run `bosun config validate` - should exit 0 on valid labels, exit 1 with errors on invalid

### Tests for US1/US2

- [x] T011 [P] [US1] Create Docker Compose file with valid labels in `internal/testutil/compose/validate-valid.yaml`
- [x] T012 [P] [US2] Create Docker Compose file with invalid labels in `internal/testutil/compose/validate-invalid.yaml`
- [x] T013 [US1] Integration test: valid config exits 0 in `integration/validate_test.go`
- [x] T014 [US2] Integration test: invalid config exits non-zero with errors in `integration/validate_test.go`

### Implementation for US1/US2

- [x] T015 [US1] Implement entity-to-scope mapping (Kind → Scope) in `internal/cmd/validate.go`
- [x] T016 [US1] Implement per-entity validation loop calling loader.FromLabels() in `internal/cmd/validate.go`
- [x] T017 [US2] Implement error collection across all entities in `internal/cmd/validate.go`
- [x] T018 [US2] Implement error formatting with entity context in `internal/cmd/validate.go`
- [x] T019 [US1] Implement success message output in `internal/cmd/validate.go`
- [x] T020 [US2] Implement error output to stderr with exit code 1 in `internal/cmd/validate.go`
- [x] T021 [US1] Implement defaults-only validation (no labels case) in `internal/cmd/validate.go`

**Checkpoint**: `bosun config validate` works for valid and invalid configs - MVP complete! ✅

---

## Phase 4: User Story 3 - Print Effective Configuration (Priority: P2) ✅

**Goal**: Show merged config as JSON with `--print` flag

**Independent Test**: Run `bosun config validate --print` - should output valid JSON

### Tests for US3

- [x] T022 [P] [US3] Integration test: --print outputs valid JSON in `integration/validate_test.go`

### Implementation for US3

- [x] T023 [US3] Implement config merging using merge.Merge() in `internal/cmd/validate.go`
- [x] T024 [US3] Implement JSON output for --print flag in `internal/cmd/validate.go`
- [x] T025 [US3] Ensure --print still reports errors (no JSON on failure) in `internal/cmd/validate.go`

**Checkpoint**: `--print` flag works, shows merged config as JSON ✅

---

## Phase 5: User Story 4 - Select Configuration Source (Priority: P2) ✅

**Goal**: Control which sources are validated with `--from` flag

**Independent Test**: Run `bosun config validate --from labels` - should ignore file config

### Tests for US4

- [x] T026 [P] [US4] Integration test: --from labels ignores file in `integration/validate_test.go` (covered by existing tests)

### Implementation for US4

- [x] T027 [US4] Implement --from labels (skip file layer) in `internal/cmd/validate.go`
- [x] T028 [US4] Implement --from file warning (not implemented) in `internal/cmd/validate.go`
- [x] T029 [US4] Implement --from auto (merge all, default) in `internal/cmd/validate.go`

**Checkpoint**: `--from` flag controls source selection ✅

---

## Phase 6: User Story 5 - Scope-Specific Validation (Priority: P3) ✅

**Goal**: Filter validation to specific entity types with `--scope` flag

**Independent Test**: Run `bosun config validate --scope container` - should skip volumes/networks

### Tests for US5

- [x] T030 [P] [US5] Integration test: --scope container skips volumes in `integration/validate_test.go`

### Implementation for US5

- [x] T031 [US5] Implement entity filtering by scope in `internal/cmd/validate.go`
- [x] T032 [US5] Ensure global scope labels are always included in `internal/cmd/validate.go`

**Checkpoint**: `--scope` flag filters by entity type ✅

---

## Phase 7: Polish & Cross-Cutting Concerns ✅

**Purpose**: Improvements that affect multiple user stories

- [x] T033 [P] Add doc comments to all exported functions in `internal/cmd/config.go`
- [x] T034 [P] Add doc comments to all exported functions in `internal/cmd/validate.go`
- [x] T035 Run quickstart.md scenarios manually to verify examples work
- [x] T036 Update cli_commands memory with new config validate command

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **US1/US2 (Phase 3)**: Depend on Foundational - Core MVP
- **US3 (Phase 4)**: Can start after Phase 3 (needs validation working first)
- **US4 (Phase 5)**: Can start after Phase 3 (needs validation working first)
- **US5 (Phase 6)**: Can start after Phase 3 (needs validation working first)
- **Polish (Phase 7)**: Depends on all phases complete

### User Story Dependencies

- **US1 + US2 (P1)**: Must be implemented together (success + failure paths)
- **US3 (P2)**: Independent of US4/US5
- **US4 (P2)**: Independent of US3/US5
- **US5 (P3)**: Independent of US3/US4

### Within Each Phase

- Tests MUST be written and FAIL before implementation
- Type definitions before logic
- Core logic before output formatting

### Parallel Opportunities

- T011/T012: Compose files can be created in parallel
- T033/T034: Doc comments can be added in parallel
- US3/US4/US5: Can be developed in parallel after MVP (Phase 3)

---

## Parallel Example: Phase 3 Tests

```bash
# Launch all test compose files together:
Task T011: "Create Docker Compose file with valid labels"
Task T012: "Create Docker Compose file with invalid labels"
```

---

## Implementation Strategy

### MVP First (Phase 1-3 Only)

1. Complete Phase 1: Setup → Command registered
2. Complete Phase 2: Foundational → Flags and Docker connection work
3. Complete Phase 3: US1/US2 → Core validation works
4. **STOP and VALIDATE**: Test with real Docker environment
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1/US2 → Test → Deploy (MVP! Core validation)
3. Add US3 → Test → Deploy (--print flag)
4. Add US4 → Test → Deploy (--from flag)
5. Add US5 → Test → Deploy (--scope flag)
6. Polish → Final release

---

## Notes

- All implementation is in `internal/cmd/validate.go` - no new packages needed
- Reuses existing: `loader.FromLabels()`, `merge.Merge()`, `schema.V1Spec()`, `schema.V1Defaults()`
- Exit codes: 0=success, 1=validation error, 2=runtime error (Docker unavailable)
- US1 and US2 are combined because they're two sides of the same coin (valid/invalid)
- File config loading (--from file) is deferred - just warn for now
