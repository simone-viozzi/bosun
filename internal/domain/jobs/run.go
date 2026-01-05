// Package jobs contains domain types and logic for job execution.
package jobs

import (
	"time"
)

// RunStatus represents the state of a job execution.
type RunStatus string

// Execution status constants.
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

// Success returns true if the execution was successful.
func (r *ExecutionResult) Success() bool {
	return r.Run.Status == RunStatusSuccess
}

// StepSuccess returns true if all steps completed successfully.
func (r *ExecutionResult) StepSuccess() bool {
	for _, step := range r.StepResults {
		if step.Status != RunStatusSuccess {
			return false
		}
	}
	return true
}
