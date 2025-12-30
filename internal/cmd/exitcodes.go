// Package cmd provides the CLI command implementations for Bosun.
package cmd

import (
	"errors"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

// Exit codes for Bosun CLI commands.
//
// These constants ensure consistent exit code semantics across all commands:
//   - ExitSuccess: Command completed without errors
//   - ExitRuntimeError: Docker unavailable, I/O failure, network error
//   - ExitValidationError: Invalid labels, missing required fields, config errors
//
// M3 Job Execution Exit Codes (10-16 range to avoid collision with existing):
//   - ExitWorkerFailed: Worker exited with non-zero code
//   - ExitStopFailed: Failed to stop stack
//   - ExitStartFailed: Failed to restart stack
//   - ExitTimeout: Operation timed out
//   - ExitImageNotFound: Worker image not found
//   - ExitJobNotFound: Job name not found
//   - ExitInterrupted: Execution interrupted (Ctrl+C)
const (
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
