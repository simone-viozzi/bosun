# Docker Compose Adapter

## Scope
Compose stack lifecycle control in `internal/adapters/docker/compose/`.

## What
Implements `ports.ComposeController` using Docker API + labels (not Compose CLI).

### Approach: API + Labels + Topological Sort
- **Stack identification**: Filter by `com.docker.compose.project` label
- **Dependency order**: Parse `com.docker.compose.depends_on` label + topological sort
- **Container control**: Docker API `ContainerStop`/`ContainerStart`

### Key Files
- `controller.go` - Main `Controller` struct implementing `ComposeController`
- `dependencies.go` - Topological sort for container ordering
- `doc.go` - Package documentation

### Dependency Ordering
Uses DFS-based topological sort (pattern from Watchtower):
- **Stop**: Reverse order (dependents first, then dependencies)
- **Start**: Forward order (dependencies first, then dependents)
- Detects circular dependencies

### Container State Detection
`IsStackRunning` polls container states via Docker API:
- Returns `true` if ALL containers are running
- Returns `false` if ANY container is stopped/exited

### Labels Used
| Label | Source | Purpose |
|-------|--------|---------|
| `com.docker.compose.project` | Compose | Stack identification |
| `com.docker.compose.service` | Compose | Service name |
| `com.docker.compose.depends_on` | Compose | Dependency graph |

### Known Limitation
Does NOT support `depends_on: condition: service_healthy`. Stacks with health-based dependencies may not restart correctly. Full Compose orchestration could be added via Compose v2 library in future.

## Why
API approach chosen over Compose CLI/library:
- Lighter dependency footprint
- Sufficient for most stacks (simple dependency chains)
- Faster execution (no subprocess overhead)

Pattern validated by Watchtower's successful API-only approach.

## Related
- `pkg_ports` - Defines `ComposeController` interface
- `patterns_from_similar_projects` - Topological sort pattern from Watchtower
