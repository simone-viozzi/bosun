# Implementation Plan: Milestone 3.75 — Critical Bug Fixes

**Branch**: `011-critical-bug-fixes` | **Date**: 2026-02-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/011-critical-bug-fixes/spec.md`

## Summary

Fix three critical bugs blocking M4: (1) corrupted worker log output caused by raw `io.Copy` on Docker's multiplexed log stream, (2) silent error swallowing and magic exit codes in the worker runner, and (3) coverage file pollution in the repository root. All fixes are localized to existing code — no new packages, no port interface changes.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Docker SDK v28.5.2 (`github.com/docker/docker`), including `pkg/stdcopy` (already a transitive dependency)
**Storage**: N/A
**Testing**: `go test` (unit), Docker Compose-based integration tests (`//go:build integration`)
**Target Platform**: Linux (Docker host)
**Project Type**: Single Go module (`github.com/simone-viozzi/bosun`)
**Performance Goals**: N/A (bug-fix milestone, no performance changes)
**Constraints**: No port interface changes; backward-compatible `WorkerResult` struct
**Scale/Scope**: 3 files modified in production code, 1 in test code, 1 in build config

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Hexagonal Architecture | ✅ PASS | Changes confined to adapter layer (`internal/adapters/docker/worker/`) and test utils. No domain/port changes. |
| II. Label-Driven Configuration | ✅ N/A | No configuration changes. |
| III. Test-First Development | ✅ PASS | Existing unit + integration tests validate the fixes. DumpLogs fix ensures test infrastructure is also correct. |
| IV. CLI-First Interface | ✅ N/A | No CLI changes. Exit codes already defined in `pkg_app_executor` memory. |
| V. Code Quality & Simplicity | ✅ PASS | Replacing magic numbers with named constants; eliminating silent error drops; using correct Docker SDK patterns. |
| VI. Plan-Driven Execution | ✅ N/A | No planner/executor flow changes. |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/011-critical-bug-fixes/
├── plan.md              # This file
├── research.md          # Phase 0 output (below)
├── data-model.md        # Phase 1 output (N/A for bug-fix — no new entities)
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code (files to modify)

```text
internal/adapters/docker/worker/
├── runner.go            # Fix streamLogs (stdcopy), stopContainer (error handling), named exit codes

internal/testutil/
├── docker.go            # Fix DumpLogs (stdcopy)

Makefile                 # Move coverage output to coverage/ subdirectory
.gitignore               # Update gitignore for new coverage paths
```

**Structure Decision**: No new files or packages. All changes are in-place fixes to existing code within the current hexagonal architecture. The coverage directory change is a build-system-only modification.

## Complexity Tracking

No constitution violations — section not applicable.
