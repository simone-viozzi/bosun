# WIP Research: Portainer Analysis

**Status**: IN PROGRESS
**Related To**: #109 (Compose Control), #110 (Worker Architecture), #117 (Failure Handling)
**Repo**: `/home/simone/workspace/bosun/portainer`

## Research Objectives

### Primary: Compose Control Question (#109)
- How does Portainer manage Compose stacks? (API vs CLI)
- How does it handle stack lifecycle (create/start/stop/remove)?
- How does it handle multi-container dependency ordering?

### Secondary: Design Patterns for Bosun
- Stack abstraction layer
- Docker client patterns
- Error handling and recovery
- Health check handling
- Project structure patterns (hexagonal arch in Go?)

---

## Findings

### 1. Stack Management Architecture

**Location**: `pkg/libstack/compose/`, `api/exec/compose_stack.go`, `api/stacks/deployments/`

**Pattern Used**:
- **3-layer abstraction**: HTTP handlers → StackManager service → libstack Deployer
- Clean separation: handlers call `ComposeStackManager.Up/Down`, which internally calls `libstack.Deployer.Deploy/Remove`
- Interface-based design: `portainer.ComposeStackManager` interface, `libstack.Deployer` interface

**Compose Implementation**:
Uses **Docker Compose v2 as a Go library** (`github.com/docker/compose/v2/pkg/compose`), NOT CLI shelling

**Code Evidence**:
```go
// Import from Docker Compose v2 library
import (
    "github.com/docker/compose/v2/pkg/api"
    "github.com/docker/compose/v2/pkg/compose"
)

// Uses compose library API
composeService := compose.NewComposeService(cli)
composeService.Up(ctx, project, opts)
composeService.Down(ctx, projectName, api.DownOptions{...})
```

**Relevance to Bosun**:
- **HIGH**: This is the key finding - they use Compose v2 as a library, not CLI
- Proves it's possible to programmatically control Compose stacks without shelling out
- Provides dependency ordering, health checks, and all compose features natively

---

### 2. Compose Control Approach

**Location**: `pkg/libstack/compose/composeplugin.go`

**Method**: **Go Library (Docker Compose v2 as imported package)**

**Start/Stop Implementation**:
```go
// Deploy (creates and starts containers)
func (c *ComposeDeployer) Deploy(ctx context.Context, filePaths []string, options libstack.DeployOptions) error {
    return c.withComposeService(ctx, filePaths, options.Options, func(composeService api.Compose, project *types.Project) error {
        // ... setup options ...
        if err := composeService.Build(ctx, project, api.BuildOptions{}); err != nil {
            return fmt.Errorf("compose build operation failed: %w", err)
        }
        if err := composeService.Up(ctx, project, opts); err != nil {
            return fmt.Errorf("compose up operation failed: %w", err)
        }
        return nil
    })
}

// Remove (stops and removes containers)
func (c *ComposeDeployer) Remove(ctx context.Context, projectName string, filePaths []string, options libstack.RemoveOptions) error {
    return withCli(ctx, options.Options, func(ctx context.Context, cli *command.DockerCli) error {
        composeService := compose.NewComposeService(cli)
        return composeService.Down(ctx, projectName, api.DownOptions{RemoveOrphans: true, Volumes: options.Volumes})
    })
}
```

**Key Operations**:
- `composeService.Up()` - creates/starts containers (like `docker compose up`)
- `composeService.Down()` - stops/removes containers (like `docker compose down`)
- `composeService.Build()` - builds images
- `composeService.Pull()` - pulls images
- `composeService.Ps()` - lists stack containers

**Health Check Handling**:
Uses `WaitForStatus()` with polling loop that calls `composeService.Ps()` and aggregates container states:
```go
func (c *ComposeDeployer) WaitForStatus(ctx context.Context, name string, status libstack.Status) libstack.WaitResult {
    for {
        containerSummaries, err = composeService.Ps(psCtx, name, api.PsOptions{All: true})
        aggregateStatus, errorMessage := aggregateStatuses(ctx, services)
        if aggregateStatus == status {
            return waitResult
        }
        time.Sleep(1 * time.Second)
    }
}
```

**Relevance to Bosun**:
- **Answer to #109**: Portainer uses **Docker Compose v2 as a Go library** (NOT API, NOT CLI)
- Rationale: Full control, no subprocess overhead, handles all Compose features (dependency order, health checks)
- **We should adopt this approach** for Bosun M3

---

### 3. Docker Client Abstraction

**Location**: `pkg/libstack/compose/composeplugin.go` (function `withCli`)

**Pattern Used**:
Wraps Docker CLI client (`github.com/docker/cli/cli/command.DockerCli`) with connection management:

```go
func withCli(ctx context.Context, options libstack.Options, cliFn func(context.Context, *command.DockerCli) error) error {
    cli, err := command.NewDockerCli(command.WithCombinedStreams(log.Logger))
    opts := flags.NewClientOptions()
    if options.Host != "" {
        opts.Hosts = []string{options.Host}
    }
    if err := cli.Initialize(opts); err != nil {
        return fmt.Errorf("unable to initialize the Docker client: %w", err)
    }
    defer cli.Client().Close()

    // Registry auth setup
    for _, r := range options.Registries {
        cli.ConfigFile().AuthConfigs[r.ServerAddress] = r
    }

    return cliFn(ctx, cli)
}
```

**Error Handling**:
- Wraps errors with context using `fmt.Errorf("...: %w", err)`
- Returns structured errors from compose operations
- Defers cleanup (client close)

**Relevance to Bosun**:
- **MEDIUM**: Shows pattern for wrapping Docker CLI client with options (host, registries)
- Registry auth injection pattern useful if Bosun needs private registries

---

### 4. Project/Hexagonal Architecture

**Location**: `api/`, `pkg/`, layering

**Layer Structure**:
- `api/http/handler/stacks/` - **HTTP handlers** (web layer)
- `api/exec/compose_stack.go` - **ComposeStackManager** (application service)
- `pkg/libstack/compose/` - **ComposeDeployer** (infrastructure adapter)
- `pkg/libstack/libstack.go` - **Deployer interface** (port)

**Dependency Direction**:
```
HTTP Handlers → ComposeStackManager (service) → libstack.Deployer (interface) ← ComposeDeployer (impl)
```

**Relevance to Bosun**:
- **HIGH**: Similar to Bosun's hexagonal architecture!
- Maps to: CLI → App Service → Port Interface ← Adapter
- Validates our design approach

---

### 5. Stack Labels & Metadata

**Location**: `pkg/libstack/compose/composeplugin.go` (function `addServiceLabels`)

**Labels Used**:
| Label | Purpose |
|-------|---------|
| `com.docker.compose.project` | Stack/project name identifier |
| `com.docker.compose.service` | Service name within stack |
| `com.docker.compose.version` | Compose version used |
| `com.docker.compose.working_dir` | Working directory |
| `com.docker.compose.config-files` | Compose files used |
| `com.docker.compose.oneoff` | Whether one-off container |
| `io.portainer.edgeStackId` | Portainer-specific edge stack ID |

**Code Evidence**:
```go
func addServiceLabels(project *types.Project, oneOff bool, edgeStackID portainer.EdgeStackID) {
    for i, s := range project.Services {
        s.CustomLabels = map[string]string{
            api.ProjectLabel:     project.Name,
            api.ServiceLabel:     s.Name,
            api.VersionLabel:     api.ComposeVersion,
            api.WorkingDirLabel:  project.WorkingDir,
            api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
            api.OneoffLabel:      oneOffLabel,
        }
        if edgeStackID > 0 {
            s.CustomLabels.Add(PortainerEdgeStackLabel, strconv.Itoa(int(edgeStackID)))
        }
        project.Services[i] = s
    }
}
```

**Stack Metadata Storage**:
- Compose library automatically adds standard `com.docker.compose.*` labels
- Portainer adds custom labels via `CustomLabels` before deployment
- Metadata persisted in container labels (queryable via Docker API)

**Relevance to Bosun**:
- **LOW-MEDIUM**: Bosun already has label-based discovery
- Pattern: Injecting custom labels before compose deployment

---

### 6. Error Handling & Recovery

**Location**: Various

**Error Types**:
- Standard Go errors wrapped with context via `fmt.Errorf("operation failed: %w", err)`
- Structured error returns from compose library (e.g., build errors, up errors)

**Recovery Patterns**:
```go
// WaitForStatus has timeout via context
psCtx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
defer cancelFunc()

// Cleanup with defer
defer cli.Close()
defer cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})

// Retry loop for status polling
for {
    if ctx.Err() != nil {
        return waitResult // timeout
    }
    time.Sleep(1 * time.Second)
    // ... check status ...
}
```

**No explicit rollback** - relies on Docker/Compose to handle failures during up/down

**Relevance to Bosun**:
- **MEDIUM**: Context-based timeouts important for FR-014 (timeout handling)
- Defer pattern for cleanup
- Polling loop pattern for waiting on async operations

---

### 2. Compose Control Approach

**Location**: `pkg/libstack/compose/`

**Method**:
<!-- CLI wrapper? API? Library? -->

**Start/Stop Implementation**:
<!-- How does it start/stop compose stacks? -->

**Health Check Handling**:
<!-- How does it wait for services to be healthy? -->

**Relevance to Bosun**:
<!-- Direct answer to #109 question -->

---

### 3. Docker Client Abstraction

**Location**: `api/docker/`

**Pattern Used**:
<!-- Interface-based? Direct client? -->

**Error Handling**:
<!-- How are Docker errors wrapped/handled? -->

**Relevance to Bosun**:
<!-- Client abstraction patterns -->

---

### 4. Project/Hexagonal Architecture

**Location**: `api/`, `pkg/`

**Layer Structure**:
- `api/` - <!-- What's here? -->
- `pkg/` - <!-- What's here? -->
- `internal/` - <!-- What's here? -->

**Dependency Direction**:
<!-- How do layers depend on each other? -->

**Relevance to Bosun**:
<!-- Architectural patterns to adopt -->

---

### 5. Stack Labels & Metadata

**Location**: `api/stacks/`

**Labels Used**:
| Label | Purpose |
|-------|---------|
| `io.portainer.*` | <!-- Purpose --> |
| ... | ... |

**Stack Metadata Storage**:
<!-- Database? Labels? Config files? -->

**Relevance to Bosun**:
<!-- How might we track stack state? -->

---

### 6. Error Handling & Recovery

**Location**: Various

**Error Types**:
<!-- Custom error types? -->

**Recovery Patterns**:
<!-- Retry logic? Rollback? -->

**Relevance to Bosun**:
<!-- FR-014, FR-023 from spec -->

---

## Key Files to Examine

- [x] `pkg/libstack/compose/` - Compose stack operations ✅
- [x] `api/stacks/` - Stack API handlers ✅
- [x] `api/docker/` - Docker client abstraction (checked via compose code) ✅
- [x] `api/http/handler/stacks/` - Stack HTTP handlers ✅
- [x] `pkg/libstack/` - Stack library interface ✅

**Key Findings**:
- ✅ Found primary answer: **Compose v2 library integration**
- ✅ Traced architecture: HTTP → Service → Adapter pattern
- ✅ Identified health check/status waiting pattern
- ✅ Found error handling and timeout patterns

---

## Summary

### Answer to #109 (API vs CLI)

**PRIMARY FINDING**: Portainer uses **Docker Compose v2 as a Go library** (`github.com/docker/compose/v2`).

**Rationale**:
1. **NOT CLI**: No `exec.Command("docker", "compose", ...)` found - only in tests
2. **NOT Raw Docker API**: Doesn't replicate Compose logic - imports Compose library
3. **Go Library Import**: Direct integration with `compose.NewComposeService(cli)` and `composeService.Up/Down/Ps/Build/Pull`

**How it works**:
- Creates Docker CLI client (`command.DockerCli`)
- Wraps it with Compose service (`compose.NewComposeService(cli)`)
- Calls Compose library functions: `Up()`, `Down()`, `Build()`, `Pull()`, `Ps()`
- Compose library handles ALL orchestration: dependency order, health checks, recreate logic, orphan cleanup

**Benefits over alternatives**:
- ✅ No subprocess overhead (vs CLI approach)
- ✅ Full Compose feature set without reimplementation (vs API-only)
- ✅ Structured errors, no text parsing
- ✅ Handles edge cases automatically (dependency graphs, health checks, volume orphans)

### Useful Patterns for Bosun

| Pattern | Source File | Applicability | Notes |
|---------|-------------|---------------|-------|
| **Compose v2 Library Integration** | `pkg/libstack/compose/composeplugin.go` | **HIGH** | Direct answer to #109 - we should use this |
| **3-Layer Architecture** | `api/exec/` → `pkg/libstack/` | **HIGH** | Matches Bosun's hexagonal design (CLI → Service → Port → Adapter) |
| **Docker CLI Wrapper** | `composeplugin.go::withCli()` | **MEDIUM** | Pattern for injecting host, registry auth into Docker client |
| **Status Polling Loop** | `status.go::WaitForStatus()` | **HIGH** | For FR-013 (wait for stack healthy) - poll `Ps()` + aggregate states |
| **Context-based Timeouts** | `status.go` | **MEDIUM** | For FR-014 (operation timeouts) |
| **Label Injection** | `composeplugin.go::addServiceLabels()` | **LOW** | Custom labels before deployment (we may not need) |

### Insights for Other Research

**#110 (Worker Architecture)**:
- Portainer uses "compose-unpacker" container pattern for remote deployments
- Not relevant to Bosun (we control stacks locally, workers are separate)

**#117 (Failure Handling)**:
- No explicit rollback in Portainer - relies on Compose library to handle failures
- Uses defer for cleanup (close clients, remove temp containers)
- Context cancellation for timeouts

**Additional Finding - Remote Stack Operations**:
- For remote Docker hosts, Portainer uses a "compose-unpacker" helper container
- Runs compose operations inside a privileged container with socket bind mount
- Relevant only if Bosun needs remote Docker host support (not in current spec)

---------|-------------|---------------|
| <!-- Pattern 1 --> | <!-- file --> | <!-- High/Medium/Low --> |
| <!-- Pattern 2 --> | <!-- file --> | <!-- High/Medium/Low --> |

### Insights for Other Research
- **#110 (Worker Architecture)**: <!-- Any relevant findings -->
- **#117 (Failure Handling)**: <!-- Any relevant findings -->

---

## Portainer Size Warning

Portainer is a large codebase. Focus on:
1. `pkg/libstack/` - Core stack operations
2. `api/stacks/` - Stack service layer
3. Skip frontend (`app/`), Kubernetes (`kubernetes/`), Edge (`edge/`)

---

*Update this memory as you examine each file.*
