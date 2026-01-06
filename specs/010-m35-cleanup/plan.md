# Implementation Plan: Milestone 3.5 - Post-M3 Cleanup & Bug Fixes

**Branch**: `010-m35-cleanup` | **Date**: 2026-01-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-m35-cleanup/spec.md`

## Summary

Address critical bugs and incomplete tasks discovered after M3 (Job Execution MVP) completion:
- **Config Validation**: Change from strict to lenient unknown-key handling (warn, don't error)
- **Execution Plan**: Make `ExecutionPlan` authoritative — planner adds start step, executor interprets plan
- **Terminology**: Rename `BackupEnabled` → `Enabled`, remove "backup" references from production code
- **Interface**: Fix `JobExecutor` interface mismatch, remove unused parameters
- **CLI**: Fix duplicate error messages
- **Docs**: Add "Running Jobs" section to README

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI), Docker SDK, testcontainers-go (testing)
**Storage**: N/A (stateless, reads Docker labels)
**Testing**: `go test` (unit), `go test -tags=integration` (integration)
**Target Platform**: Linux (primary), macOS (development)
**Project Type**: Single Go module with hexagonal architecture
**Performance Goals**: N/A (cleanup milestone, no new performance requirements)
**Constraints**: Breaking changes to config labels acceptable (pre-1.0)
**Scale/Scope**: ~15 files modified, 6 user stories, 16 functional requirements

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Hexagonal Architecture** | ✅ PASS | Changes respect layers: domain, ports, adapters, app |
| **II. Label-Driven Configuration** | ⚠️ DEVIATION | FR-001 changes unknown keys from error to warning. **Justified**: Decision #5 in `wip_smell_milestone3.md` — lenient mode improves UX for rolling deployments and custom labels. `--strict` flag preserves original behavior. |
| **III. Test-First Development** | ✅ PASS | Existing tests updated, new unit tests for executor step interpreter |
| **IV. CLI-First Interface** | ✅ PASS | All changes accessible via CLI (`--strict` flag, updated output) |
| **V. Code Quality & Simplicity** | ✅ PASS | Cleanup reduces complexity (executor refactoring, terminology consistency) |

**Gate Result**: ✅ PASS (one documented deviation with justification)

## Project Structure

### Documentation (this feature)

```text
specs/010-m35-cleanup/
├── plan.md              # This file
├── research.md          # Phase 0 output (minimal — decisions already made)
├── data-model.md        # Phase 1 output (changes to existing types)
├── quickstart.md        # Phase 1 output (verification commands)
├── contracts/           # Phase 1 output (updated interface contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── app/
│   ├── executor/        # US4: Plan interpreter, interface cleanup
│   │   └── executor.go
│   └── planner/         # US2: Add start_containers step
│       └── planner.go
├── config/
│   ├── loader/          # US1: Lenient unknown-key handling
│   │   └── loader.go
│   └── schema/          # US3: BackupEnabled → Enabled
│       └── config_v1.go
├── cmd/                 # US5: Duplicate error fix
│   ├── root.go
│   ├── job_run.go
│   ├── plan_list.go
│   ├── plan_show.go
│   └── validate.go
├── domain/
│   └── jobs/            # US2/US4: ExecutionPlan types
│       └── types.go
└── ports/               # US4: JobExecutor interface
    └── executor.go

docs/
└── config.md            # US3: Terminology update

README.md                # US6: Running Jobs section

integration/             # Updated integration tests
└── *.go
```

**Structure Decision**: Existing hexagonal structure preserved. Changes are modifications to existing files, not new packages.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Constitution II deviation (lenient validation) | Improves UX for rolling deployments, custom labels, and extension-friendliness | Strict-only validation breaks on new bosun versions with new labels; `--strict` flag preserves original behavior for users who need it |
