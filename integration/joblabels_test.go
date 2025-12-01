//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/adapters/joblabels"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"github.com/simone-viozzi/bosun/internal/testutil"
)

// Test_Integration_JobLabels_Discovery validates that jobs can be discovered
// from containers and volumes with bosun.job.* labels.
func Test_Integration_JobLabels_Discovery(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Create a DockerLabelSource to get the snapshot
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		t.Fatalf("failed to create DockerLabelSource: %v", err)
	}

	// Take a snapshot with bosun. prefix filter and project filter for isolation
	sel := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: false,
		ProjectFilter:  []string{stack.Project},
	}

	snapshot, err := source.Snapshot(ctx, sel)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Create a job discoverer and discover jobs
	discoverer := joblabels.NewDiscoverer()
	jobs, validationErrors, err := discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		t.Fatalf("DiscoverJobs failed: %v", err)
	}

	t.Logf("Discovered %d jobs with %d validation errors from project %s",
		len(jobs), len(validationErrors), stack.Project)

	// We expect exactly 1 job: "daily-backup" (merged from postgres and redis containers)
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
		for _, job := range jobs {
			t.Logf("  Found job: %s", job.Name)
		}
	}

	// Verify the job properties
	if len(jobs) > 0 {
		job := jobs[0]

		if job.Name != "daily-backup" {
			t.Errorf("Job name = %q, want %q", job.Name, "daily-backup")
		}

		if job.Schedule != "0 2 * * *" {
			t.Errorf("Schedule = %q, want %q", job.Schedule, "0 2 * * *")
		}

		if job.WorkerImage != "backup-worker:test" {
			t.Errorf("WorkerImage = %q, want %q", job.WorkerImage, "backup-worker:test")
		}

		// Should have exactly 2 target containers (postgres and redis) with project filter
		if len(job.TargetContainers) != 2 {
			t.Errorf("TargetContainers length = %d, want 2", len(job.TargetContainers))
		}

		// Should have exactly 2 attached volumes (pgdata and redis-data) with project filter
		if len(job.AttachVolumes) != 2 {
			t.Errorf("AttachVolumes length = %d, want 2", len(job.AttachVolumes))
		}

		// Verify expected volume mount paths are present
		pathsFound := make(map[string]bool)
		for _, vol := range job.AttachVolumes {
			pathsFound[vol.MountPath] = true
			if vol.Mode != "ro" {
				t.Errorf("Volume %s Mode = %q, want %q", vol.Name, vol.Mode, "ro")
			}
		}
		if !pathsFound["/backup/postgres"] {
			t.Error("Missing volume mount at /backup/postgres")
		}
		if !pathsFound["/backup/redis"] {
			t.Error("Missing volume mount at /backup/redis")
		}

		t.Logf("Job %q has %d containers, %d volumes, schedule=%q",
			job.Name, len(job.TargetContainers), len(job.AttachVolumes), job.Schedule)
	}

	// Log any validation errors found
	// Note: Docker Compose doesn't create volumes that aren't used by any service,
	// so the orphan-volume in the compose file won't actually be created.
	// Orphan detection is tested in unit tests instead.
	for _, ve := range validationErrors {
		t.Logf("Validation error: %s.%s - %s", ve.EntityKind, ve.Field, ve.Message)
	}

	t.Logf("Integration test completed successfully with project: %s", stack.Project)
}
