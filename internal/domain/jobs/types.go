package jobs

import "time"

// StepType identifies the kind of action in a plan step.
type StepType string

const (
	// StepTypeStopContainers stops the target containers.
	StepTypeStopContainers StepType = "stop_containers"

	// StepTypeRunWorker runs the worker container with attached volumes.
	StepTypeRunWorker StepType = "run_worker"

	// StepTypeStartContainers restarts the stopped containers.
	// Note: Out of scope for Milestone 2; included for forward compatibility.
	StepTypeStartContainers StepType = "start_containers"
)

// Job represents a discovered job assembled from container labels.
// A job can be contributed to by multiple containers (merged by name).
type Job struct {
	// Name is the unique identifier for this job (from bosun.job.name).
	// Required field.
	Name string `json:"name"`

	// Schedule is a validated cron expression (from bosun.job.schedule).
	// Default: "0 0 * * *" (daily at midnight).
	Schedule string `json:"schedule"`

	// TargetContainers lists container IDs that participate in this job.
	// These containers will be stopped before the worker runs.
	TargetContainers []string `json:"targetContainers"`

	// TargetStacks lists unique stack names derived from TargetContainers.
	// Used for display/grouping purposes only; actual stop is per-container.
	TargetStacks []string `json:"targetStacks"`

	// WorkerImage specifies the container image for executing the job.
	// Default: "bosun-worker:local" (placeholder for this milestone).
	WorkerImage string `json:"workerImage"`

	// AttachVolumes lists volumes to mount in the worker container.
	// Discovered from volumes with bosun.job.attach=<this-job> label.
	AttachVolumes []VolumeAttachment `json:"attachVolumes"`

	// SourceContainers tracks which containers contributed to this job.
	// Used for traceability and conflict detection.
	SourceContainers []string `json:"sourceContainers"`
}

// VolumeAttachment represents a volume attached to a job.
type VolumeAttachment struct {
	// Name is the Docker volume name.
	Name string `json:"name"`

	// MountPath is the path where the volume is mounted in the worker.
	// Default: "/data/<volume-name>" if not specified.
	MountPath string `json:"mountPath"`

	// Mode is the mount access mode: "ro" (read-only) or "rw" (read-write).
	// Default: "ro" for safety.
	Mode string `json:"mode"`
}

// ExecutionPlan represents the computed steps to execute a job.
// Plans are deterministic: same inputs produce identical outputs.
type ExecutionPlan struct {
	// JobName references the originating job.
	JobName string `json:"jobName"`

	// Steps contains ordered actions to execute.
	Steps []PlanStep `json:"steps"`

	// CreatedAt records when the plan was generated.
	// Set by the caller, not the planner (for determinism).
	CreatedAt time.Time `json:"createdAt"`
}

// PlanStep represents a single action in an execution plan.
type PlanStep struct {
	// Type identifies the kind of action.
	Type StepType `json:"type"`

	// Description is a human-readable explanation of the step.
	Description string `json:"description"`

	// ContainerIDs lists containers affected by this step (for stop/start).
	// Empty for run_worker steps.
	ContainerIDs []string `json:"containerIds,omitempty"`

	// ContainerNames provides human-readable names for ContainerIDs.
	ContainerNames []string `json:"containerNames,omitempty"`

	// WorkerImage is the image to run (for run_worker steps only).
	WorkerImage string `json:"workerImage,omitempty"`

	// VolumeMounts lists volumes to attach (for run_worker steps only).
	VolumeMounts []VolumeAttachment `json:"volumeMounts,omitempty"`

	// UseComposeStop indicates if `docker compose stop` should be used.
	// True when all containers in a stack are being stopped.
	UseComposeStop bool `json:"useComposeStop,omitempty"`

	// ComposeProject is the project name for compose commands.
	ComposeProject string `json:"composeProject,omitempty"`
}

// Stack represents a logical grouping of Docker entities.
// Determined by bosun.stack label or com.docker.compose.project metadata.
type Stack struct {
	// Name is the stack identifier.
	Name string `json:"name"`

	// Containers lists container IDs in this stack.
	Containers []string `json:"containers"`

	// Volumes lists volume names associated with this stack.
	Volumes []string `json:"volumes"`

	// Networks lists network names associated with this stack.
	Networks []string `json:"networks"`
}

// ContainerDependency represents a parsed dependency from
// the com.docker.compose.depends_on container label.
// Format: <service>:<condition>:<required>,...
type ContainerDependency struct {
	// ServiceName is the dependency service name.
	ServiceName string `json:"serviceName"`

	// Condition is the start condition (service_started, service_healthy, etc.).
	Condition string `json:"condition"`

	// Required indicates if the dependency is required.
	Required bool `json:"required"`
}
