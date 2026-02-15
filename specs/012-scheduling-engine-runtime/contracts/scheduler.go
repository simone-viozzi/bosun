// Package contracts defines the port interfaces for M4 Scheduling Engine.
package contracts

import (
	"context"
	"time"
)

// Scheduler manages cron-based job scheduling with config refresh.
// Implementation location: internal/app/scheduler/scheduler.go
type Scheduler interface {
	// Start begins the scheduler and config refresh loop.
	// Blocks until ctx is cancelled or an error occurs.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the scheduler.
	// Waits for running jobs to complete (up to ctx timeout).
	Stop(ctx context.Context) error

	// AddJob registers a job with the cron scheduler.
	// Parses schedule, applies overlap policy, and emits events.
	AddJob(ctx context.Context, job Job) error

	// RemoveJob unregisters a job from the scheduler.
	// Current execution completes if job is running.
	RemoveJob(ctx context.Context, jobName string) error

	// ListJobs returns current status of all scheduled jobs.
	// Thread-safe for concurrent access.
	ListJobs() []JobStatus
}

// SchedulerOptions configures scheduler behavior.
type SchedulerOptions struct {
	// RefreshInterval determines how often to re-discover jobs from Docker labels.
	// Default: 5 minutes
	RefreshInterval time.Duration
}

// Job represents a scheduled job (from domain model).
// This is a reference - actual implementation in internal/domain/jobs/types.go
type Job struct {
	Name          string
	Schedule      string        // Cron expression (e.g., "0 3 * * *")
	OverlapPolicy OverlapPolicy // How to handle concurrent runs
	Enabled       bool          // Whether job is active

	// Execution configuration (existing from M3)
	TargetStacks  []string
	WorkerImage   string
	AttachVolumes []string
	// ... other fields ...
}

// OverlapPolicy defines behavior when job is scheduled while previous run is active.
type OverlapPolicy string

const (
	OverlapPolicyQueue         OverlapPolicy = "queue"              // Delay next run until current completes
	OverlapPolicySkip          OverlapPolicy = "skip"               // Drop next run if current is active
	OverlapPolicyCancelRestart OverlapPolicy = "cancel-and-restart" // Stop current, start fresh (DEFERRED #176)
)

// JobStatus represents current state of a scheduled job.
// This is a reference - actual implementation in internal/domain/jobs/status.go
type JobStatus struct {
	JobName       string
	Status        RunStatus // Current execution state
	Schedule      string    // Cron expression
	OverlapPolicy OverlapPolicy
	LastRunTime   *time.Time // nil if never run
	LastResult    string     // "success" or "error: <message>"
	NextRunTime   time.Time  // Calculated by cron library
	CurrentRunID  string     // UUID if running, empty otherwise
}

// RunStatus represents execution state.
type RunStatus string

const (
	RunStatusIdle      RunStatus = "idle"      // Not currently executing
	RunStatusRunning   RunStatus = "running"   // Currently executing
	RunStatusCompleted RunStatus = "completed" // Last run succeeded
	RunStatusFailed    RunStatus = "failed"    // Last run failed
)

// Example Usage:
//
// Creating and starting a scheduler:
//
//	scheduler := scheduler.New(executor, discoverer, events, stateStore, scheduler.Options{
//	    RefreshInterval: 5 * time.Minute,
//	})
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	if err := scheduler.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// Adding a job manually:
//
//	job := jobs.Job{
//	    Name:          "daily-backup",
//	    Schedule:      "0 3 * * *",
//	    OverlapPolicy: jobs.OverlapPolicyQueue,
//	    Enabled:       true,
//	    WorkerImage:   "myapp/backup-worker:latest",
//	}
//
//	if err := scheduler.AddJob(ctx, job); err != nil {
//	    log.Fatalf("Failed to add job: %v", err)
//	}
//
// Listing jobs:
//
//	jobs := scheduler.ListJobs()
//	for _, job := range jobs {
//	    fmt.Printf("%s: %s (next: %s)\n", job.JobName, job.Status, job.NextRunTime)
//	}
//
// Graceful shutdown:
//
//	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
//	defer cancel()
//
//	if err := scheduler.Stop(shutdownCtx); err != nil {
//	    log.Printf("Shutdown failed: %v", err)
//	}
