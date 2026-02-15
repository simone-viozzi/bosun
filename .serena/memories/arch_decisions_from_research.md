# Architecture Decisions from External Research

## Scope
Key design decisions informed by analyzing Watchtower and Portainer codebases during M3 planning.

## What
Research on `containrrr/watchtower` and `portainer/portainer` informed these adopted decisions:

### Docker API over Compose Library
- Portainer uses `docker/compose/v2` as a Go library (full Compose API).
- Bosun chose **Docker API directly** — the Compose library is too heavy a dependency for Bosun's scope.
- Bosun parses `com.docker.compose.depends_on` labels from containers for dependency ordering.

### Non-Destructive Stop/Start
- Watchtower's "stop" removes containers entirely (designed for image updates).
- Bosun uses **non-destructive stop/start** — containers are stopped and restarted, never removed.

### Dependency Ordering via Compose Labels
- Watchtower requires custom labels for dependency ordering (ignores `com.docker.compose.depends_on`).
- Bosun reads native Compose labels, avoiding user duplication.

### Health Check Waiting After Start
- Watchtower does not verify container health after restart.
- Bosun polls container status after starting to confirm readiness.

### Worker Signal Handling
- Both Watchtower and Portainer use SIGTERM → grace period → SIGKILL.
- Bosun adopts the same pattern for worker container lifecycle.

## Why
Rationale documented above per decision. Research conducted during M3 spec development.

## Related
- `pkg_adapters_docker_compose` — Compose stack control adapter
- `pkg_adapters_docker_worker` — Worker container lifecycle
- `pkg_app_executor` — Plan-driven job execution
