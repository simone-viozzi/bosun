# Patterns from Similar Projects

## Scope
Valuable patterns and lessons learned from analyzing Watchtower and Portainer codebases, applicable to Bosun's container orchestration and job execution features.

## What
Extracted patterns from:
- **Watchtower** - Container update orchestrator (repo: `containrrr/watchtower`)
- **Portainer** - Docker management platform (repo: `portainer/portainer`)

These patterns informed decisions in M3 (Job Execution MVP).

## Why
Research conducted during M3 spec development to answer key architectural questions:
- API vs CLI vs Library for Compose stack control
- Container dependency ordering
- Health check and status waiting

## Related
- `pkg_adapters_docker_compose` - Uses API approach informed by Watchtower
- `pkg_adapters_docker_worker` - Worker container lifecycle

---

## Topological Sort for Container Dependencies (Watchtower)

**Source**: `pkg/sorter/sort.go` in Watchtower

Algorithm for ordering containers by dependencies using depth-first search (DFS):

```go
// Pattern: Dependencies get processed BEFORE dependents
func (ds *dependencySorter) visit(c Container) error {
    // Cycle detection via temporary mark
    if _, ok := ds.marked[c.Name()]; ok {
        return fmt.Errorf("circular reference to %s", c.Name())
    }
    ds.marked[c.Name()] = true
    defer delete(ds.marked, c.Name())

    // Visit dependencies FIRST (recursively)
    for _, linkName := range c.Links() {
        if linkedContainer := ds.findUnvisited(linkName); linkedContainer != nil {
            if err := ds.visit(*linkedContainer); err != nil {
                return err
            }
        }
    }

    // Add to sorted list AFTER dependencies
    ds.removeUnvisited(c)
    ds.sorted = append(ds.sorted, c)
    return nil
}
```

**Key properties**:
- If A depends on B, B appears before A in sorted list
- **For stopping**: iterate from END to START (stop dependents first)
- **For starting**: iterate from START to END (start dependencies first)
- Detects circular dependencies early

**Applicability**: HIGH - Critical for stop/start ordering in Compose stacks.

---

## Error Collection Pattern (Watchtower)

**Source**: `internal/actions/update.go` in Watchtower

Instead of fail-fast, collect errors and continue processing:

```go
func stopContainersInReversedOrder(...) (failed map[ContainerID]error, stopped map[ImageID]bool) {
    failed = make(map[ContainerID]error, len(containers))
    for i := len(containers) - 1; i >= 0; i-- {
        if err := stopStaleContainer(containers[i], client, params); err != nil {
            failed[containers[i].ID()] = err  // Collect error but continue
        } else {
            stopped[containers[i].SafeImageID()] = true
        }
    }
    return
}
```

**Benefits**:
- User sees full picture of what went wrong
- Partial operations can still succeed
- Better UX than cryptic "failed at container X"

**Applicability**: MEDIUM - Useful for backup job execution, less so for stack control (may want fail-fast on critical errors).

---

## Label-Based Dependency Sources (Watchtower)

**Source**: `pkg/container/container.go` in Watchtower

Watchtower checks 3 sources for container dependencies:

| Source | Support | Notes |
|--------|---------|-------|
| Custom label (`com.centurylinklabs.watchtower.depends-on`) | ✅ | Must set manually |
| Legacy `--link` flags | ✅ | From HostConfig.Links |
| `network_mode: service:container` | ✅ | Implicit dependency |
| `com.docker.compose.depends_on` | ❌ | NOT supported |

**Critical insight**: Watchtower does NOT read Compose's native `depends_on` label. Users must manually duplicate dependency info.

**Applicability**: MEDIUM - Bosun's Compose adapter uses `com.docker.compose.depends_on` from Compose labels, avoiding this limitation.

---

## Compose v2 Library Integration (Portainer)

**Source**: `pkg/libstack/compose/composeplugin.go` in Portainer

Portainer uses Docker Compose v2 as a Go library, not CLI:

```go
import (
    "github.com/docker/compose/v2/pkg/api"
    "github.com/docker/compose/v2/pkg/compose"
)

// Create compose service
composeService := compose.NewComposeService(cli)

// Use library API directly
composeService.Up(ctx, project, opts)     // Start stack
composeService.Down(ctx, name, downOpts)  // Stop stack
composeService.Ps(ctx, name, psOpts)      // List containers
composeService.Build(ctx, project, opts)  // Build images
```

**Benefits over CLI**:
- No subprocess overhead
- Structured errors (no text parsing)
- Full Compose feature set (dependency graphs, health checks, orphan cleanup)

**Decision for Bosun**: Chose NOT to use Compose library (too heavy a dependency). Instead, Bosun uses Docker API directly with custom dependency parsing from `com.docker.compose.depends_on` labels.

**Applicability**: INFORMATIONAL - Documents why Bosun took a different approach.

---

## Status Polling with WaitForStatus (Portainer)

**Source**: `pkg/libstack/compose/status.go` in Portainer

Pattern for waiting on async container state changes:

```go
func WaitForStatus(ctx context.Context, name string, status Status) WaitResult {
    for {
        select {
        case <-ctx.Done():
            return WaitResult{Timeout: true}
        default:
        }

        containerSummaries, err := composeService.Ps(psCtx, name, api.PsOptions{All: true})
        aggregateStatus := aggregateStatuses(containerSummaries)

        if aggregateStatus == status {
            return WaitResult{Status: status}
        }

        time.Sleep(1 * time.Second)
    }
}
```

**Pattern elements**:
- Context-based timeout (not duration-based)
- Polling loop with sleep interval
- Status aggregation across multiple containers
- Early exit on success or context cancellation

**Applicability**: HIGH - Bosun's `IsStackRunning` and container status checks use similar polling pattern.

---

## Docker CLI Wrapper Pattern (Portainer)

**Source**: `pkg/libstack/compose/composeplugin.go` in Portainer

Pattern for wrapping Docker client with connection management:

```go
func withCli(ctx context.Context, options Options, fn func(context.Context, *DockerCli) error) error {
    cli, err := command.NewDockerCli(command.WithCombinedStreams(log.Logger))
    if err != nil {
        return err
    }

    opts := flags.NewClientOptions()
    if options.Host != "" {
        opts.Hosts = []string{options.Host}
    }

    if err := cli.Initialize(opts); err != nil {
        return fmt.Errorf("unable to initialize Docker client: %w", err)
    }
    defer cli.Client().Close()  // Cleanup

    // Inject registry auth
    for _, r := range options.Registries {
        cli.ConfigFile().AuthConfigs[r.ServerAddress] = r
    }

    return fn(ctx, cli)
}
```

**Pattern elements**:
- Deferred cleanup
- Options injection (host, registries)
- Error wrapping with context

**Applicability**: LOW - Bosun uses simpler Docker SDK client without CLI wrapper.

---

## 3-Layer Architecture (Portainer)

**Source**: Portainer project structure

```
HTTP Handlers (api/http/handler/stacks/)
       ↓
Service Layer (api/exec/compose_stack.go)
       ↓
Port Interface (pkg/libstack/libstack.go)
       ↑
Adapter Implementation (pkg/libstack/compose/)
```

Validates hexagonal architecture approach used in Bosun:
- CLI commands → App services → Port interfaces ← Adapters

**Applicability**: HIGH - Confirms Bosun's architectural choices align with production systems.

---

## Lifecycle Hooks via Labels (Watchtower)

**Source**: `pkg/lifecycle/lifecycle.go` in Watchtower

Pattern for pre/post operation hooks configured via container labels:

```go
func ExecutePreUpdateCommand(client Client, container Container) (skip bool, err error) {
    command := container.GetLifecyclePreUpdateCommand()  // From label
    if len(command) == 0 {
        return false, nil
    }

    timeout := container.PreUpdateTimeout()  // From label
    return client.ExecuteCommand(container.ID(), command, timeout)
}
```

**Labels used**:
- `com.centurylinklabs.watchtower.lifecycle.pre-update`: Command to run
- `com.centurylinklabs.watchtower.lifecycle.pre-update-timeout`: Timeout in minutes

**Applicability**: FUTURE - Could be used for pre-backup hooks (e.g., `pg_dump` before stopping).

---

## Signal Handling for Graceful Shutdown (Watchtower)

**Source**: `pkg/container/client.go` in Watchtower

```go
func (client dockerClient) StopContainer(c Container, timeout time.Duration) error {
    signal := c.StopSignal()  // From label, default: SIGTERM
    if signal == "" {
        signal = defaultStopSignal
    }

    if c.IsRunning() {
        if err := client.api.ContainerKill(bg, idStr, signal); err != nil {
            return err
        }
    }

    _ = client.waitForStopOrTimeout(c, timeout)
    // Force remove if still running
}
```

**Pattern**:
- Configurable signal via label
- Wait with timeout for graceful stop
- Force action on timeout

**Applicability**: MEDIUM - Bosun's worker uses similar SIGTERM → grace → SIGKILL pattern.

---

## What Watchtower Gets Wrong (Lessons)

**No Compose `depends_on` support**: Must manually add Watchtower-specific labels.

**Destructive stop**: Watchtower's "stop" removes containers entirely (suitable for image updates, not backups).

**No health check waiting**: After start, doesn't verify container is healthy.

**Takeaway**: Bosun's approach (Docker API + Compose labels for deps, non-destructive stop/start, status polling) addresses these gaps.
