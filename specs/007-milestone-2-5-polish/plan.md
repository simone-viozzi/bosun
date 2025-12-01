# Implementation Plan: Milestone 2.5 – Polish Job Model, Config Schema & Test Suite

**Branch**: `007-milestone-2-5-polish` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-milestone-2-5-polish/spec.md`

## Summary

Cleanup and polish pass on Milestone 2 work: unify job validation logic into a single canonical location, implement project/stack filtering for CLI isolation, update CLI branding to match README vision, and reduce test redundancy by clarifying unit vs integration test responsibilities.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI), Docker SDK (`github.com/docker/docker`), robfig/cron/v3 (scheduling)
**Storage**: N/A (stateless CLI tool)
**Testing**: Go testing + Docker Compose for integration tests (`//go:build integration`)
**Target Platform**: Linux, macOS, Windows (Docker host)
**Project Type**: Single CLI application with hexagonal architecture
**Performance Goals**: `--project` filter returns results in <2s for 100 containers (SC-003)
**Constraints**: No breaking changes to existing CLI behavior; backward compatible
**Scale/Scope**: Refactoring existing code, ~15 files touched

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Hexagonal Architecture | ✅ PASS | Refactoring maintains domain/ports/adapters separation |
| II. Label-Driven Configuration | ✅ PASS | Unifying validation supports label-first design |
| III. Test-First Development | ✅ PASS | Test de-duplication clarifies responsibilities |
| IV. CLI-First Interface | ✅ PASS | Adding --project/--stack flags, improving help text |
| V. Code Quality & Simplicity | ✅ PASS | Reducing duplication improves maintainability |

**Gate Result**: ✅ PASS - No violations. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/007-milestone-2-5-polish/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI contract updates)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── jobs/
│       ├── types.go          # DefaultSchedule, DefaultWorkerImage, DefaultMountMode
│       └── doc.go            # Stack/project documentation (FR-004)
├── ports/
│   ├── labels.go             # Selector with ProjectFilter, StackFilter (FR-005)
│   └── planner.go            # ValidationError unified type
├── adapters/
│   ├── dockerlabels/
│   │   └── source.go         # Implement ProjectFilter, StackFilter (FR-009)
│   └── joblabels/
│       └── discoverer.go     # Refactor to use shared validation (FR-001-003)
├── config/
│   ├── schema/
│   │   └── job_labels.go     # JobLabelConfig, JobVolumeConfig (source of truth)
│   └── loader/
│       └── job_validation.go # Refactor to use shared validation (FR-001-003)
├── cmd/
│   ├── root.go               # Update branding (FR-010)
│   ├── plan_list.go          # Add --project, --stack flags (FR-006)
│   ├── plan_show.go          # Add --project, --stack flags (FR-007)
│   ├── snapshot.go           # Add --project, --stack flags (FR-008)
│   └── exitcodes.go          # Centralize exit codes (FR-011)
└── testutil/
    ├── doc.go                # Testing philosophy documentation (FR-013)
    └── compose/
        └── *.yaml            # Add header comments (FR-015)

integration/
├── plan_test.go              # Use project filtering (FR-014)
└── joblabels_test.go         # Reduce redundancy with unit tests

docs/
├── config.md                 # Generated docs (FR-017)
└── testing.md                # Testing philosophy (optional, FR-013)
```

**Structure Decision**: Existing hexagonal structure maintained. No new packages needed; this is a refactoring/polish milestone.

## Complexity Tracking

> No constitution violations; this section is empty.
