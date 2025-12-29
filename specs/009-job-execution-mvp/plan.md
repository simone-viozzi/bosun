# Implementation Plan: Job Execution MVP (Milestone 3)

**Branch**: `009-job-execution-mvp` | **Date**: 2025-12-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-job-execution-mvp/spec.md`

## Summary

Bosun M3 implements job execution orchestration: stopping Compose stacks, running worker containers with attached volumes, and restarting stacks. Uses direct Docker API with label-based topology discovery (no Compose CLI/library). Worker containers follow BYOI pattern with BOSUN_* environment variables and SIGTERM-based timeouts.

**Key Technical Decisions** (from research):
- **Compose Control**: Docker API + `com.docker.compose.*` labels + topological sort (#109)
- **Worker Contract**: BOSUN_* env vars, SIGTERM→10s→SIGKILL, exit codes only (#110)
- **Failure Handling**: 30s stop/start timeouts, always restart stack, pre-validate images (#117)

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- `github.com/docker/docker` - Docker SDK for container control
- `github.com/spf13/cobra` - CLI framework (existing)
- `github.com/google/uuid` - Run ID generation

**Storage**: N/A (stateless execution; log persistence deferred to M5)
**Testing**:
- `go test` for unit tests
- `testcontainers-go` with Docker Compose for integration tests
- `//go:build integration` tag convention

**Target Platform**: Linux (Docker host environments)
**Project Type**: Single Go module with hexagonal architecture
**Performance Goals**: Execute simple job (stop→worker→start) within 5 minutes
**Constraints**:
- 30s default timeout for stop/start operations
- 1h default timeout for worker execution
- No health check waiting in M3 (container "started" ≠ healthy)

**Scale/Scope**: Single job execution at a time (no concurrency/locking in M3)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Hexagonal Architecture** | ✅ PASS | New ports in `internal/ports/`, adapters in `internal/adapters/docker/`, service in `internal/app/` |
| **II. Label-Driven Configuration** | ✅ PASS | Uses existing `bosun.*` labels; new labels for timeouts defined in schema |
| **III. Test-First Development** | ✅ PASS | Unit tests for each component; integration tests with real Docker (#123) |
| **IV. CLI-First Interface** | ✅ PASS | `bosun job run` command with `--dry-run`, `--timeout`, `--format` flags |
| **V. Code Quality & Simplicity** | ✅ PASS | golangci-lint passes; follows existing patterns |

**Dependency Direction Check**:
```
adapters/docker/compose → ports.ComposeController (interface) → domain/jobs
adapters/docker/worker  → ports.WorkerRunner (interface)     → domain/jobs
app/executor            → ports.* (all interfaces)           → domain/jobs
cmd/job_run             → app/executor                       → ports.*
```

All dependencies point inward. ✅

## Project Structure

### Documentation (this feature)

```text
specs/009-job-execution-mvp/
├── spec.md              # Feature specification (input)
├── plan.md              # This file
├── research.md          # Phase 0 output - consolidated decisions
├── data-model.md        # Phase 1 output - domain types & interfaces
├── quickstart.md        # Phase 1 output - implementation guide
└── contracts/           # Phase 1 output - Go interface definitions
    ├── compose_controller.go
    ├── worker_runner.go
    └── job_executor.go
```

### Source Code (repository root)

```text
internal/
├── ports/
│   ├── labels.go           # (existing) LabelSource interface
│   ├── planner.go          # (existing) JobDiscoverer, JobPlanner interfaces
│   ├── compose.go          # (new #115) ComposeController interface
│   ├── worker.go           # (new #116) WorkerRunner interface
│   └── executor.go         # (new #114) JobExecutor interface
│
├── domain/jobs/
│   ├── types.go            # (existing) Job, ExecutionPlan, PlanStep
│   ├── run.go              # (new) JobRun, ExecutionResult, RunStatus
│   └── errors.go           # (new) StopError, StartError, TimeoutError
│
├── adapters/docker/
│   ├── compose/            # (new #118) ComposeController adapter
│   │   ├── controller.go   # Implements ComposeController
│   │   ├── topology.go     # Topological sort for dependencies
│   │   └── controller_test.go
│   │
│   └── worker/             # (new #119) WorkerRunner adapter
│       ├── runner.go       # Implements WorkerRunner
│       └── runner_test.go
│
├── app/
│   ├── planner/            # (existing) Plan generation
│   └── executor/           # (new #121) Executor service
│       ├── executor.go     # Orchestrates job execution
│       └── executor_test.go
│
└── cmd/
    ├── exitcodes.go        # (extend) New exit codes for M3
    └── job_run.go          # (new #122) `bosun job run` command

integration/
├── job_execution_test.go   # (new #123) Full job execution tests
└── worker_test.go          # Worker container tests
```

**Structure Decision**: Follows existing hexagonal architecture. New adapters in `internal/adapters/docker/` with subdirectories for compose control and worker execution. New service in `internal/app/executor/`.

## Complexity Tracking

> **No Constitution violations. GATE PASSED.**

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  cmd/job_run.go                                                      │    │
│  │  - `bosun job run <job-name>`                                        │    │
│  │  - Flags: --dry-run, --timeout, --format, --quiet, --keep-stopped   │    │
│  └────────────────────────────────┬────────────────────────────────────┘    │
└───────────────────────────────────┼─────────────────────────────────────────┘
                                    │
┌───────────────────────────────────┼─────────────────────────────────────────┐
│                           Application Layer                                  │
│  ┌────────────────────────────────▼────────────────────────────────────┐    │
│  │  app/executor/executor.go                                            │    │
│  │  - Orchestrates: discover → validate → stop → run → start           │    │
│  │  - Handles Ctrl+C: always attempt restart before exit               │    │
│  │  - Dry-run mode: return plan without execution                      │    │
│  └──────────┬───────────────────┬─────────────────────┬────────────────┘    │
└─────────────┼───────────────────┼─────────────────────┼─────────────────────┘
              │                   │                     │
┌─────────────┼───────────────────┼─────────────────────┼─────────────────────┐
│             │           Ports (Interfaces)            │                      │
│  ┌──────────▼──────────┐ ┌──────▼──────────┐ ┌───────▼─────────┐           │
│  │ ComposeController   │ │ WorkerRunner    │ │ JobExecutor     │           │
│  │ (#115)              │ │ (#116)          │ │ (#114)          │           │
│  │                     │ │                 │ │                 │           │
│  │ - StopStack()       │ │ - Run()         │ │ - Execute()     │           │
│  │ - StartStack()      │ │                 │ │ - DryRun()      │           │
│  │ - ListContainers()  │ │                 │ │                 │           │
│  └──────────┬──────────┘ └────────┬────────┘ └─────────────────┘           │
└─────────────┼─────────────────────┼─────────────────────────────────────────┘
              │                     │
┌─────────────┼─────────────────────┼─────────────────────────────────────────┐
│             │     Adapters (Implementations)          │                      │
│  ┌──────────▼──────────┐ ┌────────▼────────┐                                │
│  │ docker/compose/     │ │ docker/worker/  │                                │
│  │ (#118)              │ │ (#119)          │                                │
│  │                     │ │                 │                                │
│  │ - Docker SDK        │ │ - Docker SDK    │                                │
│  │ - Label discovery   │ │ - ContainerRun  │                                │
│  │ - Topological sort  │ │ - Log capture   │                                │
│  │ - Stop/Start        │ │ - Timeout/kill  │                                │
│  └─────────────────────┘ └─────────────────┘                                │
└─────────────────────────────────────────────────────────────────────────────┘
              │                     │
              └──────────┬──────────┘
                         │
              ┌──────────▼──────────┐
              │     Docker API      │
              │  (Docker daemon)    │
              └─────────────────────┘
```

## Dependency Graph

```
                    ┌──────────────┐
                    │ #122 CLI     │
                    │ job run      │
                    └──────┬───────┘
                           │ depends on
                           ▼
                    ┌──────────────┐
                    │ #121 Executor│
                    │ service      │
                    └──┬───────┬───┘
          depends on   │       │   depends on
           ┌───────────┘       └───────────┐
           ▼                               ▼
    ┌──────────────┐               ┌──────────────┐
    │ #118 Compose │               │ #119 Worker  │
    │ adapter      │               │ adapter      │
    └──────┬───────┘               └──────┬───────┘
           │ implements                   │ implements
           ▼                              ▼
    ┌──────────────┐               ┌──────────────┐
    │ #115 Compose │               │ #116 Worker  │
    │ port         │               │ port         │
    └──────────────┘               └──────────────┘
           │                              │
           └───────────┬──────────────────┘
                       │ uses
                       ▼
    ┌──────────────────────────────────────────────┐
    │ #114 JobExecutor port                        │
    │ (orchestration interface)                    │
    └──────────────────────────────────────────────┘

    Integration Tests (#123) depend on all above
```

## Implementation Order

### Phase 1: Ports (Parallel)

| Issue | Component | Deliverable | Est. |
|-------|-----------|-------------|------|
| #114 | JobExecutor port | `internal/ports/executor.go` | 0.5d |
| #115 | ComposeController port | `internal/ports/compose.go` | 0.5d |
| #116 | WorkerRunner port | `internal/ports/worker.go` | 0.5d |

All three ports can be implemented in parallel since they have no dependencies on each other.

### Phase 2: Adapters (Sequential)

| Issue | Component | Deliverable | Est. | Depends |
|-------|-----------|-------------|------|---------|
| #118 | ComposeController adapter | `internal/adapters/docker/compose/` | 2d | #115 |
| #119 | WorkerRunner adapter | `internal/adapters/docker/worker/` | 1.5d | #116 |

Adapters can be developed in parallel after their respective ports are defined.

### Phase 3: Application & CLI

| Issue | Component | Deliverable | Est. | Depends |
|-------|-----------|-------------|------|---------|
| #121 | Executor service | `internal/app/executor/` | 2d | #114, #118, #119 |
| #122 | CLI command | `internal/cmd/job_run.go` | 1d | #121 |

### Phase 4: Testing & Docs

| Issue | Component | Deliverable | Est. | Depends |
|-------|-----------|-------------|------|---------|
| #123 | Integration tests | `integration/job_execution_test.go` | 2d | #122 |
| #120 | Documentation | README updates, examples | 1d | #122 |

## Integration Points with M2

### Existing Components to Use

| M2 Component | Location | How M3 Uses It |
|--------------|----------|----------------|
| `Job` type | `internal/domain/jobs/types.go` | Input to executor |
| `ExecutionPlan` | `internal/domain/jobs/types.go` | Steps from planner |
| `JobDiscoverer` | `internal/ports/planner.go` | Find job by name |
| `JobPlanner` | `internal/ports/planner.go` | Generate execution plan |
| `LabelSource` | `internal/ports/labels.go` | Discover containers |
| `joblabels.Discoverer` | `internal/adapters/joblabels/` | Parse job labels |
| `dockerlabels.Source` | `internal/adapters/dockerlabels/` | Docker label access |

### New Labels to Add

| Label | Type | Default | Package |
|-------|------|---------|---------|
| `bosun.backup.stop-timeout` | duration | `30s` | `internal/config/schema/` |
| `bosun.backup.start-timeout` | duration | `30s` | `internal/config/schema/` |
| `bosun.backup.worker-env.*` | string | - | `internal/config/schema/` |

### Exit Codes to Add

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | `ExitSuccess` | (existing) Job succeeded |
| 10 | `ExitWorkerFailed` | Worker exited non-zero |
| 11 | `ExitStopFailed` | Failed to stop stack |
| 12 | `ExitStartFailed` | Failed to restart stack |
| 13 | `ExitTimeout` | Operation timed out |
| 14 | `ExitImageNotFound` | Worker image not found |
| 15 | `ExitJobNotFound` | Job name not found |
| 16 | `ExitInterrupted` | Execution interrupted (Ctrl+C) |

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Docker API version differences | Pin minimum Docker API version; test on multiple versions |
| Topological sort complexity | Use proven DFS algorithm from Watchtower research |
| Signal handling edge cases | Comprehensive integration tests with hanging workers |
| Concurrent job execution | Document limitation; add warning if detected (future: M4 locking) |
| Partial stack restart | Always attempt full restart; report partial failures clearly |

## Success Metrics

From spec SC-001 through SC-006:

- [ ] `bosun job run <job>` executes full stop→worker→start cycle in <5 min
- [ ] `--dry-run` output matches actual execution steps
- [ ] Exit code 0 ↔ success; non-zero ↔ failure with code reported
- [ ] Worker logs visible during/after execution
- [ ] Hung workers terminated after timeout
- [ ] Integration tests pass with real Docker Compose stacks

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
