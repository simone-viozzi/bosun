# Implementation Plan: Job Model & Planning

**Branch**: `006-backup-job-model` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-backup-job-model/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Introduce first-class `Job` and `ExecutionPlan` domain entities to Bosun. Jobs are defined via Docker labels (`bosun.job.*`) on containers and volumes. A pure-function planner transforms a Snapshot + Config into an ExecutionPlan (ordered steps for stop/run-worker). This milestone introduces no Docker side effects—only discovery, validation, and plan generation.

## Technical Context

**Language/Version**: Go 1.24.9
**Primary Dependencies**: `github.com/docker/docker` (Docker SDK), `github.com/spf13/cobra` (CLI), `github.com/docker/compose/v2` (indirect), cron parsing library (TBD in research)
**Storage**: N/A (stateless discovery from Docker API)
**Testing**: Go testing (`go test`), testcontainers-go for integration tests, `make test` / `make it`
**Target Platform**: Linux (Docker host), macOS, Windows (with Docker)
**Project Type**: Single project (hexagonal architecture)
**Performance Goals**: `plan list` <5s for 100 containers; `plan show` <2s for 20 volumes
**Constraints**: No Docker side effects (stop/start); deterministic planner output
**Scale/Scope**: Up to 100 containers, 50 volumes per environment

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Hexagonal Architecture** | ✅ PASS | Domain in `internal/domain/jobs/`, Port interface `JobPlanner` in `internal/ports/`, Adapter for label parsing in `internal/adapters/` |
| **II. Label-Driven Configuration** | ✅ PASS | All job config via `bosun.job.*` labels on containers/volumes; schema-first with `FieldSpec` types; strict validation (no unknown keys) |
| **III. Test-First Development** | ✅ PASS | Unit tests for domain/planner, integration tests with Docker Compose stacks |
| **IV. CLI-First Interface** | ✅ PASS | `bosun plan list` and `bosun plan show <job>` commands; JSON/YAML output for scripting |
| **V. Code Quality & Simplicity** | ✅ PASS | Standard Go formatting; minimal dependencies; YAGNI (no executor in this milestone) |

**Gate Result**: ✅ All principles satisfied. No violations to justify.

### Post-Design Re-check (Phase 1 Complete)

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Hexagonal Architecture** | ✅ PASS | data-model.md defines pure domain types; contracts/planner.md defines port interfaces |
| **II. Label-Driven Configuration** | ✅ PASS | research.md §4 defines schema; all config via `bosun.job.*` labels |
| **III. Test-First Development** | ✅ PASS | Contracts specify testable behaviors; integration test compose file planned |
| **IV. CLI-First Interface** | ✅ PASS | contracts/cli.md defines `plan list` and `plan show` with all flags/formats |
| **V. Code Quality & Simplicity** | ✅ PASS | Single dependency added (robfig/cron/v3); no over-engineering |

**Post-Design Gate**: ✅ PASS - Design artifacts align with constitution.

## Project Structure

### Documentation (this feature)

```text
specs/006-backup-job-model/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── planner.md       # Planner interface contract
│   └── cli.md           # CLI command contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Hexagonal Architecture Layout (existing)
internal/
├── domain/
│   ├── labels/          # Existing: LabeledEntity, Snapshot, Kind
│   └── jobs/            # NEW: Job, ExecutionPlan, PlanStep, Stack
│       └── types.go
├── ports/
│   ├── labels.go        # Existing: LabelSource, Selector
│   └── planner.go       # NEW: JobPlanner interface
├── adapters/
│   ├── dockerlabels/    # Existing: label discovery
│   └── joblabels/       # NEW: job label parsing adapter
│       ├── parser.go
│       └── parser_test.go
├── config/
│   └── schema/
│       └── job_labels.go # NEW: Job label schema (FieldSpec)
├── cmd/
│   ├── root.go          # Existing
│   ├── labels.go        # Existing
│   ├── plan.go          # NEW: plan command group
│   ├── plan_list.go     # NEW: plan list subcommand
│   └── plan_show.go     # NEW: plan show subcommand
└── app/
    └── planner/         # NEW: Pure planner implementation
        ├── planner.go
        └── planner_test.go

integration/
├── plan_test.go         # NEW: Integration tests for plan commands
└── joblabels_compose.yaml # NEW: Test compose with job labels
```

**Structure Decision**: Follows existing hexagonal layout. New `jobs` domain parallels existing `labels` domain. Planner lives in `internal/app/planner/` as application logic orchestrating domain objects.

## Complexity Tracking

> **No violations detected. Table intentionally empty.**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
