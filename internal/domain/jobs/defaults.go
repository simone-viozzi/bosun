// Package jobs defines default constants for job execution.
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
