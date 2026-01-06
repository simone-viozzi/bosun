# JobExecutor Interface Contract

**Package**: `internal/ports`
**File**: `executor.go`
**Status**: Updated (interface aligned with implementation)

## Interface Definition

```go
// JobExecutor orchestrates job execution by interpreting execution plans.
// Discovery is external - callers provide the Job directly.
type JobExecutor interface {
    // Execute runs a job by interpreting its execution plan.
    // Steps are executed in order: stop → worker → start.
    // Returns per-step results and overall status.
    Execute(ctx context.Context, job jobs.Job, opts ExecuteOptions) (ExecutionResult, error)

    // DryRun validates the job and returns the execution plan without executing.
    // Validates worker image exists, job configuration is valid.
    DryRun(ctx context.Context, job jobs.Job) (jobs.ExecutionPlan, error)
}
```

## Changes from Previous Version

| Aspect | Before | After | Rationale |
|--------|--------|-------|-----------|
| `Execute` signature | `jobName string` | `job jobs.Job` | Discovery external (Decision #2) |
| `DryRun` signature | `jobName string` | `job jobs.Job` | Consistency with Execute |
| Unused `ExecuteJob` | Existed as separate method | Removed (renamed to `Execute`) | Interface alignment |

## ExecuteOptions

```go
// ExecuteOptions configures job execution behavior.
type ExecuteOptions struct {
    // DryRun validates without executing (same as calling DryRun method)
    DryRun bool

    // Timeouts for each phase
    StopTimeout  time.Duration  // Default: 30s
    StartTimeout time.Duration  // Default: 30s
    WorkerTimeout time.Duration // Default: 1h

    // KeepWorker preserves worker container after execution (debugging)
    KeepWorker bool
}
```

## ExecutionResult

```go
// ExecutionResult contains the outcome of job execution.
type ExecutionResult struct {
    // JobRun contains job metadata and timing
    JobRun jobs.JobRun

    // Plan that was executed
    Plan jobs.ExecutionPlan

    // StepResults contains per-step outcomes (new: all steps recorded)
    StepResults []StepResult

    // Success is true if all steps completed successfully
    Success bool

    // Error contains the first error encountered (if any)
    Error error
}

// StepResult contains the outcome of a single plan step.
type StepResult struct {
    Step     jobs.PlanStep
    Success  bool
    Duration time.Duration
    Error    error
    // Output contains step-specific data (e.g., worker logs, exit code)
    Output   map[string]interface{}
}
```

## Behavioral Contract

### Execute Method

1. **Preconditions**:
   - `job` is a valid `jobs.Job` with required fields
   - Context is not cancelled
   - Worker image exists (validated before stopping stack)

2. **Postconditions**:
   - All plan steps are attempted (even if earlier steps fail)
   - Stack is always restarted (unless context cancelled)
   - Per-step results are recorded
   - Result reflects actual execution (matches DryRun output)

3. **Error Handling**:
   - Stop fails → Return error, do NOT proceed to worker
   - Worker fails → Log warning, still restart stack
   - Start fails → Return error, stack may be partial
   - Context cancelled → Attempt graceful shutdown, restart stack

### DryRun Method

1. **Preconditions**:
   - `job` is a valid `jobs.Job`

2. **Postconditions**:
   - Worker image existence is validated
   - Complete execution plan is returned (including start step)
   - No side effects (no containers stopped/started)

3. **Error Handling**:
   - Image not found → Return `ErrImageNotFound`
   - Invalid job → Return validation error

## Implementation Requirements

The implementation (`internal/app/executor/executor.go`) MUST:

1. **Interpret plan steps** — Execute steps from `plan.Steps` in order
2. **Record per-step results** — Each step gets a `StepResult`
3. **Match DryRun** — Execute and DryRun produce identical plans
4. **No unused parameters** — Constructor only accepts used dependencies
5. **No stub methods** — All interface methods fully implemented
