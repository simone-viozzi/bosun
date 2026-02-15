# Tasks: Milestone 3.75 — Critical Bug Fixes

**Input**: Design documents from `specs/011-critical-bug-fixes/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Not explicitly requested — omitting dedicated test tasks. Existing unit + integration tests validate the fixes.

**Organization**: Tasks grouped by user story (mapped to GitHub issues). US3 is independent and parallelizable with US1/US2.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 = Worker Logs (#153), US2 = Error Handling (#144), US3 = Coverage Files (#161)
- Exact file paths included in all tasks

---

## Phase 1: Setup

**Purpose**: No project initialization needed — this is a bug-fix milestone on an existing codebase. Phase is empty.

_(No tasks — codebase already exists and is set up.)_

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define exit code constants that US1 and US2 both depend on

- [X] T001 Define named exit code constants (`ExitCodeSIGKILL = 137`, `ExitCodeSIGTERM = 143`) with documentation in `internal/adapters/docker/worker/runner.go`

**Checkpoint**: Constants available — user story implementation can begin

---

## Phase 3: User Story 1 — Worker Logs Are Readable (Priority: P0) 🎯 MVP

**Goal**: Fix corrupted worker log output caused by raw `io.Copy` on Docker's multiplexed log stream. Replace with `stdcopy.StdCopy` in both production and test code.

**Independent Test**: Run `bosun job run` and verify output contains no binary header bytes

### Implementation for User Story 1

- [X] T002 [US1] Import `github.com/docker/docker/pkg/stdcopy` and replace `io.Copy(writer, reader)` with `stdcopy.StdCopy(writer, writer, reader)` in `streamLogs()` in `internal/adapters/docker/worker/runner.go`
- [X] T003 [P] [US1] Replace `io.Copy(f, rc)` with `stdcopy.StdCopy(f, f, rc)` in `DumpLogs()` in `internal/testutil/docker.go`
- [X] T004 [US1] Remove the TODO/BUG comments about missing stdcopy in both `internal/adapters/docker/worker/runner.go` and `internal/testutil/docker.go`

**Checkpoint**: Worker logs display clean text. Run `make test` to verify no regressions.

---

## Phase 4: User Story 2 — Worker Errors Are Visible (Priority: P1)

**Goal**: Eliminate silent error swallowing in the worker runner. Return/log errors from `streamLogs` and `stopContainer`. Replace magic exit code `137` with named constant.

**Independent Test**: Code inspection: `grep -r "return$" internal/adapters/docker/worker/` finds no silent returns after error checks

### Implementation for User Story 2

- [X] T005 [US2] Change `streamLogs()` signature to return `error` instead of void, return errors from `ContainerLogs` and `stdcopy.StdCopy` calls in `internal/adapters/docker/worker/runner.go`
- [X] T006 [US2] Update the goroutine in `Run()` that calls `streamLogs()` to capture and log the returned error (best-effort: log warning, don't fail job) in `internal/adapters/docker/worker/runner.go`
- [X] T007 [US2] Change `stopContainer()` signature from `int` to `(int, error)`, log `ContainerStop` error instead of ignoring it, return error from `ContainerInspect` instead of hardcoded `137`, use `ExitCodeSIGKILL` constant as fallback in `internal/adapters/docker/worker/runner.go`
- [X] T008 [US2] Update the call site in `Run()` that calls `stopContainer()` to handle the new `(int, error)` return, log the error if non-nil in `internal/adapters/docker/worker/runner.go`
- [X] T009 [US2] Remove all TODO/SMELL/BUG comments about silent error swallowing and magic exit codes in `internal/adapters/docker/worker/runner.go`

**Checkpoint**: No silent error drops in worker runner. Run `make test` to verify no regressions.

---

## Phase 5: User Story 3 — Clean Repository Root (Priority: P2)

**Goal**: Move all coverage output files from the repository root to a `coverage/` subdirectory. Update Makefile and gitignore.

**Independent Test**: Run `make coverage` and verify files land in `coverage/`, not repo root

### Implementation for User Story 3

- [X] T010 [P] [US3] Add `COVERAGE_DIR := coverage` variable to `Makefile` and update `coverage` target to use `$(COVERAGE_DIR)/coverage.out` and `$(COVERAGE_DIR)/coverage.txt` in `Makefile`
- [X] T011 [US3] Update `coverage-html` target to output `$(COVERAGE_DIR)/coverage.html` in `Makefile`
- [X] T012 [US3] Update `it-cover` target to use `$(COVERAGE_DIR)/covdata-integration` directory in `Makefile`
- [X] T013 [US3] Update `coverage-integration` target to use `$(COVERAGE_DIR)/coverage.integration.out` in `Makefile`
- [X] T014 [US3] Update `coverage-all` target to use `$(COVERAGE_DIR)/covdata-unit`, `$(COVERAGE_DIR)/covdata-merged`, `$(COVERAGE_DIR)/coverage.all.out`, `$(COVERAGE_DIR)/coverage.all.txt`, `$(COVERAGE_DIR)/coverage.all.html` in `Makefile`
- [X] T015 [P] [US3] Replace individual coverage gitignore entries (`*.out`, `coverage.html`, `coverage.txt`, `coverage.all.html`, `coverage.all.txt`, `covdata*/`) with single `coverage/` entry in `.gitignore`
- [X] T016 [US3] Delete leftover root-level coverage files and `covdata-*` directories (untracked, local cleanup)

**Checkpoint**: Repository root is clean after `make coverage`. Run `make test` to verify no regressions.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [X] T017 Run `make fmt` and `make vet` to ensure code quality in all changed files
- [X] T018 Run `make test` to verify all unit tests pass
- [X] T019 Run `make it` to verify all integration tests pass (requires Docker)
- [X] T020 Run quickstart.md verification commands to validate all three fixes end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — can start immediately
- **US1 (Phase 3)**: Depends on Phase 2 (uses constants indirectly, touches same file)
- **US2 (Phase 4)**: Depends on Phase 3 (builds on streamLogs changes from US1)
- **US3 (Phase 5)**: **No dependencies on Phase 2–4** — can run in parallel with US1/US2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P0)**: Start after Phase 2. Implements the stdcopy fix that US2 builds upon.
- **US2 (P1)**: Start after US1. Both touch `streamLogs()` in the same file — US2 changes its signature and error handling after US1 fixes the copy mechanism.
- **US3 (P2)**: **Independent** — different files entirely (Makefile, .gitignore). Can run in parallel with US1/US2.

### Within Each User Story

- Implementation tasks are ordered by dependency (signature changes before call site updates)
- TODO cleanup tasks come last within each story
- Commit after each story checkpoint

### Parallel Opportunities

- **T003** (DumpLogs fix) can run in parallel with T002 (streamLogs fix) — different files
- **T010** and **T015** can run in parallel — different files (Makefile vs .gitignore)
- **US3 (Phase 5)** can run entirely in parallel with US1 + US2 — no file overlap

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete T001 (exit code constants)
2. Complete T002–T004 (US1: stdcopy fix)
3. **STOP and VALIDATE**: `make test` + manual `bosun job run`
4. Worker logs are now clean — critical bug fixed

### Incremental Delivery

1. T001 → Foundation ready
2. T002–T004 → US1 complete → Worker logs readable (**MVP**)
3. T005–T009 → US2 complete → Errors visible, exit codes documented
4. T010–T016 → US3 complete → Repository root clean
5. T017–T020 → Polish → All validation passes

### Parallel Track

While US1 → US2 proceeds sequentially (same file):
- US3 (T010–T016) can be done in parallel by a second pass or developer
- No file conflicts between US3 and US1/US2

---

## Summary

| Metric | Value |
|--------|-------|
| **Total tasks** | 20 |
| **Foundational** | 1 (T001) |
| **US1 tasks** | 3 (T002–T004) |
| **US2 tasks** | 5 (T005–T009) |
| **US3 tasks** | 7 (T010–T016) |
| **Polish tasks** | 4 (T017–T020) |
| **Parallel opportunities** | T003∥T002, T010∥T015, US3∥US1+US2 |
| **MVP scope** | T001–T004 (US1 only — 4 tasks) |
