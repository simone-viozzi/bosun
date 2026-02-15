package jobs

import "time"

// ScheduleStatus represents the scheduler-level state of a job.
// This is distinct from RunStatus, which tracks individual execution runs.
type ScheduleStatus string

const (
	// ScheduleStatusIdle means the job is registered but not currently executing.
	ScheduleStatusIdle ScheduleStatus = "idle"

	// ScheduleStatusRunning means the job is currently executing.
	ScheduleStatusRunning ScheduleStatus = "running"

	// ScheduleStatusCompleted means the last run succeeded.
	ScheduleStatusCompleted ScheduleStatus = "completed"

	// ScheduleStatusFailed means the last run failed.
	ScheduleStatusFailed ScheduleStatus = "failed"
)

// JobStatus represents the current state of a scheduled job as tracked by the Scheduler.
// Returned by Scheduler.ListJobs() for the `bosun job list` command.
type JobStatus struct {
	// JobName is the unique job identifier (from bosun.job.name).
	JobName string `json:"jobName"`

	// Status is the current scheduler-level state.
	Status ScheduleStatus `json:"status"`

	// Schedule is the cron expression.
	Schedule string `json:"schedule"`

	// OverlapPolicy is the configured overlap behavior.
	OverlapPolicy OverlapPolicy `json:"overlapPolicy"`

	// LastRunTime records when the job last completed execution.
	// nil if the job has never run.
	LastRunTime *time.Time `json:"lastRunTime,omitempty"`

	// LastResult is "success" or "error: <message>".
	LastResult string `json:"lastResult,omitempty"`

	// NextRunTime is the next scheduled execution time, calculated by the cron library.
	NextRunTime time.Time `json:"nextRunTime"`

	// CurrentRunID is a UUID when running, empty otherwise.
	CurrentRunID string `json:"currentRunID,omitempty"`
}
