# Data Model: Job Execution MVP (M3)

**Feature Branch**: `009-job-execution-mvp`
**Date**: 2025-12-29

This document defines domain types, port interfaces, and error handling for M3.

---

## Port Interfaces

### ComposeController (`internal/ports/compose.go`) - #115

Controls Compose stack lifecycle via Docker API.

```go
// ComposeController manages Docker Compose stacks via Docker API.
// Uses container labels for stack discovery and dependency ordering.
type ComposeController interface {
    // StopStack stops all containers in a Compose stack.
    // Containers are stopped in reverse dependency order (dependents first).
    // Returns StopError if any container fails to stop.
    StopStack(ctx context.Context, projectName string, opts StopOptions) error

    // StartStack starts all containers in a Compose stack.
    // Containers are started in dependency order (dependencies first).
    // Returns StartError if any container fails to start.
    StartStack(ctx context.Context, projectName string, opts StartOptions) error

    // ListStackContainers returns all containers belonging to a stack.
    // Uses com.docker.compose.project label for identification.
    ListStackContainers(ctx context.Context, projectName string) ([]StackContainer, error)

    // IsStackRunning returns true if all containers in the stack are running.
    IsStackRunning(ctx context.Context, projectName string) (bool, error)
}

// StopOptions configures stack stop behavior.
type StopOptions struct {
    // Timeout for stopping each container (default: 30s).
    Timeout time.Duration
}

// StartOptions configures stack start behavior.
type StartOptions struct {
    // Timeout for starting each container (default: 30s).
    Timeout time.Duration
}

// StackContainer represents a container within a Compose stack.
type StackContainer struct {
    ID          string
    Name        string
    ServiceName string            // From com.docker.compose.service
    State       string            // running, exited, etc.
    DependsOn   []string          // Service names from com.docker.compose.depends_on
    Labels      map[string]string // All labels
}
```

### WorkerRunner (`internal/ports/worker.go`) - #116

Executes backup worker containers.

```go
// WorkerRunner creates and executes worker containers.
// Handles container lifecycle, log capture, and timeout enforcement.
type WorkerRunner interface {
    // Run executes a worker container and waits for completion.
    // Returns exit code, captured logs, and any error.
    // Enforces timeout via SIGTERM → grace period → SIGKILL.
    Run(ctx context.Context, config WorkerConfig) (result WorkerResult, err error)
}

// WorkerConfig defines worker container configuration.
type WorkerConfig struct {
    // Image is the container image to run (required).
    Image string

    // Env contains environment variables (BOSUN_* + user pass-through).
    Env map[string]string

    // Mounts defines volume attachments.
    Mounts []VolumeMount

    // Timeout is the maximum execution time (default: 1h).
    Timeout time.Duration

    // KeepOnFailure preserves container on non-zero exit (--keep-failed).
    KeepOnFailure bool

    // RunID is the unique execution ID for container naming.
    RunID string

    // JobName for container naming and logging.
    JobName string
}

// VolumeMount defines a volume attachment for the worker.
type VolumeMount struct {
    // Source is the Docker volume name or host path.
    Source string

    // Target is the mount path inside the container.
    Target string

    // ReadOnly mounts the volume as read-only.
    ReadOnly bool
}

// WorkerResult contains execution outcome.
type WorkerResult struct {
    // ExitCode from container (0 = success).
    ExitCode int

    // Logs captured from stdout/stderr.
    Logs string

    // ContainerID of the executed worker.
    ContainerID string

    // Duration of execution.
    Duration time.Duration

    // TimedOut indicates if execution was terminated due to timeout.
    TimedOut bool
}
```

### JobExecutor (`internal/ports/executor.go`) - #114

Orchestrates complete job execution.

```go
// JobExecutor orchestrates the full job execution lifecycle.
// Coordinates: discover job → validate → stop stack → run worker → restart stack.
type JobExecutor interface {
    // Execute runs a job by name.
    // Returns ExecutionResult with outcome details.
    Execute(ctx context.Context, jobName string, opts ExecuteOptions) (ExecutionResult, error)

    // DryRun returns the execution plan without performing actions.
    DryRun(ctx context.Context, jobName string) (ExecutionPlan, error)
}

// ExecuteOptions configures job execution.
type ExecuteOptions struct {
    // TimeoutOverride overrides all timeout settings (--timeout).
    TimeoutOverride time.Duration

    // StopTimeoutOverride overrides stop timeout (--stop-timeout).
    StopTimeoutOverride time.Duration

    // StartTimeoutOverride overrides start timeout (--start-timeout).
    StartTimeoutOverride time.Duration

    // KeepStopped skips stack restart after worker (--keep-stopped).
    KeepStopped bool

    // KeepFailedWorker preserves worker container on failure (--keep-failed).
    KeepFailedWorker bool

    // Quiet suppresses log output (--quiet).
    Quiet bool
}
```

---

## Domain Entities

### Execution Types (`internal/domain/jobs/run.go`)

```go
// RunStatus represents the state of a job execution.
type RunStatus string

const (
    RunStatusPending   RunStatus = "pending"
    RunStatusRunning   RunStatus = "running"
    RunStatusSuccess   RunStatus = "success"
    RunStatusFailed    RunStatus = "failed"
    RunStatusCancelled RunStatus = "cancelled"
)

// JobRun represents a single execution instance of a job.
type JobRun struct {
    // ID is a unique identifier (UUID v4).
    ID string `json:"id"`

    // JobName references the job being executed.
    JobName string `json:"job_name"`

    // StackName is the target Compose stack.
    StackName string `json:"stack_name"`

    // Status of the execution.
    Status RunStatus `json:"status"`

    // StartedAt is when execution began.
    StartedAt time.Time `json:"started_at"`

    // CompletedAt is when execution finished (zero if running).
    CompletedAt time.Time `json:"completed_at,omitempty"`

    // WorkerExitCode from the worker container (-1 if not run).
    WorkerExitCode int `json:"worker_exit_code"`

    // Error message if failed.
    Error string `json:"error,omitempty"`
}

// ExecutionResult is the outcome of a job execution.
type ExecutionResult struct {
    // Run contains execution metadata.
    Run JobRun `json:"run"`

    // Plan that was executed.
    Plan ExecutionPlan `json:"plan"`

    // StepResults for each executed step.
    StepResults []StepResult `json:"step_results"`

    // WorkerLogs captured from worker container.
    WorkerLogs string `json:"worker_logs,omitempty"`
}

// StepResult is the outcome of a single execution step.
type StepResult struct {
    // Step that was executed.
    Step PlanStep `json:"step"`

    // Status of this step.
    Status RunStatus `json:"status"`

    // Duration of this step.
    Duration time.Duration `json:"duration"`

    // Error if step failed.
    Error string `json:"error,omitempty"`
}
```

### Extended Job Type

Extend existing `Job` in `internal/domain/jobs/types.go`:

```go
// Job fields to add/verify exist:
type Job struct {
    Name             string            `json:"name"`
    Schedule         string            `json:"schedule,omitempty"`
    TargetContainers []string          `json:"target_containers"`
    TargetStacks     []string          `json:"target_stacks"`
    WorkerImage      string            `json:"worker_image"`
    AttachVolumes    []VolumeAttachment `json:"attach_volumes"`
    SourceContainers []string          `json:"source_containers"`

    // M3 additions:
    WorkerEnv       map[string]string `json:"worker_env,omitempty"`        // Pass-through env vars
    Timeout         time.Duration     `json:"timeout,omitempty"`           // Worker timeout
    StopTimeout     time.Duration     `json:"stop_timeout,omitempty"`      // Stack stop timeout
    StartTimeout    time.Duration     `json:"start_timeout,omitempty"`     // Stack start timeout
}
```

---

## Error Types (`internal/domain/jobs/errors.go`)

```go
package jobs

import (
    "errors"
    "fmt"
    "time"
)

// Sentinel errors for error checking.
var (
    ErrJobNotFound       = errors.New("job not found")
    ErrStackNotFound     = errors.New("stack not found")
    ErrImageNotFound     = errors.New("worker image not found")
    ErrStackPartialState = errors.New("stack in partial state")
    ErrExecutionTimeout  = errors.New("execution timeout")
)

// StopError indicates a failure during stack stop.
type StopError struct {
    StackName     string
    ContainerName string
    ContainerID   string
    Cause         error
}

func (e *StopError) Error() string {
    return fmt.Sprintf("failed to stop container %s (%s) in stack %s: %v",
        e.ContainerName, e.ContainerID, e.StackName, e.Cause)
}

func (e *StopError) Unwrap() error { return e.Cause }

// StartError indicates a failure during stack start.
type StartError struct {
    StackName     string
    ContainerName string
    ContainerID   string
    Cause         error
}

func (e *StartError) Error() string {
    return fmt.Sprintf("failed to start container %s (%s) in stack %s: %v",
        e.ContainerName, e.ContainerID, e.StackName, e.Cause)
}

func (e *StartError) Unwrap() error { return e.Cause }

// TimeoutError indicates an operation exceeded its timeout.
type TimeoutError struct {
    Operation string        // "stop", "start", "worker"
    Target    string        // Stack or container name
    Duration  time.Duration // Timeout that was exceeded
}

func (e *TimeoutError) Error() string {
    return fmt.Sprintf("%s operation on %s timed out after %s",
        e.Operation, e.Target, e.Duration)
}

func (e *TimeoutError) Is(target error) bool {
    return target == ErrExecutionTimeout
}

// WorkerError indicates worker container failure.
type WorkerError struct {
    ExitCode int
    Logs     string
}

func (e *WorkerError) Error() string {
    return fmt.Sprintf("worker failed with exit code %d", e.ExitCode)
}

// ValidationError for pre-execution validation failures.
type ExecutionValidationError struct {
    JobName string
    Field   string
    Message string
}

func (e *ExecutionValidationError) Error() string {
    return fmt.Sprintf("job %s validation failed: %s - %s",
        e.JobName, e.Field, e.Message)
}
```

---

## Exit Codes (`internal/cmd/exitcodes.go`)

Extend existing exit codes:

```go
package cmd

// Exit codes for Bosun CLI.
const (
    // Existing codes
    ExitSuccess         = 0
    ExitRuntimeError    = 1
    ExitValidationError = 2

    // M3 additions
    ExitWorkerFailed  = 10 // Worker exited with non-zero code
    ExitStopFailed    = 11 // Failed to stop stack
    ExitStartFailed   = 12 // Failed to restart stack
    ExitTimeout       = 13 // Operation timed out
    ExitImageNotFound = 14 // Worker image not found
    ExitJobNotFound   = 15 // Job name not found
    ExitInterrupted   = 16 // Execution interrupted (Ctrl+C)
)

// ExitCodeFromError maps domain errors to exit codes.
func ExitCodeFromError(err error) int {
    if err == nil {
        return ExitSuccess
    }

    switch {
    case errors.Is(err, jobs.ErrJobNotFound):
        return ExitJobNotFound
    case errors.Is(err, jobs.ErrImageNotFound):
        return ExitImageNotFound
    case errors.Is(err, jobs.ErrExecutionTimeout):
        return ExitTimeout
    default:
        var stopErr *jobs.StopError
        if errors.As(err, &stopErr) {
            return ExitStopFailed
        }
        var startErr *jobs.StartError
        if errors.As(err, &startErr) {
            return ExitStartFailed
        }
        var workerErr *jobs.WorkerError
        if errors.As(err, &workerErr) {
            return ExitWorkerFailed
        }
        return ExitRuntimeError
    }
}
```

---

## Configuration Labels

### New Labels for M3

Add to `internal/config/schema/job_labels.go`:

| Label Key | Type | Default | Description |
|-----------|------|---------|-------------|
| `bosun.job.stop-timeout` | duration | `30s` | Timeout for stopping each container |
| `bosun.job.start-timeout` | duration | `30s` | Timeout for starting each container |
| `bosun.job.worker-env.*` | string | - | Pass-through env vars (prefix stripped) |

```go
// Label definitions to add:
var (
    LabelStopTimeout = LabelSpec{
        Key:         "bosun.job.stop-timeout",
        Type:        TypeDuration,
        Default:     "30s",
        Scope:       ScopeContainer,
        Description: "Timeout for stopping containers before worker execution",
    }

    LabelStartTimeout = LabelSpec{
        Key:         "bosun.job.start-timeout",
        Type:        TypeDuration,
        Default:     "30s",
        Scope:       ScopeContainer,
        Description: "Timeout for starting containers after worker execution",
    }

    // Worker env vars use prefix matching, not individual label specs
    LabelWorkerEnvPrefix = "bosun.job.worker-env."
)
```

---

## Constants and Defaults

```go
// internal/domain/jobs/defaults.go

package jobs

import "time"

// Default timeouts for job execution.
const (
    DefaultStopTimeout   = 30 * time.Second
    DefaultStartTimeout  = 30 * time.Second
    DefaultWorkerTimeout = 1 * time.Hour
    GracePeriod          = 10 * time.Second // SIGTERM → SIGKILL
)

// Container naming format.
const (
    WorkerContainerNameFormat = "bosun-worker-%s-%s" // job-name, run-id[0:8]
)

// Environment variable names injected by Bosun.
const (
    EnvJobName = "BOSUN_JOB_NAME"
    EnvRunID   = "BOSUN_RUN_ID"
    EnvStack   = "BOSUN_STACK"
    EnvDryRun  = "BOSUN_DRY_RUN"
)
```

---

## Type Relationships

```
┌─────────────────────────────────────────────────────────────────┐
│                        Domain Layer                              │
│                                                                  │
│  ┌─────────────┐    ┌────────────────┐    ┌─────────────────┐  │
│  │ Job         │───►│ ExecutionPlan  │───►│ PlanStep        │  │
│  │ (input)     │    │ (from planner) │    │ (action)        │  │
│  └─────────────┘    └────────────────┘    └─────────────────┘  │
│                                                                  │
│  ┌─────────────┐    ┌────────────────┐    ┌─────────────────┐  │
│  │ JobRun      │───►│ ExecutionResult│───►│ StepResult      │  │
│  │ (tracking)  │    │ (outcome)      │    │ (step outcome)  │  │
│  └─────────────┘    └────────────────┘    └─────────────────┘  │
│                                                                  │
│  ┌─────────────┐    ┌────────────────┐    ┌─────────────────┐  │
│  │ StopError   │    │ StartError     │    │ WorkerError     │  │
│  │ TimeoutError│    │ ValidationError│    │ (error types)   │  │
│  └─────────────┘    └────────────────┘    └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        Ports Layer                               │
│                                                                  │
│  ┌─────────────────────┐  ┌─────────────────┐                   │
│  │ ComposeController   │  │ WorkerRunner    │                   │
│  │ - StopStack()       │  │ - Run()         │                   │
│  │ - StartStack()      │  │                 │                   │
│  │ - ListContainers()  │  │                 │                   │
│  └─────────────────────┘  └─────────────────┘                   │
│           │                        │                             │
│           │    ┌───────────────────┘                            │
│           │    │                                                 │
│           ▼    ▼                                                 │
│  ┌─────────────────────┐                                        │
│  │ JobExecutor         │                                        │
│  │ - Execute()         │ (uses both ports)                      │
│  │ - DryRun()          │                                        │
│  └─────────────────────┘                                        │
└─────────────────────────────────────────────────────────────────┘
```
