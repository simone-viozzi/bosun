# Project Status & Roadmap

## Current Status: Pre-Alpha

Bosun does not work end-to-end yet. Core components are being developed.

## What Bosun Is

**A job orchestrator for Docker environments.**

Given Docker Compose stacks with `bosun.*` labels, Bosun can:
- Discover job definitions from labels
- Generate execution plans
- (Future) Stop stacks, start worker containers with volumes attached, restart stacks

**Not a backup tool.** Backups are one use case. Workers can do anything: maintenance, migrations, health checks, data processing.

## Milestone Status

### ✅ Milestone 1: Label-driven Configuration v1
- Code-first schema with validation
- `bosun config validate` CLI
- Auto-generated docs

### 🟡 Milestone 2: Job Model & Planning (partial)
- Job and ExecutionPlan domain models ✅
- Planner service ✅
- `bosun plan` CLI ✅
- Needs polish per #78 (CI ordering)

### ❌ Milestone 3: Job Execution MVP (#85)
- Compose control (down/up) - research needed (#109)
- Worker container execution
- `bosun job run` CLI

### ❌ Milestone 4: Scheduling Engine (#86)
- Cron-based scheduler
- Concurrency controls - research needed (#108)
- `bosun daemon` mode

### ❌ Milestone 5: Observability (#87)
- Structured logging
- Status introspection
- Worker signals - research needed (#110)

### ❌ Milestone 6: v1.0 Release (#88)
- All features complete
- Documentation
- Release artifacts

## Open Research Items

- **#108**: Job concurrency strategy (semaphore, parallelism)
- **#109**: Compose control strategy (Docker API vs CLI)
- **#110**: Worker architecture (signals, base images, examples)

## Tech Stack

- **Language**: Go 1.24.9
- **Architecture**: Hexagonal (Ports & Adapters)
- **CLI**: Cobra
- **Testing**: Integration tests with Docker Compose
- **CI**: GitHub Actions, golangci-lint, Codecov
