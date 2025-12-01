package joblabels

import (
	"context"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
	"github.com/simone-viozzi/bosun/internal/domain/labels"
)

func TestNewDiscoverer(t *testing.T) {
	d := NewDiscoverer()
	if d == nil {
		t.Fatal("NewDiscoverer() returned nil")
	}
}

func TestDiscoverJobs_EmptySnapshot(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{},
		TakenAt:  time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(foundJobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(foundJobs))
	}
	if len(validationErrors) != 0 {
		t.Errorf("expected 0 validation errors, got %d", len(validationErrors))
	}
}

func TestDiscoverJobs_SingleContainer(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-123",
				Name: "postgres",
				Labels: map[string]string{
					schema.LabelJobEnabled:     "true",
					schema.LabelJobName:        "daily-backup",
					schema.LabelJobSchedule:    "0 2 * * *",
					schema.LabelJobWorkerImage: "backup:v1",
				},
				Meta: map[string]string{
					"compose.project": "myapp",
				},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("expected 0 validation errors, got %d: %v", len(validationErrors), validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}

	job := foundJobs[0]
	if job.Name != "daily-backup" {
		t.Errorf("Name = %q, want %q", job.Name, "daily-backup")
	}
	if job.Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want %q", job.Schedule, "0 2 * * *")
	}
	if job.WorkerImage != "backup:v1" {
		t.Errorf("WorkerImage = %q, want %q", job.WorkerImage, "backup:v1")
	}
	if len(job.TargetContainers) != 1 || job.TargetContainers[0] != "container-123" {
		t.Errorf("TargetContainers = %v, want [container-123]", job.TargetContainers)
	}
	if len(job.TargetStacks) != 1 || job.TargetStacks[0] != "myapp" {
		t.Errorf("TargetStacks = %v, want [myapp]", job.TargetStacks)
	}
}

func TestDiscoverJobs_MergeMultipleContainers(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "postgres-1",
				Name: "postgres",
				Labels: map[string]string{
					schema.LabelJobEnabled:  "true",
					schema.LabelJobName:     "daily-backup",
					schema.LabelJobSchedule: "0 2 * * *",
				},
				Meta: map[string]string{"compose.project": "myapp"},
			},
			{
				Kind: labels.KindContainer,
				ID:   "redis-1",
				Name: "redis",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "daily-backup",
				},
				Meta: map[string]string{"compose.project": "myapp"},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 merged job, got %d", len(foundJobs))
	}

	job := foundJobs[0]
	if len(job.TargetContainers) != 2 {
		t.Errorf("TargetContainers length = %d, want 2", len(job.TargetContainers))
	}
	if len(job.SourceContainers) != 2 {
		t.Errorf("SourceContainers length = %d, want 2", len(job.SourceContainers))
	}
}

func TestDiscoverJobs_DefaultValues(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "minimal-job",
					// No schedule or worker image - should use defaults
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}

	job := foundJobs[0]
	if job.Schedule != schema.DefaultJobSchedule() {
		t.Errorf("Schedule = %q, want default %q", job.Schedule, schema.DefaultJobSchedule())
	}
	if job.WorkerImage != schema.DefaultJobWorkerImage() {
		t.Errorf("WorkerImage = %q, want default %q", job.WorkerImage, schema.DefaultJobWorkerImage())
	}
}

func TestDiscoverJobs_MissingJobName(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					// Missing bosun.job.name
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(foundJobs) != 0 {
		t.Errorf("expected 0 jobs when name is missing, got %d", len(foundJobs))
	}
	if len(validationErrors) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(validationErrors))
	}

	ve := validationErrors[0]
	if ve.Field != schema.LabelJobName {
		t.Errorf("Field = %q, want %q", ve.Field, schema.LabelJobName)
	}
	if ve.EntityID != "container-1" {
		t.Errorf("EntityID = %q, want %q", ve.EntityID, "container-1")
	}
}

func TestDiscoverJobs_InvalidCronSchedule(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled:  "true",
					schema.LabelJobName:     "bad-cron-job",
					schema.LabelJobSchedule: "invalid cron",
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("expected 1 validation error, got %d: %v", len(validationErrors), validationErrors)
	}

	ve := validationErrors[0]
	if ve.Field != schema.LabelJobSchedule {
		t.Errorf("Field = %q, want %q", ve.Field, schema.LabelJobSchedule)
	}

	// Job should still be created with default schedule
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}
	if foundJobs[0].Schedule != schema.DefaultJobSchedule() {
		t.Errorf("Schedule = %q, want default %q after invalid", foundJobs[0].Schedule, schema.DefaultJobSchedule())
	}
}

func TestDiscoverJobs_ConflictingSchedule(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-1",
				Name: "app1",
				Labels: map[string]string{
					schema.LabelJobEnabled:  "true",
					schema.LabelJobName:     "shared-job",
					schema.LabelJobSchedule: "0 2 * * *",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindContainer,
				ID:   "container-2",
				Name: "app2",
				Labels: map[string]string{
					schema.LabelJobEnabled:  "true",
					schema.LabelJobName:     "shared-job",
					schema.LabelJobSchedule: "0 3 * * *", // Different schedule!
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("expected 1 validation error for conflicting schedule, got %d: %v", len(validationErrors), validationErrors)
	}

	ve := validationErrors[0]
	if ve.Field != schema.LabelJobSchedule {
		t.Errorf("Field = %q, want %q", ve.Field, schema.LabelJobSchedule)
	}
	if !containsSubstring(ve.Message, "conflicting") {
		t.Errorf("Message should contain 'conflicting': %q", ve.Message)
	}

	// Job should still be created with first schedule
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}
	if foundJobs[0].Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want first value %q", foundJobs[0].Schedule, "0 2 * * *")
	}
}

func TestDiscoverJobs_VolumeAttachment(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "postgres-1",
				Name: "postgres",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "backup-job",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindVolume,
				ID:   "pgdata",
				Name: "pgdata",
				Labels: map[string]string{
					schema.LabelJobAttach:    "backup-job",
					schema.LabelJobMountPath: "/var/lib/postgresql/data",
					schema.LabelJobMountMode: "ro",
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}

	job := foundJobs[0]
	if len(job.AttachVolumes) != 1 {
		t.Fatalf("AttachVolumes length = %d, want 1", len(job.AttachVolumes))
	}

	vol := job.AttachVolumes[0]
	if vol.Name != "pgdata" {
		t.Errorf("Volume Name = %q, want %q", vol.Name, "pgdata")
	}
	if vol.MountPath != "/var/lib/postgresql/data" {
		t.Errorf("MountPath = %q, want %q", vol.MountPath, "/var/lib/postgresql/data")
	}
	if vol.Mode != "ro" {
		t.Errorf("Mode = %q, want %q", vol.Mode, "ro")
	}
}

func TestDiscoverJobs_VolumeDefaultMountPath(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "app-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "backup",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindVolume,
				ID:   "myvolume",
				Name: "myvolume",
				Labels: map[string]string{
					schema.LabelJobAttach: "backup",
					// No mount path - should default to /mnt/{volume_name}
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}

	vol := foundJobs[0].AttachVolumes[0]
	if vol.MountPath != "/mnt/myvolume" {
		t.Errorf("MountPath = %q, want default %q", vol.MountPath, "/mnt/myvolume")
	}
	if vol.Mode != schema.DefaultJobMountMode() {
		t.Errorf("Mode = %q, want default %q", vol.Mode, schema.DefaultJobMountMode())
	}
}

func TestDiscoverJobs_VolumeReferencesUnknownJob(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindVolume,
				ID:   "orphan-vol",
				Name: "orphan-vol",
				Labels: map[string]string{
					schema.LabelJobAttach: "nonexistent-job",
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(foundJobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(foundJobs))
	}
	if len(validationErrors) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(validationErrors))
	}

	ve := validationErrors[0]
	if ve.Field != schema.LabelJobAttach {
		t.Errorf("Field = %q, want %q", ve.Field, schema.LabelJobAttach)
	}
	if !containsSubstring(ve.Message, "unknown job") {
		t.Errorf("Message should contain 'unknown job': %q", ve.Message)
	}
}

func TestDiscoverJobs_InvalidMountMode(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "app-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "backup",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindVolume,
				ID:   "vol-1",
				Name: "vol-1",
				Labels: map[string]string{
					schema.LabelJobAttach:    "backup",
					schema.LabelJobMountMode: "invalid",
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(validationErrors))
	}

	ve := validationErrors[0]
	if ve.Field != schema.LabelJobMountMode {
		t.Errorf("Field = %q, want %q", ve.Field, schema.LabelJobMountMode)
	}

	// Volume should still be attached with default mode
	if len(foundJobs) != 1 || len(foundJobs[0].AttachVolumes) != 1 {
		t.Fatal("job should have 1 attached volume")
	}
	if foundJobs[0].AttachVolumes[0].Mode != schema.DefaultJobMountMode() {
		t.Errorf("Mode = %q, want default %q", foundJobs[0].AttachVolumes[0].Mode, schema.DefaultJobMountMode())
	}
}

func TestDiscoverJobs_StackResolution(t *testing.T) {
	tests := []struct {
		name          string
		labels        map[string]string
		meta          map[string]string
		expectedStack string
	}{
		{
			name: "bosun.stack takes precedence",
			labels: map[string]string{
				schema.LabelJobEnabled: "true",
				schema.LabelJobName:    "job",
				LabelStack:             "custom-stack",
			},
			meta: map[string]string{
				"compose.project": "compose-project",
			},
			expectedStack: "custom-stack",
		},
		{
			name: "falls back to compose.project",
			labels: map[string]string{
				schema.LabelJobEnabled: "true",
				schema.LabelJobName:    "job",
			},
			meta: map[string]string{
				"compose.project": "my-compose",
			},
			expectedStack: "my-compose",
		},
		{
			name: "no stack when not set",
			labels: map[string]string{
				schema.LabelJobEnabled: "true",
				schema.LabelJobName:    "job",
			},
			meta:          map[string]string{},
			expectedStack: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiscoverer()
			snapshot := labels.Snapshot{
				Entities: []labels.LabeledEntity{
					{
						Kind:   labels.KindContainer,
						ID:     "container-1",
						Name:   "app",
						Labels: tt.labels,
						Meta:   tt.meta,
					},
				},
				TakenAt: time.Now(),
			}

			foundJobs, _, err := d.DiscoverJobs(context.Background(), snapshot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(foundJobs) != 1 {
				t.Fatalf("expected 1 job, got %d", len(foundJobs))
			}

			if tt.expectedStack == "" {
				if len(foundJobs[0].TargetStacks) != 0 {
					t.Errorf("TargetStacks = %v, want empty", foundJobs[0].TargetStacks)
				}
			} else {
				if len(foundJobs[0].TargetStacks) != 1 || foundJobs[0].TargetStacks[0] != tt.expectedStack {
					t.Errorf("TargetStacks = %v, want [%s]", foundJobs[0].TargetStacks, tt.expectedStack)
				}
			}
		})
	}
}

func TestDiscoverJobs_IgnoresDisabledContainers(t *testing.T) {
	d := NewDiscoverer()
	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "enabled-1",
				Name: "enabled",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "enabled-job",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindContainer,
				ID:   "disabled-1",
				Name: "disabled",
				Labels: map[string]string{
					schema.LabelJobEnabled: "false",
					schema.LabelJobName:    "disabled-job",
				},
				Meta: map[string]string{},
			},
			{
				Kind: labels.KindContainer,
				ID:   "no-label-1",
				Name: "no-label",
				Labels: map[string]string{
					schema.LabelJobName: "no-enabled-job",
					// Missing bosun.job.enabled
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	foundJobs, validationErrors, err := d.DiscoverJobs(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
	if len(foundJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(foundJobs))
	}
	if foundJobs[0].Name != "enabled-job" {
		t.Errorf("Job name = %q, want %q", foundJobs[0].Name, "enabled-job")
	}
}

func TestDiscoverJobs_ContextCancellation(t *testing.T) {
	d := NewDiscoverer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	snapshot := labels.Snapshot{
		Entities: []labels.LabeledEntity{
			{
				Kind: labels.KindContainer,
				ID:   "container-1",
				Name: "app",
				Labels: map[string]string{
					schema.LabelJobEnabled: "true",
					schema.LabelJobName:    "job",
				},
				Meta: map[string]string{},
			},
		},
		TakenAt: time.Now(),
	}

	_, _, err := d.DiscoverJobs(ctx, snapshot)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestIsJobEnabled(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"False", false},
		{"0", false},
		{"", false},
		{"yes", false}, // Not supported
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			lbls := map[string]string{schema.LabelJobEnabled: tt.value}
			result := isJobEnabled(lbls)
			if result != tt.expected {
				t.Errorf("isJobEnabled(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// containsSubstring is a helper to check substring presence.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
