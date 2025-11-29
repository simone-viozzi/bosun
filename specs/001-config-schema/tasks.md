# Tasks: Code-First Config Schema with Tags

**Input**: Design documents from `/specs/001-config-schema/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, quickstart.md ✅

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md, using hexagonal architecture:
- **Config schema**: `internal/config/schema/`
- **Tests**: `internal/config/schema/*_test.go` (alongside source files)

---

## Phase 1: Setup ✅

**Purpose**: Package initialization and type foundations

- [X] T001 Create package directory structure at `internal/config/schema/`
- [X] T002 [P] Create `internal/config/schema/doc.go` with package documentation

---

## Phase 2: Foundational (Core Types) ✅

**Purpose**: Define core types that ALL user stories depend on

**⚠️ CRITICAL**: Types must be complete before tag parsing can be implemented

- [X] T003 Define `Scope` type and constants (ScopeContainer, ScopeVolume, ScopeNetwork, ScopeGlobal) in `internal/config/schema/types.go`
- [X] T004 Define `ConfigType` type and constants (TypeString, TypeBool, TypeInt, TypeDuration, TypeSize, TypeEnum, TypeList) in `internal/config/schema/types.go`
- [X] T005 Define `FieldSpec` struct with all attributes (Key, Scope, Type, Default, Enum, Required, Doc, Deprecated, FieldName) in `internal/config/schema/types.go`
- [X] T006 Define `Spec` type as `map[string]FieldSpec` with helper methods (Keys, Get, Scopes) in `internal/config/schema/types.go`
- [X] T007 [P] Add validation helpers `IsValidScope(s string) bool` and `IsValidConfigType(t string) bool` in `internal/config/schema/types.go`
- [X] T008 [P] Write unit tests for Scope and ConfigType validation in `internal/config/schema/types_test.go`

**Checkpoint**: Core types ready - tag parsing can now be implemented ✅

---

## Phase 3: User Story 1+2+4 - Tag Parsing (Priority: P1) 🎯 MVP ✅

**Goal**: Parse `bosun:` struct tags into machine-readable Spec via `ParseTags[T]()`

**Independent Test**: Call `ParseTags` on a test struct and verify all metadata extracted correctly

**Why combined**: US1 (define fields), US2 (ParseTags), and US4 (type support) are tightly coupled - parsing must support all types to be useful

### Implementation for Tag Parsing

- [X] T009 Implement tag string parser `parseTagValue(tagValue string) (map[string]string, error)` that handles quoted doc strings in `internal/config/schema/tags.go`
- [X] T010 Implement `parseFieldSpec(fieldName string, tagParts map[string]string) (FieldSpec, error)` with validation in `internal/config/schema/tags.go`
- [X] T011 Implement `ParseTags[T any]() (Spec, error)` generic function using reflection in `internal/config/schema/tags.go`
- [X] T012 Add embedded struct handling in `ParseTags` (recursive field processing) in `internal/config/schema/tags.go`
- [X] T013 Add duplicate key detection in `ParseTags` in `internal/config/schema/tags.go`
- [X] T014 [P] Write unit tests for `parseTagValue` (basic parsing, quoted strings, edge cases) in `internal/config/schema/tags_test.go`
- [X] T015 [P] Write unit tests for `parseFieldSpec` (all types, validation errors) in `internal/config/schema/tags_test.go`
- [X] T016 Write unit tests for `ParseTags` (full struct, embedded structs, error cases) in `internal/config/schema/tags_test.go`

**Checkpoint**: `ParseTags[T]()` functional with all 7 types supported ✅

---

## Phase 4: User Story 3 - Default Hydration (Priority: P2) ✅

**Goal**: Create `DefaultOf[T]()` to populate struct with default values from tags

**Independent Test**: Call `DefaultOf[ConfigV1]()` and verify all defaults match tag declarations

### Implementation for Default Hydration

- [X] T017 Implement value parsers for each type (string, bool, int, duration, size, list) in `internal/config/schema/defaults.go`
- [X] T018 Implement `DefaultOf[T any]() (T, error)` generic function using reflection in `internal/config/schema/defaults.go`
- [X] T019 [P] Write unit tests for individual value parsers in `internal/config/schema/defaults_test.go`
- [X] T020 Write unit tests for `DefaultOf` (all types, missing defaults, parse errors) in `internal/config/schema/defaults_test.go`

**Checkpoint**: `DefaultOf[T]()` returns correctly populated structs ✅

---

## Phase 5: User Story 5 - V1 Config Definition (Priority: P1) ✅

**Goal**: Define concrete ConfigV1 struct demonstrating all supported types

**Independent Test**: `ParseTags[ConfigV1]()` succeeds and returns spec with all expected keys

### Implementation for ConfigV1

- [X] T021 Define `GlobalConfig` embedded struct with Instance field in `internal/config/schema/config_v1.go`
- [X] T022 Define `ContainerConfig` embedded struct with StopGracePeriod, HealthCheckInterval, AutoRestart, LogLevel fields in `internal/config/schema/config_v1.go`
- [X] T023 Define `VolumeConfig` embedded struct with BackupEnabled, MaxSize fields in `internal/config/schema/config_v1.go`
- [X] T024 Define `NetworkConfig` embedded struct with Priority field in `internal/config/schema/config_v1.go`
- [X] T025 Define `ConfigV1` struct embedding all config groups in `internal/config/schema/config_v1.go`
- [X] T026 [P] Write integration test verifying `ParseTags[ConfigV1]()` returns complete spec in `internal/config/schema/config_v1_test.go`
- [X] T027 [P] Write integration test verifying `DefaultOf[ConfigV1]()` returns correct defaults in `internal/config/schema/config_v1_test.go`

**Checkpoint**: ConfigV1 fully defined and tested ✅

---

## Phase 6: User Story 5 - Deprecated Flag (Priority: P3) ✅

**Goal**: Support `deprecated=true` tag option for future migrations

**Independent Test**: Field with `deprecated=true` tag has `Deprecated: true` in FieldSpec

### Implementation for Deprecated

- [X] T028 Add deprecated field to a test struct and verify parsing in `internal/config/schema/tags_test.go`

**Checkpoint**: Deprecated flag parsed correctly (no-op for now) ✅

---

## Phase 7: Polish & Validation ✅

**Purpose**: Final cleanup and validation

- [X] T029 Run `go fmt` and `go vet` on all schema files
- [X] T030 Run `golangci-lint` and fix any issues
- [X] T031 Verify all acceptance criteria from spec.md are met
- [X] T032 Run quickstart.md examples to validate API usage

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational Types) ─── BLOCKS ALL ───→ Phase 3, 4, 5
    ↓
Phase 3 (Tag Parsing - P1 MVP)
    ↓
Phase 4 (Defaults - P2) ←─ can start after T011
    ↓
Phase 5 (ConfigV1 - P1) ←─ can start after T011
    ↓
Phase 6 (Deprecated - P3) ←─ already covered in T009-T016
    ↓
Phase 7 (Polish)
```

### Parallel Opportunities

**Within Phase 2**:
```bash
# After T006 completes:
T007 (validation helpers) || T008 (types tests)
```

**Within Phase 3**:
```bash
# After T013 completes:
T014 (parseTagValue tests) || T015 (parseFieldSpec tests)
```

**Within Phase 4**:
```bash
# After T018 completes:
T019 (parser tests) || T020 (DefaultOf tests)
```

**Within Phase 5**:
```bash
# After T025 completes:
T026 (ParseTags test) || T027 (DefaultOf test)
```

---

## Implementation Strategy

### MVP First (Tag Parsing Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational Types
3. Complete Phase 3: Tag Parsing (ParseTags[T])
4. Complete Phase 5: ConfigV1 (T021-T025 only, skip tests)
5. **VALIDATE**: `ParseTags[ConfigV1]()` works
6. This delivers the core value for downstream issues #58, #59, #61

### Full Implementation

1. MVP above
2. Add Phase 4: Default Hydration
3. Add Phase 5 tests: T026, T027
4. Add Phase 7: Polish

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 32 |
| Setup Tasks | 2 |
| Foundational Tasks | 6 |
| US1+2+4 (Tag Parsing) Tasks | 8 |
| US3 (Defaults) Tasks | 4 |
| US5 (ConfigV1) Tasks | 7 |
| US5 (Deprecated) Tasks | 1 |
| Polish Tasks | 4 |
| Parallel Tasks | 10 |

### MVP Scope

Tasks T001-T016, T021-T025 (21 tasks) deliver a complete, working `ParseTags[T]()` with ConfigV1.

### Files Created

| File | Purpose |
|------|---------|
| `internal/config/schema/doc.go` | Package documentation |
| `internal/config/schema/types.go` | Scope, ConfigType, FieldSpec, Spec |
| `internal/config/schema/types_test.go` | Type validation tests |
| `internal/config/schema/tags.go` | Tag parsing, ParseTags[T]() |
| `internal/config/schema/tags_test.go` | Tag parsing tests |
| `internal/config/schema/defaults.go` | DefaultOf[T]() |
| `internal/config/schema/defaults_test.go` | Default hydration tests |
| `internal/config/schema/config_v1.go` | ConfigV1 struct definition |
| `internal/config/schema/config_v1_test.go` | ConfigV1 integration tests |
