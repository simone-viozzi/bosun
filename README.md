# Bosun

> **Bosun is a Docker-label–driven job orchestrator for single hosts that can safely stop stacks, run worker containers with attached resources on a schedule, and start everything back up — with configuration expressed primarily via Docker labels.**

Think of Bosun as: **"cron + docker-compose + labels + safety rails"**.

## Vision

### Goals

* **Label-driven configuration**
  * Primary source of truth is **`bosun.*` labels** on containers, volumes, networks, and compose projects.
  * You can understand "what Bosun will do" by looking at labels and a `bosun config validate` output.

* **Safe orchestration of Docker Compose stacks**
  * Discover stacks and resources via labels and Compose metadata.
  * **Down/Up compose stacks** safely (respecting timeouts, order, and limits).
  * Operate on **a single Docker host** (at least for v1).

* **Generic worker concept**
  * Jobs are executed via **worker containers**:
    * image configurable,
    * volumes/networks attached according to config,
    * command/entrypoint configurable,
    * environment and secrets configurable (later).
  * Bosun is responsible for **wiring** containers and volumes correctly, not for implementing backup logic itself.

* **Scheduling / automation**
  * Cron-like schedules (e.g. `0 3 * * *`) for jobs.
  * Eventually: "run once", "manual trigger", and "on event" hooks.

* **Primary use case (v1)**
  * Automated **backups of Docker Compose stacks**:
    * shut down app stack(s),
    * attach volumes to a worker container,
    * run backup script,
    * bring stack(s) back up,
    * report success/failure.

### Non-goals (at least for now)

* Not a **cluster orchestrator** (no swarm/k8s).
* Not a **full backup solution** (snapshot engines, retention, remote stores) — it **runs your backup container** which does the real work.
* Not a general CI runner or step engine.

## Project Status

Bosun is in active development. See the [Roadmap](#roadmap) section for the planned milestones.

**Current state:**
- ✅ Hexagonal architecture foundation
- ✅ Label discovery for containers, volumes, and networks
- ✅ CLI snapshot command
- ✅ Configuration schema and validation
- ✅ Backup job model and label discovery
- ✅ Job planner with execution plan generation
- ✅ `bosun plan list` and `bosun plan show` CLI commands
- 🚧 Backup execution (Milestone 3 - next)

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker (required for integration tests)
- Make

### Installation

```bash
# Build from source
make build    # Builds to bin/bosun

# Or run directly
make run      # Runs from source
```

### Quick Start: Define a Backup Job

Add Bosun labels to your Docker Compose file to define a backup job:

```yaml
# docker-compose.yaml
services:
  app:
    image: myapp:latest
    labels:
      bosun.job.enabled: "true"
      bosun.job.name: "daily-backup"
      bosun.job.schedule: "0 2 * * *"  # 2 AM daily
    volumes:
      - app-data:/data

volumes:
  app-data:
    labels:
      bosun.job.attach: "daily-backup"      # Job name to attach to
      bosun.job.mount.path: "/backup/app"   # Mount path in worker
      bosun.job.mount.mode: "ro"            # Read-only (default)
```

### Usage

```bash
# Discover and list backup jobs
bosun plan list

# Show execution plan for a job
bosun plan show daily-backup

# Validate all labels (config + job)
bosun config validate

# View all Docker entities with bosun.* labels
bosun labels snapshot

# Include stopped containers
bosun plan list --stopped
```

## Job Labels Reference

### Container Labels

| Label | Type | Description |
|-------|------|-------------|
| `bosun.job.enabled` | boolean | Enable job participation (required) |
| `bosun.job.name` | string | Unique job identifier (required when enabled) |
| `bosun.job.schedule` | cron | Cron expression (default: `0 0 * * *`) |
| `bosun.job.worker.image` | string | Worker container image |

### Volume Labels

| Label | Type | Description |
|-------|------|-------------|
| `bosun.job.attach` | boolean | Attach volume to backup worker |
| `bosun.job.name` | string | Job this volume belongs to |
| `bosun.job.mountpath` | string | Mount path in worker (default: `/backup/<volume-name>`) |
| `bosun.job.readonly` | boolean | Mount as read-only (default: true) |

See [docs/config.md](docs/config.md) for the complete configuration reference.

## Roadmap

Bosun development is organized into six milestones:

### Milestone 1 – Label-driven Configuration v1
Treat `bosun.*` labels as a real schema with strict validation and good tooling.
- Code-first config schema with rich struct tags
- Label parser with strict validation (fail on unknown keys)
- Config merger (defaults < file < env < labels)
- `bosun config validate` CLI command
- Auto-generated documentation and JSON Schema

### Milestone 2 – Backup Job Model & Planning
Introduce a first-class "backup job" concept and planner (no side effects yet).
- `BackupJob` and `BackupPlan` domain models
- Label schema for describing backup jobs
- Pure, testable planner that produces executable plans
- `bosun plan list` / `bosun plan show` CLI commands

### Milestone 3 – Backup Execution MVP
Execute backup plans on demand, correctly orchestrating Docker resources.
- Compose control adapter (down/up with timeouts)
- Worker container execution with volume attachment
- Orchestration logic (stop → backup → start)
- `bosun job run <name>` CLI command
- Safety features (timeouts, dry-run mode)

### Milestone 4 – Scheduling Engine & Runtime
Add a cron-like scheduler for periodic job execution.
- In-process scheduler with cron expression support
- Long-lived runtime process
- Job concurrency limits
- `bosun job list` CLI with schedule status

### Milestone 5 – Observability, UX & Production Hardening
Make Bosun pleasant and safe to run in production.
- Structured logging with job/plan IDs
- Status introspection (recent runs, failures)
- Crash-safe job state persistence
- Polished documentation and examples

### Milestone 6 – v1.0: Backup-focused Release
Ship a focused v1 that does one thing well: automated, scheduled backups of Docker Compose stacks using configurable worker containers.

## Architecture

Bosun follows **hexagonal/clean architecture** principles:

```
cmd/bosun/           # Application entry point
internal/
├── domain/          # Business logic entities and rules
├── ports/           # Interface definitions for external dependencies
├── adapters/        # Concrete implementations
│   ├── docker/      # Docker client utilities
│   ├── dockerlabels/# Label discovery implementation
│   ├── http/        # HTTP adapters (future)
│   └── storage/     # Storage adapters (future)
└── app/             # Application wiring and orchestration
```

## Testing

Bosun includes comprehensive unit and integration tests.

```bash
make test     # Run unit tests
make it       # Run integration tests
make itv      # Run integration tests (verbose)
```

**Prerequisites for integration tests:**
- Docker must be installed and running
- Integration tests use the `integration` build tag (`//go:build integration`)

## Development

### Code Quality

```bash
make fmt      # Format code
make vet      # Run go vet
make tidy     # Tidy dependencies
```

### Project Conventions

- **Commit Messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/)
- **Branching**: Semi-linear history enforced - rebase PRs, no merge commits
- **Testing**: Integration tests use build tag `//go:build integration`

## Contributing

Contributions are welcome! Please ensure all tests pass and follow the project conventions before submitting a PR.

## License

See [LICENSE](LICENSE) file for details.
