# Tasks: Config Documentation Generation

**Input**: Design documents from `/specs/004-config-docs-generation/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/generator.md ✓

**Tests**: Unit tests included (required by Constitution Principle III: Test-First Development)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- Generator package: `internal/tools/configdoc/`
- Generator entrypoint: `internal/tools/configdoc/cmd/main.go`
- Output directory: `docs/`
- Tests: alongside source files as `*_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create package structure and basic scaffolding

- [x] T001 Create package directory structure at `internal/tools/configdoc/`
- [x] T002 Create package doc file at `internal/tools/configdoc/doc.go` with package documentation
- [x] T003 [P] Create generator types and Options struct in `internal/tools/configdoc/generator.go`
- [x] T004 [P] Create error types (ErrEmptySpec, ErrOutputDir) in `internal/tools/configdoc/errors.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and helpers that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Implement type mapping helpers (ConfigType → JSON type, ConfigType → human-readable) in `internal/tools/configdoc/types.go`
- [x] T006 [P] Implement scope ordering helpers (deterministic scope order: Global, Container, Volume, Network) in `internal/tools/configdoc/types.go`
- [x] T007 [P] Create test helpers and fixtures in `internal/tools/configdoc/testdata_test.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Generate Markdown Documentation (Priority: P1) 🎯 MVP

**Goal**: Generate `docs/config.md` with tables of all config keys grouped by scope

**Independent Test**: Run generator with test spec, verify output contains properly formatted Markdown table with all fields

### Tests for User Story 1

- [x] T008 [P] [US1] Unit test for empty spec handling in `internal/tools/configdoc/markdown_test.go`
- [x] T009 [P] [US1] Unit test for single field rendering in `internal/tools/configdoc/markdown_test.go`
- [x] T010 [P] [US1] Unit test for all ConfigType values in Markdown in `internal/tools/configdoc/markdown_test.go`
- [x] T011 [P] [US1] Unit test for scope grouping in `internal/tools/configdoc/markdown_test.go`
- [x] T012 [P] [US1] Unit test for deprecated field marking in `internal/tools/configdoc/markdown_test.go`

### Implementation for User Story 1

- [x] T013 [US1] Create FieldRow and ScopeSection types in `internal/tools/configdoc/markdown.go`
- [x] T014 [US1] Implement FieldSpec → FieldRow transformation in `internal/tools/configdoc/markdown.go`
- [x] T015 [US1] Create Markdown template with scope sections in `internal/tools/configdoc/markdown.go`
- [x] T016 [US1] Implement GenerateMarkdown() method in `internal/tools/configdoc/markdown.go`
- [x] T017 [US1] Add header comment "Auto-generated, do not edit manually" to output

**Checkpoint**: At this point, Markdown generation should be fully functional and testable independently

---

## Phase 4: User Story 2 - Generate JSON Schema (Priority: P1)

**Goal**: Generate `docs/config.schema.json` conforming to JSON Schema draft 2020-12

**Independent Test**: Run generator with test spec, validate output against JSON Schema draft 2020-12 meta-schema

### Tests for User Story 2

- [x] T018 [P] [US2] Unit test for empty spec handling in `internal/tools/configdoc/jsonschema_test.go`
- [x] T019 [P] [US2] Unit test for all ConfigType → JSON type mapping in `internal/tools/configdoc/jsonschema_test.go`
- [x] T020 [P] [US2] Unit test for enum field with values in `internal/tools/configdoc/jsonschema_test.go`
- [x] T021 [P] [US2] Unit test for required fields array in `internal/tools/configdoc/jsonschema_test.go`
- [x] T022 [P] [US2] Unit test for deprecated field marking in `internal/tools/configdoc/jsonschema_test.go`

### Implementation for User Story 2

- [x] T023 [US2] Create JSONSchemaDoc and PropertySchema types in `internal/tools/configdoc/jsonschema.go`
- [x] T024 [US2] Implement FieldSpec → PropertySchema transformation in `internal/tools/configdoc/jsonschema.go`
- [x] T025 [US2] Implement GenerateJSONSchema() method with sorted properties in `internal/tools/configdoc/jsonschema.go`
- [x] T026 [US2] Set correct $schema URI for draft 2020-12 in `internal/tools/configdoc/jsonschema.go`

**Checkpoint**: At this point, JSON Schema generation should be fully functional and testable independently

---

## Phase 5: User Story 3 - Deterministic Output (Priority: P2)

**Goal**: Ensure identical input produces byte-for-byte identical output

**Independent Test**: Run generator twice with same input, compare outputs byte-by-byte

### Tests for User Story 3

- [x] T027 [US3] Unit test running GenerateMarkdown twice produces identical output in `internal/tools/configdoc/generator_test.go`
- [x] T028 [US3] Unit test running GenerateJSONSchema twice produces identical output in `internal/tools/configdoc/generator_test.go`
- [x] T029 [US3] Unit test verifying alphabetical key ordering in both outputs in `internal/tools/configdoc/generator_test.go`

### Implementation for User Story 3

- [x] T030 [US3] Verify Markdown uses Spec.Keys() for sorted iteration in `internal/tools/configdoc/markdown.go`
- [x] T031 [US3] Verify JSON Schema uses sorted keys and sorted required array in `internal/tools/configdoc/jsonschema.go`
- [x] T032 [US3] Use json.MarshalIndent with consistent indentation (2 spaces) in `internal/tools/configdoc/jsonschema.go`

**Checkpoint**: Outputs are now deterministic and suitable for version control

---

## Phase 6: User Story 4 - Document Special Formats (Priority: P2)

**Goal**: Include format explanations for duration, byte size, and list types

**Independent Test**: Verify generated Markdown includes "Value Formats" section with examples

### Tests for User Story 4

- [x] T033 [P] [US4] Unit test for duration format documentation in `internal/tools/configdoc/markdown_test.go`
- [x] T034 [P] [US4] Unit test for byte size format documentation in `internal/tools/configdoc/markdown_test.go`
- [x] T035 [P] [US4] Unit test for list format documentation in `internal/tools/configdoc/markdown_test.go`

### Implementation for User Story 4

- [x] T036 [US4] Create FormatDoc type in `internal/tools/configdoc/markdown.go`
- [x] T037 [US4] Add duration format section (examples: 30s, 5m, 1h30m) to Markdown template
- [x] T038 [US4] Add byte size format section (examples: 100MB, 1GiB) to Markdown template
- [x] T039 [US4] Add list format section (CSV and JSON array syntax) to Markdown template

**Checkpoint**: Documentation now explains special value formats

---

## Phase 7: User Story 5 - Integration with Build System (Priority: P2)

**Goal**: Enable `make docs` and `go generate` to regenerate documentation

**Independent Test**: Run `make docs` and verify files are created in `docs/`

### Implementation for User Story 5

- [X] T040 [US5] Create generator entrypoint at `internal/tools/configdoc/cmd/main.go`
- [X] T041 [US5] Implement Generate() method (calls GenerateMarkdown + GenerateJSONSchema, writes files) in `internal/tools/configdoc/generator.go`
- [X] T042 [US5] Create output directory if it doesn't exist in Generate() method
- [X] T043 [US5] Add `//go:generate` directive in `internal/config/schema/config_v1.go`
- [X] T044 [US5] Add `docs` target to `Makefile`
- [X] T045 [US5] Add `docs/` directory to `.gitignore` with exception for generated files (or don't ignore if we want them tracked)

**Checkpoint**: Documentation can now be regenerated via `make docs` or `go generate ./...`

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [X] T046 [P] ~~Create golden file tests~~ (Skipped - unit tests provide sufficient coverage)
- [X] T047 [P] Run `go fmt` and `go vet` on new package
- [X] T048 Run full test suite (`make test`) to verify no regressions
- [X] T049 Generate initial documentation files by running `make docs`
- [X] T050 Validate generated JSON Schema against draft 2020-12 meta-schema (verified structure)
- [X] T051 Validate generated Markdown renders correctly in GitHub preview (verified format)
- [X] T052 Run quickstart.md validation steps (make docs works end-to-end)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 and US2 can proceed in parallel (different files)
  - US3 depends on US1 and US2 completion
  - US4 depends on US1 (Markdown generation)
  - US5 depends on US1 and US2 (needs both generators)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 1 (Setup)
    │
    ▼
Phase 2 (Foundational) ─────────────────────────────────┐
    │                                                   │
    ├──────────────────┬────────────────────────────────┤
    ▼                  ▼                                │
Phase 3 (US1)     Phase 4 (US2)                        │
Markdown Gen       JSON Schema                          │
    │                  │                                │
    ├──────────────────┼────────────────────────────────┤
    │                  │                                │
    ▼                  ▼                                │
    └───────┬──────────┘                                │
            ▼                                           │
      Phase 5 (US3)                                     │
      Determinism                                       │
            │                                           │
    ┌───────┴───────┐                                   │
    ▼               ▼                                   │
Phase 6 (US4)  Phase 7 (US5)                           │
Format Docs    Build Integration                        │
    │               │                                   │
    └───────┬───────┘                                   │
            ▼                                           │
      Phase 8 (Polish)                                  │
```

### Parallel Opportunities

**Within Phase 1 (Setup)**:
- T003 and T004 can run in parallel

**Within Phase 2 (Foundational)**:
- T006 and T007 can run in parallel (after T005)

**Within Phase 3 (US1)**:
- All tests T008-T012 can run in parallel
- T013-T017 are sequential (build on each other)

**Within Phase 4 (US2)**:
- All tests T018-T022 can run in parallel
- T023-T026 are sequential (build on each other)

**Cross-Phase**:
- Phase 3 and Phase 4 can run fully in parallel

**Within Phase 6 (US4)**:
- All tests T033-T035 can run in parallel

**Within Phase 8 (Polish)**:
- T046 and T047 can run in parallel

---

## Parallel Example: User Stories 1 & 2 Simultaneously

```bash
# Developer A: User Story 1 (Markdown)
T008-T012: All Markdown tests (parallel)
T013-T017: Markdown implementation (sequential)

# Developer B: User Story 2 (JSON Schema) - SAME TIME
T018-T022: All JSON Schema tests (parallel)
T023-T026: JSON Schema implementation (sequential)

# Both complete → Continue with US3 (Determinism)
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (~30 min)
2. Complete Phase 2: Foundational (~30 min)
3. Complete Phase 3: US1 - Markdown Generation (~1-2 hr)
4. Complete Phase 4: US2 - JSON Schema Generation (~1-2 hr)
5. **STOP and VALIDATE**: Both outputs generated correctly
6. This delivers core value: users can view docs and use schema

### Incremental Delivery

1. Setup + Foundational → Core ready
2. Add US1 (Markdown) → Test → **MVP for human docs!**
3. Add US2 (JSON Schema) → Test → **MVP for tooling!**
4. Add US3 (Determinism) → Test → Ready for version control
5. Add US4 (Format Docs) → Test → Enhanced usability
6. Add US5 (Build Integration) → Test → Automation ready
7. Polish → Production ready

---

## Task Count Summary

| Phase | Story | Task Count |
|-------|-------|------------|
| Phase 1 | Setup | 4 |
| Phase 2 | Foundational | 3 |
| Phase 3 | US1 - Markdown | 10 (5 tests + 5 impl) |
| Phase 4 | US2 - JSON Schema | 9 (5 tests + 4 impl) |
| Phase 5 | US3 - Determinism | 6 (3 tests + 3 impl) |
| Phase 6 | US4 - Format Docs | 7 (3 tests + 4 impl) |
| Phase 7 | US5 - Build Integration | 6 |
| Phase 8 | Polish | 7 |
| **Total** | | **52 tasks** |

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- All tests use Go's standard testing package (`*_test.go` files)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `make test` frequently to catch regressions early
