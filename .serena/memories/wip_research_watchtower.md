# WIP Research: Watchtower Analysis

**Status**: ✅ COMPLETE
**Related To**: #109 (Compose Control), #110 (Worker Architecture), #117 (Failure Handling)
**Repo**: `/home/simone/workspace/bosun/watchtower`
**Completed**: 2025-12-28

## Research Objectives

### Primary: Compose Control Question (#109)
- How does Watchtower interact with Docker? (API vs CLI)
- How does it handle container stop/start sequences?
- How does it handle dependencies between containers?

### Secondary: Design Patterns for Bosun
- Container client abstraction patterns
- Error handling for Docker operations
- Graceful shutdown / signal handling
- Lifecycle hooks implementation
- Configuration via labels vs environment

---

## Findings

### 1. Docker Client Architecture

**Location**: `pkg/container/client.go`

**Pattern Used**: Direct Docker SDK API (NOT CLI)

Watchtower uses `github.com/docker/docker/client` SDK exclusively for all Docker operations. No shelling out to docker CLI.

**Code Evidence**:
```go
// Import from pkg/container/client.go:3-14
import (
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    sdkClient "github.com/docker/docker/client"
    // ...
)

// NewClient creates the API client
func NewClient(opts ClientOptions) Client {
    cli, err := sdkClient.NewClientWithOpts(sdkClient.FromEnv)
    if err != nil {
        log.Fatalf("Error instantiating Docker client: %s", err)
    }
    return dockerClient{
        api:           cli,
        ClientOptions: opts,
    }
}

// StopContainer uses API directly
func (client dockerClient) StopContainer(c t.Container, timeout time.Duration) error {
    bg := context.Background()
    signal := c.StopSignal()
    if signal == "" {
        signal = defaultStopSignal
    }

    if c.IsRunning() {
        log.Infof("Stopping %s (%s) with %s", c.Name(), shortID, signal)
        if err := client.api.ContainerKill(bg, idStr, signal); err != nil {
            return err
        }
    }

    _ = client.waitForStopOrTimeout(c, timeout)

    // Remove container after stop
    if err := client.api.ContainerRemove(bg, idStr, types.ContainerRemoveOptions{
        Force: true,
        RemoveVolumes: client.RemoveVolumes
    }); err != nil {
        return err
    }
    return nil
}
```

**Relevance to Bosun**:
- **HIGH**: Confirms API-only approach is viable for container operations
- Bosun should use Docker SDK for fine-grained control
- For Compose stacks, we still need to decide: use API with compose labels OR shell to `docker compose` CLI

---

### 2. Container Stop/Start Flow

**Location**: `internal/actions/update.go`, `pkg/container/client.go`

**Sequence**:
1. **Scan for stale containers** - Check which containers need updates
2. **Sort by dependencies** - Use `sorter.SortByDependencies()` for correct order
3. **Stop in REVERSED order** - Stop dependents before dependencies
4. **Execute pre-update hooks** (optional) - Run lifecycle commands
5. **Wait for stop with timeout** - Poll ContainerInspect until stopped
6. **Remove old containers** - ContainerRemove (keeps volumes by default)
7. **Restart in SORTED order** - Start dependencies before dependents
8. **Execute post-update hooks** (optional)

**Code Evidence**:
```go
// From Update() in internal/actions/update.go:66-73
containers, err = sorter.SortByDependencies(containers)
if err != nil {
    return nil, err
}

UpdateImplicitRestart(containers)  // Mark linked containers for restart

// Stop in reverse order
failedStop, stoppedImages := stopContainersInReversedOrder(containersToUpdate, client, params)
progress.UpdateFailed(failedStop)

// Start in sorted order
failedStart := restartContainersInSortedOrder(containersToUpdate, client, params, stoppedImages)
progress.UpdateFailed(failedStart)
```

**Signal Handling**:
- Uses `ContainerKill()` with configurable signal (default: SIGTERM)
- Signal can be set via label: `com.centurylinklabs.watchtower.stop-signal`
- Waits for graceful stop with timeout
- Force removes container if still running after timeout

**Timeout Behavior**:
- Global timeout passed via `params.Timeout`
- Per-container pre-update timeout: `com.centurylinklabs.watchtower.lifecycle.pre-update-timeout` (minutes)
- Polling loop: 1 second intervals checking `ContainerInspect().State.Running`

**Relevance to Bosun**:
- **HIGH**: This is exactly the pattern Bosun needs for compose stack control
- Stop/Start order is CRITICAL for database → app → proxy chains
- Bosun must implement similar dependency sorting
- Timeout + signal handling is essential for graceful shutdown

---

### 3. Lifecycle Hooks

**Location**: `pkg/lifecycle/lifecycle.go`

**Hook Types**:
- Pre-check hooks (`lifecycle.pre-check`) - Before scanning for updates
- Post-check hooks (`lifecycle.post-check`) - After scanning
- **Pre-update hooks** (`lifecycle.pre-update`) - Before stopping container
- **Post-update hooks** (`lifecycle.post-update`) - After starting container

**Implementation Pattern**:
Commands are stored as labels on containers and executed via `docker exec`:

```go
// From ExecutePreUpdateCommand() in pkg/lifecycle/lifecycle.go:63-80
func ExecutePreUpdateCommand(client container.Client, container types.Container) (SkipUpdate bool, err error) {
    timeout := container.PreUpdateTimeout()
    command := container.GetLifecyclePreUpdateCommand()  // Reads from label

    if len(command) == 0 {
        clog.Debug("No pre-update command supplied. Skipping")
        return false, nil
    }

    if !container.IsRunning() || container.IsRestarting() {
        clog.Debug("Container is not running. Skipping pre-update command.")
        return false, nil
    }

    clog.Debug("Executing pre-update command.")
    return client.ExecuteCommand(container.ID(), command, timeout)
}

// ExecuteCommand uses Docker API ContainerExecCreate + ContainerExecStart
// Exit code 75 (EX_TEMPFAIL) → skips update without error
```

**Label Configuration**:
- `com.centurylinklabs.watchtower.lifecycle.pre-update`: Command to run
- `com.centurylinklabs.watchtower.lifecycle.pre-update-timeout`: Timeout in minutes (default: 1)
- Command runs INSIDE the container via `docker exec`
- Exit code 75 = temporary failure, skip update

**Relevance to Bosun**:
- **MEDIUM-HIGH**: Bosun could use similar pattern for backup preparation
- Example: `bosun.backup.pre-job: "pg_dump -Fc > /backup/dump.sql"`
- Different from Watchtower: Bosun's backup command runs AFTER stopping, not before
- Pattern is still valuable: label-based configuration + docker exec execution

---

### 4. Label-Based Configuration

**Location**: `pkg/container/container.go`, `pkg/container/metadata.go`

**Labels Used**:
| Label | Purpose | Default |
|-------|---------|---------|
| `com.centurylinklabs.watchtower.enable` | Enable/disable updates | true |
| `com.centurylinklabs.watchtower.monitor-only` | Monitor but don't update | false |
| `com.centurylinklabs.watchtower.no-pull` | Skip image pull | false |
| `com.centurylinklabs.watchtower.stop-signal` | Custom stop signal | SIGTERM |
| `com.centurylinklabs.watchtower.depends-on` | Comma-separated dependencies | (empty) |
| `com.centurylinklabs.watchtower.lifecycle.pre-update` | Pre-update command | (empty) |
| `com.centurylinklabs.watchtower.lifecycle.post-update` | Post-update command | (empty) |
| `com.centurylinklabs.watchtower.lifecycle.pre-update-timeout` | Timeout (minutes) | 1 |
| `com.centurylinklabs.watchtower.lifecycle.post-update-timeout` | Timeout (minutes) | 1 |

**Dependency Resolution**:
```go
// From (Container).Links() in pkg/container/container.go:173-205
func (c Container) Links() []string {
    var links []string

    // 1. Check custom depends-on label
    dependsOnLabelValue := c.getLabelValueOrEmpty(dependsOnLabel)
    if dependsOnLabelValue != "" {
        for _, link := range strings.Split(dependsOnLabelValue, ",") {
            if !strings.HasPrefix(link, "/") {
                link = "/" + link
            }
            links = append(links, link)
        }
        return links
    }

    // 2. Check HostConfig.Links (legacy docker --link)
    for _, link := range c.containerInfo.HostConfig.Links {
        name := strings.Split(link, ":")[0]
        links = append(links, name)
    }

    // 3. Check NetworkMode container:xxx (implicit dependency)
    networkMode := c.containerInfo.HostConfig.NetworkMode
    if networkMode.IsContainer() {
        links = append(links, networkMode.ConnectedContainer())
    }

    return links
}
```

**Relevance to Bosun**:
- **HIGH**: Watchtower has 3 ways to detect dependencies (label, links, network mode)
- Bosun should leverage `com.docker.compose.depends_on` label from Compose
- Custom `bosun.depends-on` label for override/additions
- NetworkMode container sharing is an implicit dependency!

**⚠️ CRITICAL: Watchtower does NOT read Compose's native `depends_on`**:

Watchtower only reads its own label (`com.centurylinklabs.watchtower.depends-on`), NOT Docker Compose's native label (`com.docker.compose.depends_on`):

```go
// From metadata.go - only Watchtower's own label is defined
dependsOnLabel = "com.centurylinklabs.watchtower.depends-on"  // ← Watchtower's label

// container.go - Links() function only checks Watchtower's label
dependsOnLabelValue := c.getLabelValueOrEmpty(dependsOnLabel)  // ← Only checks Watchtower label
```

| Dependency Source | Watchtower Support |
|-------------------|-------------------|
| `com.centurylinklabs.watchtower.depends-on` label | ✅ Yes (must set manually) |
| `com.docker.compose.depends_on` (Compose's label) | ❌ **NO** |
| Legacy `--link` flags | ✅ Yes |
| `network_mode: service:container` | ✅ Yes (implicit) |

**Implication**: If you use Compose `depends_on` in your docker-compose.yml, Watchtower will NOT automatically respect it. You must manually add the Watchtower-specific label:

```yaml
services:
  app:
    depends_on:
      - db    # <-- Watchtower IGNORES this
    labels:
      com.centurylinklabs.watchtower.depends-on: "db"  # <-- Must add manually!
```

This is why Bosun should use the **Compose v2 library** (Portainer's approach) - it reads the compose file directly and knows all dependencies natively without extra labels.

---

### 5. Error Handling Patterns

**Location**: `pkg/container/errors.go`, `internal/actions/update.go`

**Pattern Used**:
Simple sentinel errors (not custom error types or wrapping):

```go
// From pkg/container/errors.go
var errorNoImageInfo = errors.New("no available image info")
var errorNoContainerInfo = errors.New("no available container info")
var errorInvalidConfig = errors.New("container configuration missing or invalid")
var errorLabelNotFound = errors.New("label was not found in container")
```

**Error Handling in Update Flow**:
```go
// From stopStaleContainer() in internal/actions/update.go:137-172
func stopStaleContainer(container types.Container, client container.Client, params types.UpdateParams) error {
    if container.IsWatchtower() {
        log.Debugf("This is the watchtower container %s", container.Name())
        return nil  // Skip self
    }

    if !container.ToRestart() {
        return nil  // Skip non-restartable
    }

    // Pre-update hook
    if params.LifecycleHooks {
        skipUpdate, err := lifecycle.ExecutePreUpdateCommand(client, container)
        if err != nil {
            log.Error(err)
            log.Info("Skipping container as the pre-update command failed")
            return err  // Propagate error, skip this container
        }
        if skipUpdate {
            log.Debug("Skipping container as the pre-update command returned exit code 75 (EX_TEMPFAIL)")
            return errors.New("skipping container as the pre-update command returned exit code 75 (EX_TEMPFAIL)")
        }
    }

    // Stop container
    if err := client.StopContainer(container, params.Timeout); err != nil {
        log.Error(err)
        return err  // Propagate error but continue with other containers
    }
    return nil
}

// Errors are collected in map, update continues:
func stopContainersInReversedOrder(...) (failed map[types.ContainerID]error, stopped map[types.ImageID]bool) {
    failed = make(map[types.ContainerID]error, len(containers))
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

**Relevance to Bosun**:
- **MEDIUM**: Error collection pattern is useful - don't fail fast, collect all errors
- Bosun should follow similar: if stack stop fails partially, collect which containers failed
- Allows user to see full picture of what went wrong
- Different needs: Bosun might want to fail fast on critical errors (can't stop DB)

---

### 6. Container Sorting/Ordering

**Location**: `pkg/sorter/sort.go`

**Algorithm**: Topological sort using depth-first search (DFS)

```go
// From SortByDependencies() and visit()
func (ds *dependencySorter) Sort(containers []types.Container) ([]types.Container, error) {
    ds.unvisited = containers
    ds.marked = map[string]bool{}  // For cycle detection

    for len(ds.unvisited) > 0 {
        if err := ds.visit(ds.unvisited[0]); err != nil {
            return nil, err
        }
    }

    return ds.sorted, nil
}

func (ds *dependencySorter) visit(c types.Container) error {
    // Cycle detection
    if _, ok := ds.marked[c.Name()]; ok {
        return fmt.Errorf("circular reference to %s", c.Name())
    }

    // Mark visited (temporarily for cycle detection)
    ds.marked[c.Name()] = true
    defer delete(ds.marked, c.Name())

    // Recursively visit dependencies FIRST
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

**Dependency Handling**:
- Dependencies are added to sorted list BEFORE dependents
- Result: sorted[0] has no dependencies, sorted[n-1] depends on most
- **For stopping**: Reverse the array (stop dependents first)
- **For starting**: Use sorted order (start dependencies first)
- Detects circular dependencies and returns error

**Key Properties**:
1. If A depends on B, B appears before A in sorted list
2. Stopping: iterate from end to start
3. Starting: iterate from start to end
4. Handles missing links gracefully (findUnvisited returns nil)

**Relevance to Bosun**:
- **CRITICAL**: This is the exact algorithm Bosun needs
- Compose `depends_on` creates dependency graph
- Must use topological sort to determine stop/start order
- Cycle detection prevents infinite loops
- Pattern: Dependencies get processed BEFORE dependents

---

## Key Files to Examine

- [x] `pkg/container/client.go` - Docker client wrapper (uses SDK API directly)
- [x] `pkg/container/container.go` - Container model (labels, dependencies)
- [x] `internal/actions/update.go` - Main update logic (stop/start flow)
- [x] `pkg/lifecycle/lifecycle.go` - Lifecycle hooks (pre/post commands)
- [x] `pkg/sorter/sort.go` - Container ordering (topological sort)
- [x] `pkg/container/metadata.go` - Label constants
- [x] `pkg/container/errors.go` - Error handling

---

## Summary

### Answer to #109 (API vs CLI)

**Watchtower uses DIRECT DOCKER SDK API exclusively.**

**Rationale:**
- Uses `github.com/docker/docker/client` for all operations
- NO shelling out to docker CLI anywhere in codebase
- Calls `ContainerKill`, `ContainerRemove`, `ContainerCreate`, `ContainerStart` directly
- Gets fine-grained control over signals, timeouts, error handling

**Implications for Bosun:**
For individual container operations, API is clearly superior. However, Bosun faces a different problem:
- Watchtower manages **individual containers** with manually defined dependencies
- Bosun needs to manage **Compose stacks** where dependency metadata lives in compose.yaml
- **Key question**: Can Bosun read Compose dependency graph from labels, or must it shell to `docker compose`?

**Answer**: Compose stores `com.docker.compose.depends_on` labels, BUT:
- Label only has service NAMES, not full dependency conditions
- `depends_on: {condition: service_healthy}` logic NOT in labels
- Health check waiting logic lives in docker-compose binary

**Recommendation for Bosun (#109):**
1. **For simple stacks**: Use API + topological sort on `com.docker.compose.depends_on` labels
2. **For complex stacks with health checks**: Shell to `docker compose stop/start`
3. **Hybrid approach**: Read labels to identify stack containers, but use CLI for orchestration

### Useful Patterns for Bosun

| Pattern | Source File | Applicability | Description |
|---------|-------------|---------------|-------------|
| **Topological Sort** | `pkg/sorter/sort.go` | **HIGH** | DFS-based dependency sorting. Must have for stack control. Can detect circular deps. |
| **Label-Based Config** | `pkg/container/container.go` | **HIGH** | Read config from labels with defaults. Bosun should read `bosun.*` labels for overrides. |
| **Error Collection** | `internal/actions/update.go` | **MEDIUM-HIGH** | Don't fail fast - collect all errors in map, continue processing, report at end. |
| **Interface Abstraction** | `pkg/container/client.go` | **MEDIUM** | `Client` interface wraps Docker SDK. Enables testing with mocks. |
| **Wait with Timeout** | `pkg/container/client.go` | **MEDIUM** | Poll ContainerInspect in loop with timeout. Simple but effective. |
| **Lifecycle Hooks** | `pkg/lifecycle/lifecycle.go` | **MEDIUM** | Label-based pre/post commands via docker exec. Useful for backup prep. |

### Insights for Other Research

**#110 (Worker Architecture)**:
- Watchtower runs all operations in main goroutine (synchronous)
- No worker pool pattern observed
- Containers are processed sequentially in dependency order
- Bosun could parallelize backup jobs, but stack control should be sequential

**#117 (Failure Handling)**:
- Watchtower collects errors but continues processing remaining containers
- Each container stop failure is logged but doesn't halt the batch
- For Bosun: Stack control should probably fail fast (can't continue if DB fails to stop)
- Backup jobs can continue even if one job fails (collect errors like Watchtower)

**Health Check Waiting**:
- Watchtower does NOT wait for health checks after starting containers
- Only waits for container state to be "running"
- For Bosun: If using API approach, must implement health check waiting ourselves
- If using `docker compose up`, Compose handles health check waiting automatically

### Critical Findings

1. **No Compose Integration**: Watchtower manages individual containers, not stacks
2. **Dependency Source**: Uses custom labels + legacy links + network mode
3. **No Health Waiting**: Starts containers but doesn't verify health
4. **Stop = Kill + Remove**: Stopping removes the container entirely (recreate on start)
5. **Rolling vs Batch**: Supports both rolling restart and batch stop/start

### Warnings

⚠️ **Watchtower's approach assumes containers are cattle, not pets**:
- Stops = removes container completely
- Creates new container on restart (not just start existing one)
- This works for Watchtower's use case (updating images) but is OVERKILL for Bosun

⚠️ **Missing features Bosun needs**:
- No native Compose stack handling
- No health check waiting after start
- No graceful handling of `depends_on: {condition: service_healthy}`

⚠️ **Compose labels may not be sufficient**:
- `com.docker.compose.depends_on` exists but lacks condition details
- Must parse compose.yaml OR use docker-compose CLI for full orchestration

---

**FINAL VERDICT**: Watchtower validates that pure-API approach works for container control,
BUT Bosun should seriously consider hybrid approach or full CLI delegation for Compose stacks.

---

*Update this memory as you examine each file.*
