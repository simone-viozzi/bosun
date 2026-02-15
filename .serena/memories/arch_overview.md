# Architecture Overview

## Scope
High-level architecture of Bosun: hexagonal design, package structure, and core principles.

## What
Bosun is a **job orchestrator for Docker environments**. It discovers job definitions from Docker labels, generates execution plans, and orchestrates worker containers with attached volumes.

**Not a backup tool**: While backups are a common use case, Bosun is a generic orchestrator. Workers can perform any task: backups, maintenance, migrations, health checks, etc.

### Hexagonal Architecture (Ports & Adapters)
- **Domain** (`internal/domain/`): Core business logic, no external dependencies
  - `labels/`: Label types and constants
  - `jobs/`: Job, ExecutionPlan, PlanStep types
- **Ports** (`internal/ports/`): Interface contracts
  - `LabelSource`: Discovers labeled Docker entities
  - `JobDiscoverer`: Transforms labels into Jobs
  - `JobPlanner`: Generates ExecutionPlans from Jobs
  - `JobExecutor`: Orchestrates job execution (plan-driven)
  - `ComposeController`: Stack lifecycle management
  - `WorkerRunner`: Worker container execution
- **Adapters** (`internal/adapters/`): External system integrations
  - `dockerlabels/`: Docker API label discovery
  - `joblabels/`: Job parsing from labels
  - `docker/compose/`: Compose stack control (stop/start)
  - `docker/worker/`: Worker container lifecycle
  _(Empty placeholder directories like `http/`, `storage/` have been removed; add new adapters here as needed.)_
- **Application** (`internal/app/`): Orchestration layer
  - `planner/`: Plan generation service
  - `executor/`: Plan-driven job execution
- **Config** (`internal/config/`): Configuration system
  - `schema/`: Code-first schema definitions
  - `loader/`: Label parsing and validation
  - `merge/`: Multi-source config merging
- **CLI** (`internal/cmd/`): Command implementations

### Key Data Flow
```
Docker Labels → LabelSource.Snapshot() → JobDiscoverer.DiscoverJobs()
→ JobPlanner.Plan() → ExecutionPlan → JobExecutor.Execute()
    │
    └─► StopContainers → RunWorker → StartContainers
```

## Why
- **Testability**: Mock adapters for unit tests; real Docker for integration tests
- **Technology agnostic**: Can swap implementations (e.g., Kubernetes adapter later)
- **Maintainable**: Clear separation prevents coupling

## Related
- `arch_development_lifecycle` - Development commands and workflows
- `pkg_domain_jobs` - Job and ExecutionPlan details
- `pkg_ports` - Port interface contracts
