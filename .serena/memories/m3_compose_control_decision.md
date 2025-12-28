# M3 Decision: Compose Control Strategy

**Decision Date**: 2025-12-28
**GitHub Issue**: #109
**Status**: ✅ DECIDED

## Decision

**M3 MVP**: Use **Docker API + Labels + Topological Sort** (Phase 1)

**Future**: Add **Compose v2 Library** as optional backend for complex stacks (Phase 2)

## Rationale

### Why NOT Compose v2 Library for M3?

1. **Large dependency** - `github.com/docker/compose/v2` brings significant code
2. **Most stacks are simple** - Don't need `depends_on: condition: service_healthy`
3. **Faster MVP** - Can ship sooner with simpler approach
4. **Proven pattern** - Watchtower successfully uses API-only approach

### Why API + Labels Works for M3

1. **Stack identification**: Read `com.docker.compose.project` label
2. **Dependency order**: Read `com.docker.compose.depends_on` label + topological sort
3. **Configuration**: Read `bosun.*` labels for job settings
4. **Container control**: Docker API `ContainerStop`/`ContainerStart`

### Known Limitation (Documented)

> ⚠️ M3 does not support `depends_on: condition: service_healthy`.
> Stacks with health-based dependencies may not restart correctly.
> Full Compose orchestration planned for future milestone.

## Research Summary

### Watchtower (API-only approach)
- Uses Docker SDK API exclusively
- Implements topological sort for dependency ordering
- Does NOT read Compose's native `depends_on` (uses own label)
- Does NOT wait for health checks after start
- Pattern: Stop dependents → dependencies, Start dependencies → dependents

### Portainer (Compose v2 library approach)
- Uses `github.com/docker/compose/v2/pkg/compose` as Go library
- Full Compose feature set (health checks, conditions, orphan cleanup)
- `composeService.Up()` / `composeService.Down()` for orchestration
- `WaitForStatus()` polling for health verification

## M3 Implementation Plan

### Labels to Read

| Label | Source | Purpose |
|-------|--------|---------|
| `com.docker.compose.project` | Compose | Identify stack name |
| `com.docker.compose.service` | Compose | Service name within stack |
| `com.docker.compose.depends_on` | Compose | Dependency graph (partial) |
| `bosun.backup.enabled` | User | Enable/disable backup |
| `bosun.backup.worker-image` | User | Worker container image |
| `bosun.backup.timeout` | User | Job timeout |
| `bosun.backup.stop-stack` | User | Whether to stop stack |

### Algorithm

```
1. List containers with label com.docker.compose.project=<stack>
2. Build dependency graph from com.docker.compose.depends_on
3. Topological sort (DFS with cycle detection)
4. Stop: Iterate reversed sorted order (dependents first)
5. Run worker container
6. Start: Iterate sorted order (dependencies first)
7. (No health check waiting in M3)
```

### Port Interface (Simplified)

```go
type ComposeController interface {
    // StopStack stops all containers in a compose stack
    // Respects dependency order (dependents stopped first)
    StopStack(ctx context.Context, projectName string) error

    // StartStack starts all containers in a compose stack
    // Respects dependency order (dependencies started first)
    StartStack(ctx context.Context, projectName string) error

    // ListStackContainers returns containers belonging to a stack
    ListStackContainers(ctx context.Context, projectName string) ([]Container, error)
}
```

## Future: Phase 2 (Compose v2 Library)

**Trigger**: User feedback requesting health check support, or complex stack failures

**Implementation**:
- Add `github.com/docker/compose/v2` dependency
- Create alternative adapter using `compose.NewComposeService()`
- Config option: `compose.backend: api|library`
- Default to `api` for backward compatibility

**Benefits**:
- Full `depends_on` conditions support
- Automatic health check waiting
- Orphan container cleanup
- Build/pull support if needed

## Files to Create

- `internal/ports/compose.go` - ComposeController interface (#115)
- `internal/adapters/docker/compose/` - API-based adapter (#118)
- Future: `internal/adapters/compose/` - Compose v2 library adapter

## Related Issues

- #109 - This research (CLOSED)
- #115 - ComposeController port interface
- #118 - ComposeController adapter (API-based)
- TBD - Future: Compose v2 library support

---

*This decision may be revisited if M3 users report issues with complex stacks.*
