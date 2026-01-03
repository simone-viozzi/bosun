// Package ports defines the WorkerRunner interface.
// This file is a contract definition for implementation in #119.
//
// GitHub Issue: #116
// Spec: specs/009-job-execution-mvp/spec.md
package ports

import (
	"context"
	"time"
)

// WorkerRunner creates and executes worker containers.
// Handles container lifecycle, log capture, and timeout enforcement.
//
// Implementation Notes:
//   - Container naming: bosun-worker-{job-name}-{run-id[0:8]}
//   - Logs: Captured via Docker ContainerLogs API
//   - Timeout: SIGTERM → 10s grace → SIGKILL
//   - Cleanup: Container removed after execution (unless KeepOnFailure)
//
// Worker Contract:
//   - Exit code 0 = success
//   - Exit code non-zero = failure
//   - SIGTERM = graceful shutdown requested
//   - Environment: BOSUN_* vars + user pass-through
type WorkerRunner interface {
	// Run executes a worker container and waits for completion.
	//
	// Behavior:
	//   1. Create container with config
	//   2. Start container
	//   3. Stream logs to result
	//   4. Wait for exit or timeout
	//   5. On timeout: SIGTERM → 10s → SIGKILL
	//   6. Remove container (unless KeepOnFailure and failed)
	//
	// Returns:
	//   - WorkerResult with exit code, logs, duration
	//   - Error if container creation/start fails
	//   - Error is NOT returned for non-zero exit code (check result.ExitCode)
	Run(ctx context.Context, config WorkerConfig) (WorkerResult, error)
}

// WorkerConfig defines worker container configuration.
type WorkerConfig struct {
	// Image is the container image to run (required).
	// Must be available locally or pullable.
	Image string

	// Env contains environment variables.
	// Includes BOSUN_* metadata and user pass-through from labels.
	Env map[string]string

	// Mounts defines volume attachments for the worker.
	Mounts []VolumeMount

	// Timeout is the maximum execution time.
	// Default: 1h (from jobs.DefaultWorkerTimeout)
	// After timeout, SIGTERM is sent followed by SIGKILL.
	Timeout time.Duration

	// KeepOnFailure preserves container on non-zero exit.
	// Useful for debugging failed workers.
	// Default: false (container always removed)
	KeepOnFailure bool

	// RunID is the unique execution ID (UUID v4).
	// Used for container naming and BOSUN_RUN_ID env var.
	RunID string

	// JobName is the job identifier.
	// Used for container naming and BOSUN_JOB_NAME env var.
	JobName string

	// StackName is the target stack name.
	// Used for BOSUN_STACK env var.
	StackName string

	// DryRun indicates if this is a dry-run execution.
	// Used for BOSUN_DRY_RUN env var.
	DryRun bool
}

// VolumeMount defines a volume attachment for the worker.
type VolumeMount struct {
	// Source is the Docker volume name or host path.
	Source string

	// Target is the mount path inside the container.
	// Default: /data/{volume-name}
	Target string

	// ReadOnly mounts the volume as read-only.
	// Default: true (safety first)
	ReadOnly bool
}

// WorkerResult contains execution outcome.
type WorkerResult struct {
	// ExitCode from container.
	// 0 = success, non-zero = failure.
	// Special codes: 137 = SIGKILL (timeout), 143 = SIGTERM.
	ExitCode int

	// Logs captured from stdout/stderr.
	// Combined output, not separated by stream.
	Logs string

	// ContainerID of the executed worker.
	// Empty if container was removed.
	ContainerID string

	// Duration of execution.
	Duration time.Duration

	// TimedOut indicates if execution was terminated due to timeout.
	// When true, ExitCode is typically 137 (SIGKILL) or 143 (SIGTERM).
	TimedOut bool
}

// Success returns true if worker completed successfully.
func (r WorkerResult) Success() bool {
	return r.ExitCode == 0 && !r.TimedOut
}

// DefaultWorkerConfig returns a config with default values filled in.
// Caller must still set Image, RunID, JobName, and StackName.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Env:           make(map[string]string),
		Mounts:        nil,
		Timeout:       1 * time.Hour,
		KeepOnFailure: false,
		DryRun:        false,
	}
}

// BuildEnv constructs the full environment for the worker container.
// Adds BOSUN_* variables to any existing env vars.
func (c *WorkerConfig) BuildEnv() []string {
	env := make([]string, 0, len(c.Env)+4)

	// Bosun metadata
	env = append(env,
		"BOSUN_JOB_NAME="+c.JobName,
		"BOSUN_RUN_ID="+c.RunID,
		"BOSUN_STACK="+c.StackName,
	)
	if c.DryRun {
		env = append(env, "BOSUN_DRY_RUN=true")
	} else {
		env = append(env, "BOSUN_DRY_RUN=false")
	}

	// User pass-through vars
	for k, v := range c.Env {
		env = append(env, k+"="+v)
	}

	return env
}
