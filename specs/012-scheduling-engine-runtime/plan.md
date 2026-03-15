# Implementation Plan: Scheduling Engine & Runtime

**Branch**: `012-scheduling-engine-runtime` | **Date**: 2026-02-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-scheduling-engine-runtime/spec.md`

## Summary

Implement a long-lived daemon (`bosun daemon`) that automatically discovers backup/maintenance jobs from Docker labels, schedules them via cron expressions, and executes them through the existing `JobExecutor` with a **three-layer concurrency model**: per-job overlap policies (Layer 1), a global semaphore for parallelism control (Layer 2), and per-stack mutexes preventing concurrent access to the same Compose stack (Layer 3). The design introduces an app-level service factory (`Bootstrap`) to resolve CLI→adapter coupling (#141), an `EventEmitter` port for lifecycle observability, and periodic config refresh for hot-reconfiguration without daemon restart.

Technical approach derived from #108 research: leverage `robfig/cron/v3` (already in go.mod) for scheduling with built-in overlap wrappers, `golang.org/x/sync/semaphore` for the global concurrency limit, and `sync.Mutex`-based `StackLockManager` with sorted multi-stack lock acquisition to prevent deadlocks. Circuit-breaker logic auto-disables jobs after 3 consecutive failures.

## Technical Context

**Language/Version**: Go 1.24+ (module `github.com/simone-viozzi/bosun`)
**Primary Dependencies**:
- `robfig/cron/v3 v3.0.1` — cron scheduling + overlap policy wrappers (`DelayIfStillRunning`, `SkipIfStillRunning`)
- `golang.org/x/sync v0.19.0` — `semaphore.Weighted` for global concurrency control
- `github.com/docker/docker` — Docker SDK for label discovery and container control
- `github.com/spf13/cobra` — CLI framework
- `log/slog` — structured logging (Go stdlib)

**Storage**: N/A (all state in-memory for M4; persistent scheduling deferred to #177)
**Testing**: `go test` + `testcontainers-go` for integration tests in `integration/`; `//go:build integration` tag; race detector enabled
**Target Platform**: Linux (Docker host, single-node deployment)
**Project Type**: Single Go module, hexagonal architecture
**Performance Goals**: <10s scheduling jitter, <1s `bosun job list` response time, 72h+ stable daemon operation
**Constraints**: <50MB memory growth over 24h, graceful shutdown within 60s, no external state dependencies
**Scale/Scope**: Typical deployment: 5–50 scheduled jobs across 5–20 Compose stacks on a single host

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Hexagonal Architecture — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Domain Independence | ✅ | `OverlapPolicy`, `JobStatus`, `RunStatus` are pure domain types in `internal/domain/jobs/` with no imports from adapters |
| Ports Define Contracts | ✅ | `EventEmitter` interface defined in `internal/ports/executor.go`; `Scheduler` consumes ports only |
| Adapters Implement Ports | ✅ | `LogEmitter` in `internal/adapters/events/` implements `EventEmitter` port |
| Dependency Direction | ✅ | `cmd/` → `app/` → `ports/` → `domain/`; no reverse imports |
| CLI is Thin | ✅ | `daemon.go` and `job_list.go` parse flags, call `app.Bootstrap()`, delegate to `Scheduler` |

### II. Label-Driven Configuration — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Primary Source | ✅ | `bosun.job.schedule`, `bosun.job.overlap-policy`, `bosun.job.enabled` labels drive scheduling |
| Schema-First | ✅ | New labels registered in `internal/config/schema/` with type, scope, and documentation |
| Lenient Validation | ✅ | Unknown keys warn by default; `--strict` for errors (existing behavior) |
| Precedence | ✅ | Labels remain authoritative; no file-based config override in M4 |

### III. Test-First Development — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Unit Tests | ✅ | Scheduler, StackLockManager, LogEmitter, OverlapPolicy each have unit test files |
| Integration Tests | ✅ | `integration/scheduling_test.go` with real Docker daemon, accelerated cron, race detector |
| Test Compose Files | ✅ | Fixtures in `internal/testutil/compose/` |
| Coverage | ✅ | Core concurrency and scheduling paths covered; trivial adapters minimal coverage |

### IV. CLI-First Interface — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Cobra Framework | ✅ | `bosun daemon` and `bosun job list` commands via Cobra |
| Output Protocol | ✅ | `job list` supports `--format text\|json\|yaml`; daemon logs to stderr |
| Flags Over Prompts | ✅ | `--parallelism`, `--refresh-interval`, `--format` flags; no interactive prompts |
| Exit Codes | ✅ | Existing 10-16 range for job errors; new daemon exit codes TBD (0=clean, 1=error) |

### V. Code Quality & Simplicity — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Go Standards | ✅ | `go fmt`, `go vet`, `golangci-lint` enforced via pre-commit and CI |
| YAGNI | ✅ | Cancel-and-restart deferred (#176); persistent scheduling deferred (#177); notification adapters deferred |
| No Magic | ✅ | Explicit sorted lock ordering; explicit circuit-breaker threshold (3 failures); no hidden state |

### VI. Plan-Driven Execution — ✅ PASS

| Principle | Compliance | Evidence |
|-----------|-----------|----------|
| Planner Generates Plan | ✅ | Existing `JobPlanner.Plan()` produces `ExecutionPlan` (unchanged) |
| Executor Interprets Plan | ✅ | Scheduler calls `Executor.Execute()` which follows plan steps (unchanged from M3) |
| DryRun Matches Execute | ✅ | `bosun job run --dry-run` still matches actual execution |
| Always Restart | ✅ | Stack restarted even on worker failure (unchanged from M3) |

**Gate Result**: ✅ All 6 constitution principles PASS. No violations requiring justification.

## Project Structure

### Documentation (this feature)

```text
specs/012-scheduling-engine-runtime/
├── plan.md              # This file
├── research.md          # Three-layer concurrency model research from #108
├── data-model.md        # Domain types, ports, adapters, data flow
├── quickstart.md        # Week-by-week implementation guide with code snippets
├── contracts/           # Go interface contracts
│   ├── event_emitter.go
│   ├── job_state_store.go
│   ├── scheduler.go
│   └── stack_lock_manager.go
└── tasks.md             # Task breakdown mapped to GitHub issues #168-#176
```

### Source Code (repository root)

```text
internal/
├── domain/jobs/
│   ├── types.go            [EXTENDED] Job struct + Schedule, OverlapPolicy, Enabled fields
│   ├── status.go           [NEW]     JobStatus, RunStatus types
│   └── overlap.go          [NEW]     OverlapPolicy type + ValidateOverlapPolicy()
│
├── ports/
│   ├── executor.go         [EXTENDED] Add EventEmitter interface
│   └── state.go            [NEW]     JobStateStore interface + JobState type
│
├── app/
│   ├── app.go              [EXTENDED] Services struct + Bootstrap() factory (#141)
│   ├── scheduler/          [NEW]
│   │   ├── scheduler.go    [NEW]     Scheduler struct: cron, refresh loop, status tracking
│   │   └── refresh.go      [NEW]     Config refresh: diff-based add/remove/update logic
│   ├── concurrency/        [NEW]
│   │   └── stack_lock.go   [NEW]     StackLockManager with sorted multi-stack locking
│   └── executor/
│       └── executor.go     [EXTENDED] Inject EventEmitter, emit lifecycle events
│
├── adapters/
│   ├── events/             [NEW]
│   │   └── log_emitter.go  [NEW]     LogEmitter: structured slog-based EventEmitter
│   └── state/              [NEW]
│       └── memory.go       [NEW]     InMemoryStateStore: sync.Map-backed, no durability
│
├── config/schema/
│   └── job_labels.go       [EXTENDED] Add schedule, overlap-policy, enabled labels
│
└── cmd/
    ├── daemon.go           [NEW]     bosun daemon --parallelism --refresh-interval
    ├── job_list.go         [NEW]     bosun job list --format text|json|yaml
    └── root.go             [EXTENDED] Register daemon and job list sub-commands

integration/
├── scheduling_test.go      [NEW]     Scheduler integration with real Docker
├── concurrency_test.go     [NEW]     Stack locks + global semaphore under load
└── daemon_test.go          [NEW]     Daemon lifecycle, signal handling, config refresh
```

**Structure Decision**: Single Go module, hexagonal architecture. New code follows existing package conventions — domain types in `internal/domain/jobs/`, port interfaces in `internal/ports/`, app-layer orchestration in `internal/app/`, adapters in `internal/adapters/`, CLI commands in `internal/cmd/`. Two new app-layer packages (`scheduler/`, `concurrency/`) contain the core scheduling engine. A `JobStateStore` port with `InMemoryStateStore` adapter is included from M4 to ensure #177 (persistent scheduling) is a drop-in adapter swap rather than a scheduler refactor.

## Complexity Tracking

> No constitution violations — no justifications required.

*N/A*
