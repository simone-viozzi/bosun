# Quickstart: Job Execution MVP (M3)

**Feature Branch**: `009-job-execution-mvp`
**Date**: 2025-12-29

This guide helps developers implement M3 components quickly.

---

## Prerequisites

1. **Branch**: `009-job-execution-mvp`
2. **Go**: 1.24+
3. **Docker**: Running daemon with API access
4. **Dependencies**: `go mod download`

---

## Implementation Order

```
Phase 1: Ports (parallel, ~0.5d each)
├── #115 ComposeController port
├── #116 WorkerRunner port
└── #114 JobExecutor port

Phase 2: Adapters (can parallel after ports)
├── #118 ComposeController adapter (~2d)
└── #119 WorkerRunner adapter (~1.5d)

Phase 3: Application & CLI
├── #121 Executor service (~2d) - after both adapters
└── #122 CLI command (~1d) - after executor

Phase 4: Integration
├── #123 Integration tests (~2d)
└── #120 Documentation (~1d)
```

---

## Quick Commands

```bash
# Run all unit tests
make test

# Run integration tests (requires Docker)
make it

# Build binary
make build

# Test specific package
go test ./internal/ports/...

# Test with verbose output
go test -v ./internal/adapters/docker/compose/...

# Lint
make lint
```

---

## Issue #115: ComposeController Port

**File**: `internal/ports/compose.go`

Copy interface from `specs/009-job-execution-mvp/contracts/compose_controller.go`.

**Checklist**:
- [ ] Copy interface definition
- [ ] Add package doc comment
- [ ] Import `time` package
- [ ] Run `go build ./internal/ports/...`

**Time**: 30 minutes

---

## Issue #116: WorkerRunner Port

**File**: `internal/ports/worker.go`

Copy interface from `specs/009-job-execution-mvp/contracts/worker_runner.go`.

**Checklist**:
- [ ] Copy interface definition
- [ ] Add helper methods (`Success()`, `BuildEnv()`)
- [ ] Run `go build ./internal/ports/...`

**Time**: 30 minutes

---

## Issue #114: JobExecutor Port

**File**: `internal/ports/executor.go`

Copy interface from `specs/009-job-execution-mvp/contracts/job_executor.go`.

**Checklist**:
- [ ] Copy interface definition
- [ ] Import `github.com/simone-viozzi/bosun/internal/domain/jobs`
- [ ] Add helper methods (`Success()`, `StepSuccess()`)
- [ ] Run `go build ./internal/ports/...`

**Time**: 30 minutes

---

## Issue #118: ComposeController Adapter

**Directory**: `internal/adapters/docker/compose/`

**Files**:
```
compose/
├── doc.go            # Package documentation
├── controller.go     # Main implementation
├── topology.go       # Topological sort for dependencies
└── controller_test.go
```

### Key Implementation Points

1. **Constructor**:
```go
func NewController(client *docker.Client) *Controller {
    return &Controller{docker: client}
}
```

2. **ListStackContainers**: Filter by `com.docker.compose.project` label
```go
containers, err := c.docker.ContainerList(ctx, container.ListOptions{
    Filters: filters.NewArgs(
        filters.Arg("label", "com.docker.compose.project="+projectName),
    ),
})
```

3. **Topological Sort**: DFS with cycle detection
```go
// See watchtower/internal/actions/update.go for reference
func topologicalSort(containers []StackContainer) ([]StackContainer, error) {
    // Build adjacency list from DependsOn
    // DFS with visited/inStack tracking
    // Reverse for stop order
}
```

4. **StopStack**: Stop in reverse topological order
```go
for _, c := range reversed(sorted) {
    timeout := int(opts.Timeout.Seconds())
    err := c.docker.ContainerStop(ctx, c.ID, container.StopOptions{
        Timeout: &timeout,
    })
}
```

**Test with**: Mock Docker client, verify stop/start order

**Time**: 2 days

---

## Issue #119: WorkerRunner Adapter

**Directory**: `internal/adapters/docker/worker/`

**Files**:
```
worker/
├── doc.go         # Package documentation
├── runner.go      # Main implementation
└── runner_test.go
```

### Key Implementation Points

1. **Container Creation**:
```go
resp, err := r.docker.ContainerCreate(ctx,
    &container.Config{
        Image: config.Image,
        Env:   config.BuildEnv(),
    },
    &container.HostConfig{
        Binds: formatMounts(config.Mounts),
    },
    nil, nil,
    containerName(config.JobName, config.RunID),
)
```

2. **Timeout Handling**:
```go
timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
defer cancel()

statusCh, errCh := r.docker.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

select {
case err := <-errCh:
    // Timeout - send SIGTERM, wait grace, SIGKILL
case status := <-statusCh:
    return WorkerResult{ExitCode: int(status.StatusCode), ...}, nil
}
```

3. **Log Capture**:
```go
logs, err := r.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
    ShowStdout: true,
    ShowStderr: true,
    Follow:     true,
})
```

4. **Cleanup**:
```go
defer func() {
    if !config.KeepOnFailure || result.ExitCode == 0 {
        r.docker.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
    }
}()
```

**Test with**: testcontainers-go, verify timeout behavior

**Time**: 1.5 days

---

## Issue #121: Executor Service

**Directory**: `internal/app/executor/`

**Files**:
```
executor/
├── doc.go          # Package documentation
├── executor.go     # Main implementation
└── executor_test.go
```

### Key Implementation Points

1. **Constructor** (dependency injection):
```go
func NewExecutor(
    discoverer ports.JobDiscoverer,
    planner ports.JobPlanner,
    compose ports.ComposeController,
    worker ports.WorkerRunner,
) *Executor
```

2. **Execute Flow**:
```go
func (e *Executor) Execute(ctx context.Context, jobName string, opts ExecuteOptions) (ExecutionResult, error) {
    // 1. Discover job
    job, err := e.findJob(ctx, jobName)

    // 2. Validate worker image
    if err := e.validateImage(ctx, job.WorkerImage); err != nil {
        return result, err // Stack NOT stopped
    }

    // 3. Stop stack
    stackStopped := false
    if err := e.compose.StopStack(ctx, job.TargetStacks[0], stopOpts); err != nil {
        return result, err
    }
    stackStopped = true

    // 4. Run worker (with cleanup guarantee)
    defer func() {
        if stackStopped && !opts.KeepStopped {
            e.compose.StartStack(ctx, job.TargetStacks[0], startOpts)
        }
    }()

    workerResult, err := e.worker.Run(ctx, workerConfig)

    // 5. Build result
    return result, nil
}
```

3. **Signal Handling**:
```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigCh
    cancel() // Cancel context, trigger cleanup
}()
```

**Test with**: Mocked ports, verify orchestration order

**Time**: 2 days

---

## Issue #122: CLI Command

**File**: `internal/cmd/job_run.go`

### Key Implementation Points

1. **Command Structure**:
```go
var jobRunCmd = &cobra.Command{
    Use:   "run <job-name>",
    Short: "Execute a backup job",
    Long:  `Execute a backup job by stopping the stack, running the worker, and restarting.`,
    Args:  cobra.ExactArgs(1),
    RunE:  runJobRun,
}

func init() {
    jobCmd.AddCommand(jobRunCmd)

    jobRunCmd.Flags().Bool("dry-run", false, "Show plan without executing")
    jobRunCmd.Flags().Duration("timeout", 0, "Override worker timeout")
    jobRunCmd.Flags().Duration("stop-timeout", 0, "Override stop timeout")
    jobRunCmd.Flags().Duration("start-timeout", 0, "Override start timeout")
    jobRunCmd.Flags().Bool("keep-stopped", false, "Don't restart stack after worker")
    jobRunCmd.Flags().Bool("keep-failed", false, "Keep worker container on failure")
    jobRunCmd.Flags().Bool("quiet", false, "Suppress log output")
    jobRunCmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
}
```

2. **Dry-run Output**:
```go
if dryRun {
    plan, err := executor.DryRun(ctx, jobName)
    // Format and print plan
    return nil
}
```

3. **Exit Code Mapping**:
```go
result, err := executor.Execute(ctx, jobName, opts)
if err != nil {
    os.Exit(ExitCodeFromError(err))
}
if !result.Success() {
    os.Exit(ExitWorkerFailed)
}
```

**Time**: 1 day

---

## Testing Tips

### Unit Tests

```go
// Mock ComposeController
type mockCompose struct {
    stopCalled  bool
    startCalled bool
    stopErr     error
}

func (m *mockCompose) StopStack(ctx context.Context, project string, opts ports.StopOptions) error {
    m.stopCalled = true
    return m.stopErr
}
```

### Integration Tests

```go
func TestJobExecution_HappyPath(t *testing.T) {
    if testing.Short() {
        t.Skip("integration test")
    }

    // Use testcontainers-go to spin up test stack
    compose := testutil.NewComposeTestHarness(t, "testdata/simple-stack")
    compose.Up(t)
    defer compose.Down(t)

    // Run job
    result, err := executor.Execute(ctx, "test-job", ports.DefaultExecuteOptions())
    require.NoError(t, err)
    assert.True(t, result.Success())
}
```

---

## Common Pitfalls

1. **Forgetting to restart stack on error**: Always use `defer` for cleanup
2. **Container name conflicts**: Include RunID in container name
3. **Log buffering**: Stream logs in real-time, don't wait for completion
4. **Context cancellation**: Propagate context cancellation to all operations
5. **Volume removal**: NEVER remove volumes, only containers

---

## Reference Files

- **Watchtower patterns**: `watchtower/internal/actions/update.go`
- **Existing ports**: `internal/ports/planner.go`, `internal/ports/labels.go`
- **Existing adapters**: `internal/adapters/dockerlabels/source.go`
- **Domain types**: `internal/domain/jobs/types.go`
- **Exit codes**: `internal/cmd/exitcodes.go`
