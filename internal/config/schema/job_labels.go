package schema

import (
	"strings"
	"time"
)

// Job label schema defines the Docker labels used to configure backup jobs.
// Labels are applied to containers and volumes to define how Bosun discovers
// and executes backup jobs.

// JobLabelConfig defines the configuration labels that can be applied to containers
// to define backup jobs.
type JobLabelConfig struct {
	// Enabled indicates whether this container participates in a backup job.
	// Containers without this label or with enabled=false are ignored.
	Enabled bool `bosun:"key=bosun.job.enabled,scope=container,type=bool,default=false,doc='Enable job participation for this container'"`

	// Name is the unique identifier for the backup job.
	// Multiple containers with the same name are merged into one job.
	// Required when enabled=true.
	Name string `bosun:"key=bosun.job.name,scope=container,type=string,doc='Unique job identifier'"`

	// Schedule is a cron expression defining when the job runs.
	// Uses standard 5-field cron format: minute hour day-of-month month day-of-week.
	Schedule string `bosun:"key=bosun.job.schedule,scope=container,type=string,default='0 0 * * *',doc='Cron schedule expression (default: daily at midnight)'"`

	// WorkerImage specifies the Docker image to use for executing the backup.
	WorkerImage string `bosun:"key=bosun.job.worker.image,scope=container,type=string,default=bosun-worker:local,doc='Docker image for backup worker'"`

	// M3 additions: Timeout configurations
	// StopTimeout specifies the timeout for stopping containers before worker execution.
	StopTimeout time.Duration `bosun:"key=bosun.job.stop-timeout,scope=container,type=duration,default=30s,doc='Timeout for stopping each container (default: 30s)'"`

	// StartTimeout specifies the timeout for starting containers after worker execution.
	StartTimeout time.Duration `bosun:"key=bosun.job.start-timeout,scope=container,type=duration,default=30s,doc='Timeout for starting each container (default: 30s)'"`

	// WorkerTimeout specifies the maximum execution time for the worker container.
	WorkerTimeout time.Duration `bosun:"key=bosun.job.timeout,scope=container,type=duration,default=1h,doc='Timeout for worker execution (default: 1h)'"`
}

// JobVolumeConfig defines the configuration labels that can be applied to volumes
// to attach them to backup jobs.
type JobVolumeConfig struct {
	// Attach specifies which job(s) this volume should be attached to.
	// The value is a job name (matching bosun.job.name on containers).
	Attach string `bosun:"key=bosun.job.attach,scope=volume,type=string,doc='Job name to attach this volume to'"`

	// MountPath specifies where the volume should be mounted in the worker container.
	MountPath string `bosun:"key=bosun.job.mount.path,scope=volume,type=string,doc='Mount path in worker container'"`

	// Mode specifies the mount mode: 'ro' (read-only) or 'rw' (read-write).
	Mode string `bosun:"key=bosun.job.mount.mode,scope=volume,type=enum,enum=ro|rw,default=ro,doc='Volume mount mode (ro=read-only, rw=read-write)'"`
}

// JobSpec returns the parsed specification for job-related labels.
// This includes both container labels (JobLabelConfig) and volume labels (JobVolumeConfig).
func JobSpec() Spec {
	containerSpec, err := ParseTags[JobLabelConfig]()
	if err != nil {
		panic("failed to parse JobLabelConfig tags: " + err.Error())
	}

	volumeSpec, err := ParseTags[JobVolumeConfig]()
	if err != nil {
		panic("failed to parse JobVolumeConfig tags: " + err.Error())
	}

	// Merge both specs
	combined := make(Spec, len(containerSpec)+len(volumeSpec))
	for k, v := range containerSpec {
		combined[k] = v
	}
	for k, v := range volumeSpec {
		combined[k] = v
	}

	return combined
}

// Label key constants extracted from JobLabelConfig and JobVolumeConfig struct tags.
// These are the canonical source of truth for job label keys.
const (
	// Container job labels
	LabelJobEnabled     = "bosun.job.enabled"
	LabelJobName        = "bosun.job.name"
	LabelJobSchedule    = "bosun.job.schedule"
	LabelJobWorkerImage = "bosun.job.worker.image"

	// M3 additions: Timeout labels
	LabelJobStopTimeout   = "bosun.job.stop-timeout"
	LabelJobStartTimeout  = "bosun.job.start-timeout"
	LabelJobWorkerTimeout = "bosun.job.timeout"

	// Worker environment variables (prefix-based)
	LabelJobWorkerEnvPrefix = "bosun.job.worker.env."

	// Volume job labels
	LabelJobAttach    = "bosun.job.attach"
	LabelJobMountPath = "bosun.job.mount.path"
	LabelJobMountMode = "bosun.job.mount.mode"
)

// DefaultJobSchedule returns the default cron schedule for jobs.
// Value extracted from JobLabelConfig struct tag.
func DefaultJobSchedule() string {
	return "0 0 * * *"
}

// DefaultJobWorkerImage returns the default worker image for jobs.
// Value extracted from JobLabelConfig struct tag.
func DefaultJobWorkerImage() string {
	return "bosun-worker:local"
}

// DefaultJobMountMode returns the default mount mode for job volumes.
// Value extracted from JobVolumeConfig struct tag.
func DefaultJobMountMode() string {
	return "ro"
}

// ValidMountModes contains the allowed values for bosun.job.mount.mode.
var ValidMountModes = []string{"ro", "rw"}

// NormalizeMountMode normalizes mount mode to lowercase.
// Returns the normalized value and true if valid, or empty string and false if invalid.
func NormalizeMountMode(mode string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	for _, valid := range ValidMountModes {
		if normalized == valid {
			return normalized, true
		}
	}
	return "", false
}
