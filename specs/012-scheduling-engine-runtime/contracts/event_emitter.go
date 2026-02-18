//go:build ignore

// Package contracts defines the port interfaces for M4 Scheduling Engine.
// These are reference implementations that match the actual ports in internal/ports/.
package contracts

import (
	"context"
	"time"
)

// EventEmitter emits job lifecycle events for observability.
// In M4, only a log-based adapter is implemented (LogEmitter).
// Notification adapters (Slack, email, webhook) are deferred to M5+.
//
// Implementation location: internal/ports/executor.go (interface)
// Default adapter: internal/adapters/events/log_emitter.go
type EventEmitter interface {
	// EmitJobScheduled is called when job is registered with scheduler.
	EmitJobScheduled(ctx context.Context, jobName string, schedule string, nextRun time.Time)

	// EmitJobStarted is called when job execution begins.
	EmitJobStarted(ctx context.Context, jobName string, runID string)

	// EmitJobCompleted is called when job finishes successfully.
	EmitJobCompleted(ctx context.Context, jobName string, runID string, duration time.Duration)

	// EmitJobFailed is called when job execution fails.
	EmitJobFailed(ctx context.Context, jobName string, runID string, err error, duration time.Duration)

	// EmitJobSkipped is called when scheduled run is skipped (overlap policy).
	EmitJobSkipped(ctx context.Context, jobName string, reason string)

	// EmitJobAdded is called when config refresh discovers new job.
	EmitJobAdded(ctx context.Context, jobName string)

	// EmitJobRemoved is called when config refresh detects removed job.
	EmitJobRemoved(ctx context.Context, jobName string)

	// EmitJobChanged is called when job schedule/policy changes.
	EmitJobChanged(ctx context.Context, jobName string, oldSchedule, newSchedule string)

	// EmitJobCircuitBroken is called when a job is auto-disabled after consecutive failures.
	// The job will not be re-enabled by config refresh; manual intervention is required.
	EmitJobCircuitBroken(ctx context.Context, jobName string, consecutiveFailures int)

	// EmitJobDuplicateName is called when duplicate job names detected during discovery.
	// First-seen definition is used; subsequent duplicates are ignored.
	EmitJobDuplicateName(ctx context.Context, jobName string, containerID string)
}
