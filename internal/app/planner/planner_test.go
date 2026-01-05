package planner

import (
	"context"
	"testing"

	"github.com/simone-viozzi/bosun/internal/config/schema"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestPlan_EmptyJob(t *testing.T) {
	p := New()
	job := jobs.Job{
		Name:        "empty-job",
		Schedule:    schema.DefaultJobSchedule(),
		WorkerImage: schema.DefaultJobWorkerImage(),
	}

	plan, err := p.Plan(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.JobName != "empty-job" {
		t.Errorf("JobName = %q, want %q", plan.JobName, "empty-job")
	}

	// Should have only run-worker step (no containers to stop)
	if len(plan.Steps) != 1 {
		t.Fatalf("Steps length = %d, want 1", len(plan.Steps))
	}

	if plan.Steps[0].Type != jobs.StepTypeRunWorker {
		t.Errorf("Steps[0].Type = %q, want %q", plan.Steps[0].Type, jobs.StepTypeRunWorker)
	}
}

func TestPlan_WithContainers(t *testing.T) {
	p := New()
	job := jobs.Job{
		Name:             "backup-job",
		Schedule:         "0 2 * * *",
		WorkerImage:      "backup:v1",
		TargetContainers: []string{"container-a", "container-b"},
		TargetStacks:     []string{"mystack"},
	}

	plan, err := p.Plan(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 steps: stop + run-worker + start
	if len(plan.Steps) != 3 {
		t.Fatalf("Steps length = %d, want 3", len(plan.Steps))
	}

	// First step should be stop_containers
	if plan.Steps[0].Type != jobs.StepTypeStopContainers {
		t.Errorf("Steps[0].Type = %q, want %q", plan.Steps[0].Type, jobs.StepTypeStopContainers)
	}
	if len(plan.Steps[0].ContainerIDs) != 2 {
		t.Errorf("Steps[0].ContainerIDs length = %d, want 2", len(plan.Steps[0].ContainerIDs))
	}

	// Second step should be run_worker
	if plan.Steps[1].Type != jobs.StepTypeRunWorker {
		t.Errorf("Steps[1].Type = %q, want %q", plan.Steps[1].Type, jobs.StepTypeRunWorker)
	}
	if plan.Steps[1].WorkerImage != "backup:v1" {
		t.Errorf("Steps[1].WorkerImage = %q, want %q", plan.Steps[1].WorkerImage, "backup:v1")
	}

	// Third step should be start_containers
	if plan.Steps[2].Type != jobs.StepTypeStartContainers {
		t.Errorf("Steps[2].Type = %q, want %q", plan.Steps[2].Type, jobs.StepTypeStartContainers)
	}
	if len(plan.Steps[2].ContainerIDs) != 2 {
		t.Errorf("Steps[2].ContainerIDs length = %d, want 2", len(plan.Steps[2].ContainerIDs))
	}
}

func TestPlan_WithVolumes(t *testing.T) {
	p := New()
	job := jobs.Job{
		Name:        "backup-job",
		WorkerImage: "backup:v1",
		AttachVolumes: []jobs.VolumeAttachment{
			{Name: "pgdata", MountPath: "/data/postgres", Mode: "ro"},
			{Name: "redis", MountPath: "/data/redis", Mode: "ro"},
		},
	}

	plan, err := p.Plan(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 1 step: run-worker (no containers)
	if len(plan.Steps) != 1 {
		t.Fatalf("Steps length = %d, want 1", len(plan.Steps))
	}

	runStep := plan.Steps[0]
	if len(runStep.VolumeMounts) != 2 {
		t.Errorf("VolumeMounts length = %d, want 2", len(runStep.VolumeMounts))
	}
}

func TestPlan_Determinism(t *testing.T) {
	p := New()
	job := jobs.Job{
		Name:             "deterministic-job",
		Schedule:         "0 0 * * *",
		WorkerImage:      "worker:v1",
		TargetContainers: []string{"z-container", "a-container", "m-container"},
		AttachVolumes: []jobs.VolumeAttachment{
			{Name: "z-volume", MountPath: "/z", Mode: "ro"},
			{Name: "a-volume", MountPath: "/a", Mode: "ro"},
		},
	}

	// Generate plan multiple times
	plan1, err := p.Plan(context.Background(), job)
	if err != nil {
		t.Fatalf("plan1 error: %v", err)
	}

	plan2, err := p.Plan(context.Background(), job)
	if err != nil {
		t.Fatalf("plan2 error: %v", err)
	}

	// Plans should be identical (except CreatedAt)
	if len(plan1.Steps) != len(plan2.Steps) {
		t.Fatalf("Steps lengths differ: %d vs %d", len(plan1.Steps), len(plan2.Steps))
	}

	// Verify container IDs are sorted
	stopStep := plan1.Steps[0]
	if stopStep.ContainerIDs[0] != "a-container" {
		t.Errorf("Containers not sorted: first = %q, want %q", stopStep.ContainerIDs[0], "a-container")
	}
	if stopStep.ContainerIDs[1] != "m-container" {
		t.Errorf("Containers not sorted: second = %q, want %q", stopStep.ContainerIDs[1], "m-container")
	}
	if stopStep.ContainerIDs[2] != "z-container" {
		t.Errorf("Containers not sorted: third = %q, want %q", stopStep.ContainerIDs[2], "z-container")
	}

	// Verify volumes are sorted
	runStep := plan1.Steps[1]
	if runStep.VolumeMounts[0].Name != "a-volume" {
		t.Errorf("Volumes not sorted: first = %q, want %q", runStep.VolumeMounts[0].Name, "a-volume")
	}
	if runStep.VolumeMounts[1].Name != "z-volume" {
		t.Errorf("Volumes not sorted: second = %q, want %q", runStep.VolumeMounts[1].Name, "z-volume")
	}
}

func TestPlan_UseComposeStop(t *testing.T) {
	p := New()

	t.Run("single stack enables compose stop", func(t *testing.T) {
		job := jobs.Job{
			Name:             "compose-job",
			WorkerImage:      "worker:v1",
			TargetContainers: []string{"a", "b"},
			TargetStacks:     []string{"mystack"},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopStep := plan.Steps[0]
		if !stopStep.UseCompose {
			t.Error("UseCompose should be true for single stack")
		}
		if stopStep.ComposeProject != "mystack" {
			t.Errorf("ComposeProject = %q, want %q", stopStep.ComposeProject, "mystack")
		}
	})

	t.Run("multiple stacks disables compose stop", func(t *testing.T) {
		job := jobs.Job{
			Name:             "multi-stack-job",
			WorkerImage:      "worker:v1",
			TargetContainers: []string{"a", "b"},
			TargetStacks:     []string{"stack1", "stack2"},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopStep := plan.Steps[0]
		if stopStep.UseCompose {
			t.Error("UseCompose should be false for multiple stacks")
		}
	})

	t.Run("no stacks disables compose stop", func(t *testing.T) {
		job := jobs.Job{
			Name:             "no-stack-job",
			WorkerImage:      "worker:v1",
			TargetContainers: []string{"a"},
			TargetStacks:     []string{},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopStep := plan.Steps[0]
		if stopStep.UseCompose {
			t.Error("UseCompose should be false when no stacks")
		}
	})
}

func TestPlan_StepDescriptions(t *testing.T) {
	p := New()

	t.Run("stop single container description", func(t *testing.T) {
		job := jobs.Job{
			Name:             "single-container",
			WorkerImage:      "worker:v1",
			TargetContainers: []string{"my-container"},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		desc := plan.Steps[0].Description
		if !containsSubstr(desc, "Stop container") {
			t.Errorf("Description should mention 'Stop container': %q", desc)
		}
	})

	t.Run("run worker with single volume description", func(t *testing.T) {
		job := jobs.Job{
			Name:        "volume-job",
			WorkerImage: "backup:v1",
			AttachVolumes: []jobs.VolumeAttachment{
				{Name: "pgdata", MountPath: "/data", Mode: "ro"},
			},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		desc := plan.Steps[0].Description
		if !containsSubstr(desc, "pgdata") {
			t.Errorf("Description should mention volume name: %q", desc)
		}
		if !containsSubstr(desc, "/data") {
			t.Errorf("Description should mention mount path: %q", desc)
		}
	})

	t.Run("run worker with multiple volumes description", func(t *testing.T) {
		job := jobs.Job{
			Name:        "multi-volume-job",
			WorkerImage: "backup:v1",
			AttachVolumes: []jobs.VolumeAttachment{
				{Name: "vol1", MountPath: "/v1", Mode: "ro"},
				{Name: "vol2", MountPath: "/v2", Mode: "ro"},
				{Name: "vol3", MountPath: "/v3", Mode: "ro"},
			},
		}

		plan, err := p.Plan(context.Background(), job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		desc := plan.Steps[0].Description
		if !containsSubstr(desc, "3 volumes") {
			t.Errorf("Description should mention '3 volumes': %q", desc)
		}
	})
}

func TestPlan_ContextCancellation(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job := jobs.Job{
		Name:        "test",
		WorkerImage: "test:v1",
	}

	_, err := p.Plan(ctx, job)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// containsSubstr is a helper to check if a string contains a substring.
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
