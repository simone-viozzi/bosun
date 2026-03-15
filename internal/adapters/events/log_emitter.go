package events

import (
	"context"
	"log/slog"
	"time"
)

// LogEmitter implements ports.EventEmitter by writing structured logs.
// This is the default M4 adapter; notification adapters (Slack, email)
// are deferred to M5+.
type LogEmitter struct {
	logger *slog.Logger
}

// NewLogEmitter creates a new log-based event emitter.
func NewLogEmitter(logger *slog.Logger) *LogEmitter {
	return &LogEmitter{logger: logger}
}

func (e *LogEmitter) EmitJobScheduled(ctx context.Context, jobName string, schedule string, nextRun time.Time) {
	e.logger.InfoContext(ctx, "Job scheduled",
		"event", "job.scheduled",
		"job_name", jobName,
		"schedule", schedule,
		"next_run", nextRun,
	)
}

func (e *LogEmitter) EmitJobStarted(ctx context.Context, jobName string, runID string) {
	e.logger.InfoContext(ctx, "Job started",
		"event", "job.started",
		"job_name", jobName,
		"run_id", runID,
	)
}

func (e *LogEmitter) EmitJobCompleted(ctx context.Context, jobName string, runID string, duration time.Duration) {
	e.logger.InfoContext(ctx, "Job completed",
		"event", "job.completed",
		"job_name", jobName,
		"run_id", runID,
		"duration", duration,
	)
}

func (e *LogEmitter) EmitJobFailed(ctx context.Context, jobName string, runID string, err error, duration time.Duration) {
	e.logger.ErrorContext(ctx, "Job failed",
		"event", "job.failed",
		"job_name", jobName,
		"run_id", runID,
		"error", err,
		"duration", duration,
	)
}

func (e *LogEmitter) EmitJobSkipped(ctx context.Context, jobName string, reason string) {
	e.logger.WarnContext(ctx, "Job skipped",
		"event", "job.skipped",
		"job_name", jobName,
		"reason", reason,
	)
}

func (e *LogEmitter) EmitJobAdded(ctx context.Context, jobName string) {
	e.logger.InfoContext(ctx, "Job added",
		"event", "job.added",
		"job_name", jobName,
	)
}

func (e *LogEmitter) EmitJobRemoved(ctx context.Context, jobName string) {
	e.logger.InfoContext(ctx, "Job removed",
		"event", "job.removed",
		"job_name", jobName,
	)
}

func (e *LogEmitter) EmitJobChanged(ctx context.Context, jobName string, oldSchedule, newSchedule string) {
	e.logger.InfoContext(ctx, "Job changed",
		"event", "job.changed",
		"job_name", jobName,
		"old_schedule", oldSchedule,
		"new_schedule", newSchedule,
	)
}

func (e *LogEmitter) EmitJobCircuitBroken(ctx context.Context, jobName string, consecutiveFailures int) {
	e.logger.ErrorContext(ctx, "Job circuit-broken (auto-disabled)",
		"event", "job.circuit_broken",
		"job_name", jobName,
		"consecutive_failures", consecutiveFailures,
	)
}

func (e *LogEmitter) EmitJobDuplicateName(ctx context.Context, jobName string, containerID string) {
	e.logger.WarnContext(ctx, "Duplicate job name detected",
		"event", "job.duplicate_name",
		"job_name", jobName,
		"container_id", containerID,
	)
}
