// Package ports defines the JobExecutor interface.
// This file is a contract definition for implementation in #121.
//
// GitHub Issue: #114
// Spec: specs/009-job-execution-mvp/spec.md
package ports

import (
	"context"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

// JobExecutor orchestrates the full job execution lifecycle.
//
// Execution Flow:
//  1. Discover job by name (via JobDiscoverer)
//  2. Generate execution plan (via JobPlanner)
//  3. Pre-validate worker image (fail fast)
//  4. Stop target stack (via ComposeController)
//  5. Run worker container (via WorkerRunner)
//  6. Restart stack (via ComposeController) - ALWAYS, even on worker failure
//
// Signal Handling:
//   - On Ctrl+C (SIGINT) or SIGTERM: Cancel in-flight operations
//   - Always attempt to restart stack before exiting
//
// Error Handling:
//   - Pre-validation failure: Return error, stack unchanged
//   - Stop failure: Return error, stack may be partially stopped
//   - Worker failure: Restart stack anyway, return worker error
//   - Start failure: Return error, stack may be partially started
type JobExecutor interface {
	// Execute runs a job by name.
	//
	// Behavior:
	//   1. Look up job by name
	//   2. Validate worker image is available
	//   3. Stop target stack
	//   4. Run worker container
	//   5. Restart stack (always, unless KeepStopped)
	//
	// Returns:
	//   - ExecutionResult with full execution details
	//   - Error for fatal failures (not for worker non-zero exit)
	Execute(ctx context.Context, jobName string, opts ExecuteOptions) (ExecutionResult, error)

	// DryRun returns the execution plan without performing any actions.
	// Useful for previewing what Execute would do.
	//
	// Behavior:
	//   1. Look up job by name
	//   2. Generate execution plan
	//   3. Return plan without execution
	//
	// Returns:
	//   - ExecutionPlan describing planned steps
	//   - Error if job not found or plan generation fails
	DryRun(ctx context.Context, jobName string) (jobs.ExecutionPlan, error)
}

// ExecuteOptions configures job execution.
type ExecuteOptions struct {
	// TimeoutOverride overrides the worker timeout.
	// Zero value means use job's configured timeout or default.
	// Corresponds to --timeout CLI flag.
	TimeoutOverride time.Duration

	// StopTimeoutOverride overrides the stack stop timeout.
	// Zero value means use job's configured timeout or default.
	// Corresponds to --stop-timeout CLI flag.
	StopTimeoutOverride time.Duration

	// StartTimeoutOverride overrides the stack start timeout.
	// Zero value means use job's configured timeout or default.
	// Corresponds to --start-timeout CLI flag.
	StartTimeoutOverride time.Duration

	// KeepStopped skips stack restart after worker completion.
	// Useful for maintenance mode or debugging.
	// Default: false (always restart)
	// Corresponds to --keep-stopped CLI flag.
	KeepStopped bool

	// KeepFailedWorker preserves worker container on non-zero exit.
	// Useful for debugging failed workers.
	// Default: false (always remove)
	// Corresponds to --keep-failed CLI flag.
	KeepFailedWorker bool

	// Quiet suppresses log streaming during execution.
	// Logs are still captured in result.
	// Default: false
	// Corresponds to --quiet CLI flag.
	Quiet bool

	// LogWriter receives real-time log output if not nil.
	// Ignored if Quiet is true.
	LogWriter interface {
		Write([]byte) (int, error)
	}
}

// DefaultExecuteOptions returns options with default values.
func DefaultExecuteOptions() ExecuteOptions {
	return ExecuteOptions{
		KeepStopped:      false,
		KeepFailedWorker: false,
		Quiet:            false,
	}
}

// ExecutionResult is the outcome of a job execution.
type ExecutionResult struct {
	// Run contains execution metadata (ID, status, timing).
	Run jobs.JobRun

	// Plan that was executed.
	Plan jobs.ExecutionPlan

	// StepResults for each executed step.
	StepResults []StepResult

	// WorkerLogs captured from worker container.
	// May be empty if worker didn't run or Quiet was true.
	WorkerLogs string
}

// Success returns true if the job completed successfully.
// All steps must have succeeded and worker must have exit code 0.
func (r ExecutionResult) Success() bool {
	return r.Run.Status == jobs.RunStatusSuccess && r.Run.WorkerExitCode == 0
}

// StepResult is the outcome of a single execution step.
type StepResult struct {
	// Step that was executed.
	Step jobs.PlanStep

	// Status of this step.
	Status jobs.RunStatus

	// StartedAt is when the step began.
	StartedAt time.Time

	// Duration of this step.
	Duration time.Duration

	// Error message if step failed.
	Error string

	// Details contains step-specific output.
	// For stop/start: list of container names affected.
	// For worker: container ID.
	Details string
}

// StepSuccess returns true if this step completed successfully.
func (r StepResult) StepSuccess() bool {
	return r.Status == jobs.RunStatusSuccess
}
