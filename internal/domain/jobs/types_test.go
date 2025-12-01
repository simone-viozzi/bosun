package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

func TestJobDefaults(t *testing.T) {
	t.Run("DefaultSchedule is valid cron expression", func(t *testing.T) {
		// Default schedule should be midnight daily
		if schema.DefaultJobSchedule() != "0 0 * * *" {
			t.Errorf("DefaultJobSchedule() = %q, want %q", schema.DefaultJobSchedule(), "0 0 * * *")
		}
	})

	t.Run("DefaultWorkerImage is set", func(t *testing.T) {
		if schema.DefaultJobWorkerImage() == "" {
			t.Error("DefaultJobWorkerImage() should not be empty")
		}
		if schema.DefaultJobWorkerImage() != "bosun-worker:local" {
			t.Errorf("DefaultJobWorkerImage() = %q, want %q", schema.DefaultJobWorkerImage(), "bosun-worker:local")
		}
	})

	t.Run("DefaultMountMode is read-only", func(t *testing.T) {
		if schema.DefaultJobMountMode() != "ro" {
			t.Errorf("DefaultJobMountMode() = %q, want %q", schema.DefaultJobMountMode(), "ro")
		}
	})
}

func TestStepTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		stepType StepType
		want     string
	}{
		{"StopContainers", StepTypeStopContainers, "stop_containers"},
		{"RunWorker", StepTypeRunWorker, "run_worker"},
		{"StartContainers", StepTypeStartContainers, "start_containers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.stepType) != tt.want {
				t.Errorf("StepType %s = %q, want %q", tt.name, tt.stepType, tt.want)
			}
		})
	}
}

func TestJobSerialization(t *testing.T) {
	job := Job{
		Name:             "daily-backup",
		Schedule:         "0 2 * * *",
		TargetContainers: []string{"postgres", "redis"},
		TargetStacks:     []string{"mystack"},
		WorkerImage:      "backup-worker:v1",
		AttachVolumes: []VolumeAttachment{
			{Name: "data-vol", MountPath: "/data", Mode: "ro"},
		},
		SourceContainers: []string{"postgres-container-abc123"},
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal Job: %v", err)
	}

	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Job: %v", err)
	}

	if decoded.Name != job.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, job.Name)
	}
	if decoded.Schedule != job.Schedule {
		t.Errorf("Schedule = %q, want %q", decoded.Schedule, job.Schedule)
	}
	if len(decoded.TargetContainers) != len(job.TargetContainers) {
		t.Errorf("TargetContainers length = %d, want %d", len(decoded.TargetContainers), len(job.TargetContainers))
	}
	if len(decoded.AttachVolumes) != 1 {
		t.Fatalf("AttachVolumes length = %d, want 1", len(decoded.AttachVolumes))
	}
	if decoded.AttachVolumes[0].Name != "data-vol" {
		t.Errorf("AttachVolumes[0].Name = %q, want %q", decoded.AttachVolumes[0].Name, "data-vol")
	}
}

func TestExecutionPlanSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	plan := ExecutionPlan{
		JobName:   "test-job",
		CreatedAt: now,
		Steps: []PlanStep{
			{
				Type:           StepTypeStopContainers,
				Description:    "Stop containers",
				ContainerIDs:   []string{"abc123"},
				ContainerNames: []string{"postgres"},
			},
			{
				Type:         StepTypeRunWorker,
				Description:  "Run backup",
				WorkerImage:  "backup:latest",
				VolumeMounts: []VolumeAttachment{{Name: "data", MountPath: "/backup", Mode: "ro"}},
			},
			{
				Type:           StepTypeStartContainers,
				Description:    "Start containers",
				ContainerIDs:   []string{"abc123"},
				ContainerNames: []string{"postgres"},
			},
		},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("failed to marshal ExecutionPlan: %v", err)
	}

	var decoded ExecutionPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ExecutionPlan: %v", err)
	}

	if decoded.JobName != plan.JobName {
		t.Errorf("JobName = %q, want %q", decoded.JobName, plan.JobName)
	}
	if !decoded.CreatedAt.Equal(plan.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", decoded.CreatedAt, plan.CreatedAt)
	}
	if len(decoded.Steps) != 3 {
		t.Fatalf("Steps length = %d, want 3", len(decoded.Steps))
	}
	if decoded.Steps[0].Type != StepTypeStopContainers {
		t.Errorf("Steps[0].Type = %q, want %q", decoded.Steps[0].Type, StepTypeStopContainers)
	}
	if decoded.Steps[1].WorkerImage != "backup:latest" {
		t.Errorf("Steps[1].WorkerImage = %q, want %q", decoded.Steps[1].WorkerImage, "backup:latest")
	}
}

func TestVolumeAttachmentSerialization(t *testing.T) {
	vol := VolumeAttachment{
		Name:      "pgdata",
		MountPath: "/var/lib/postgresql/data",
		Mode:      "ro",
	}

	data, err := json.Marshal(vol)
	if err != nil {
		t.Fatalf("failed to marshal VolumeAttachment: %v", err)
	}

	var decoded VolumeAttachment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal VolumeAttachment: %v", err)
	}

	if decoded.Name != vol.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, vol.Name)
	}
	if decoded.MountPath != vol.MountPath {
		t.Errorf("MountPath = %q, want %q", decoded.MountPath, vol.MountPath)
	}
	if decoded.Mode != vol.Mode {
		t.Errorf("Mode = %q, want %q", decoded.Mode, vol.Mode)
	}
}

func TestStackSerialization(t *testing.T) {
	stack := Stack{
		Name:       "myapp",
		Containers: []string{"web", "db", "cache"},
		Volumes:    []string{"db-data", "cache-data"},
		Networks:   []string{"frontend", "backend"},
	}

	data, err := json.Marshal(stack)
	if err != nil {
		t.Fatalf("failed to marshal Stack: %v", err)
	}

	var decoded Stack
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Stack: %v", err)
	}

	if decoded.Name != stack.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, stack.Name)
	}
	if len(decoded.Containers) != 3 {
		t.Errorf("Containers length = %d, want 3", len(decoded.Containers))
	}
	if len(decoded.Volumes) != 2 {
		t.Errorf("Volumes length = %d, want 2", len(decoded.Volumes))
	}
	if len(decoded.Networks) != 2 {
		t.Errorf("Networks length = %d, want 2", len(decoded.Networks))
	}
}

func TestContainerDependencySerialization(t *testing.T) {
	dep := ContainerDependency{
		ServiceName: "postgres",
		Condition:   "service_healthy",
		Required:    true,
	}

	data, err := json.Marshal(dep)
	if err != nil {
		t.Fatalf("failed to marshal ContainerDependency: %v", err)
	}

	var decoded ContainerDependency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ContainerDependency: %v", err)
	}

	if decoded.ServiceName != dep.ServiceName {
		t.Errorf("ServiceName = %q, want %q", decoded.ServiceName, dep.ServiceName)
	}
	if decoded.Condition != dep.Condition {
		t.Errorf("Condition = %q, want %q", decoded.Condition, dep.Condition)
	}
	if decoded.Required != dep.Required {
		t.Errorf("Required = %v, want %v", decoded.Required, dep.Required)
	}
}

func TestPlanStepOmitsEmptyFields(t *testing.T) {
	// A stop_containers step should omit worker-related fields
	step := PlanStep{
		Type:           StepTypeStopContainers,
		Description:    "Stop containers for backup",
		ContainerIDs:   []string{"abc123"},
		ContainerNames: []string{"postgres"},
		// WorkerImage, VolumeMounts left empty
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("failed to marshal PlanStep: %v", err)
	}

	jsonStr := string(data)

	// Check that empty fields are omitted
	if contains(jsonStr, "worker_image") {
		t.Error("JSON should not contain worker_image when empty")
	}
	if contains(jsonStr, "volume_mounts") {
		t.Error("JSON should not contain volume_mounts when empty")
	}
	if contains(jsonStr, "use_compose_stop") {
		t.Error("JSON should not contain use_compose_stop when false")
	}

	// Check that required fields are present
	if !contains(jsonStr, "type") {
		t.Error("JSON should contain type field")
	}
	if !contains(jsonStr, "description") {
		t.Error("JSON should contain description field")
	}
}

// contains is a helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
