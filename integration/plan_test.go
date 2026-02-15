//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
	"github.com/simone-viozzi/bosun/internal/testutil"
)

// planListOutput matches the JSON/YAML output structure of `bosun plan list`.
type planListOutput struct {
	Jobs             []jobs.Job              `json:"jobs" yaml:"jobs"`
	ValidationErrors []ports.ValidationError `json:"validationErrors,omitempty" yaml:"validationErrors,omitempty"`
}

// Test_Integration_PlanList_TextFormat validates the text output of `bosun plan list`.
func Test_Integration_PlanList_TextFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan list --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "list", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// We expect exit code 1 due to validation errors (orphan volume)
	if err == nil {
		t.Log("Command succeeded (no validation errors)")
	}

	output := stdout.String()
	t.Logf("Text output:\n%s", output)

	// Verify table headers are present
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "SCHEDULE") {
		t.Error("Expected table headers (NAME, SCHEDULE) in text output")
	}

	// Verify our job is listed
	if !strings.Contains(output, "daily-backup") {
		t.Error("Expected 'daily-backup' job in output")
	}

	// Verify schedule is shown
	if !strings.Contains(output, "0 2 * * *") {
		t.Error("Expected schedule '0 2 * * *' in output")
	}
}

// Test_Integration_PlanList_JSONFormat validates JSON output of `bosun plan list`.
func Test_Integration_PlanList_JSONFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan list --format json --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "list", "--format", "json", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run() // Ignore exit code (may be 1 due to validation errors)

	output := stdout.Bytes()
	t.Logf("JSON output:\n%s", string(output))

	// Parse JSON
	var result planListOutput
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify we have exactly 1 job with project filter
	if len(result.Jobs) != 1 {
		t.Fatalf("Expected exactly 1 job in JSON output, got %d", len(result.Jobs))
	}

	// Verify it's our job
	foundJob := &result.Jobs[0]
	if foundJob.Name != "daily-backup" {
		t.Fatalf("Expected 'daily-backup' job, got %q", foundJob.Name)
	}

	// Verify job properties
	if foundJob.Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want %q", foundJob.Schedule, "0 2 * * *")
	}
	if foundJob.WorkerImage != "backup-worker:test" {
		t.Errorf("WorkerImage = %q, want %q", foundJob.WorkerImage, "backup-worker:test")
	}
	if len(foundJob.TargetContainers) != 2 {
		t.Errorf("TargetContainers = %d, want 2", len(foundJob.TargetContainers))
	}
	if len(foundJob.AttachVolumes) != 2 {
		t.Errorf("AttachVolumes = %d, want 2", len(foundJob.AttachVolumes))
	}

	// Verify validation errors are present (for orphan volume)
	if len(result.ValidationErrors) == 0 {
		t.Log("No validation errors (orphan volume may not have been created)")
	}
}

// Test_Integration_PlanList_YAMLFormat validates YAML output of `bosun plan list`.
func Test_Integration_PlanList_YAMLFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan list --format yaml --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "list", "--format", "yaml", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run() // Ignore exit code (may be 1 due to validation errors)

	output := stdout.Bytes()
	t.Logf("YAML output:\n%s", string(output))

	// Parse YAML
	var result planListOutput
	if err := yaml.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	// Verify we have exactly 1 job with project filter
	if len(result.Jobs) != 1 {
		t.Fatalf("Expected exactly 1 job in YAML output, got %d", len(result.Jobs))
	}

	// Verify it's our job
	if result.Jobs[0].Name != "daily-backup" {
		t.Errorf("Expected 'daily-backup' job in YAML output, got %q", result.Jobs[0].Name)
	}
}

// Test_Integration_PlanList_NoJobs validates the "no jobs" message when filtering
// to a project without any job labels.
func Test_Integration_PlanList_NoJobs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack WITHOUT job labels
	stack := testutil.StartCompose(t, ctx, "dockerlabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan list --project <project>` - this project has no job labels
	cmd := runBosun(ctx, t, bosunBin, "plan", "list", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command error: %v", err)
	}

	stderrStr := stderr.String()
	t.Logf("stderr:\n%s", stderrStr)
	t.Logf("stdout:\n%s", stdout.String())

	// Should show "No jobs found" message on stderr
	if !strings.Contains(stderrStr, "No jobs found") {
		t.Error("Expected 'No jobs found' message on stderr")
	}
}

// Test_Integration_PlanShow_TextFormat validates text output of `bosun plan show <job>`.
func Test_Integration_PlanShow_TextFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan show daily-backup --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "show", "daily-backup", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command error: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	t.Logf("Text output:\n%s", output)

	// Verify job info headers
	if !strings.Contains(output, "daily-backup") {
		t.Error("Expected job name 'daily-backup' in text output")
	}
	if !strings.Contains(output, "0 2 * * *") {
		t.Error("Expected schedule '0 2 * * *' in text output")
	}

	// Verify execution steps are shown
	if !strings.Contains(output, "Steps (") {
		t.Error("Expected 'Steps' section in output")
	}

	// Verify stop steps come before run-worker
	stopIdx := strings.Index(output, "stop_containers")
	runIdx := strings.Index(output, "run_worker")
	if stopIdx == -1 || runIdx == -1 {
		t.Error("Expected both 'stop_containers' and 'run_worker' steps in output")
	} else if stopIdx > runIdx {
		t.Error("Expected stop steps to come before run-worker step")
	}
}

// Test_Integration_PlanShow_JSONFormat validates JSON output of `bosun plan show <job>`.
func Test_Integration_PlanShow_JSONFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan show daily-backup --format json --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "show", "daily-backup", "--format", "json", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command error: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.Bytes()
	t.Logf("JSON output:\n%s", string(output))

	// Parse JSON - the output is an ExecutionPlan directly
	var result jobs.ExecutionPlan
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify job name
	if result.JobName != "daily-backup" {
		t.Errorf("JobName = %q, want %q", result.JobName, "daily-backup")
	}

	// Verify plan has steps
	if len(result.Steps) == 0 {
		t.Fatal("Expected steps in execution plan")
	}

	// Verify step ordering: stop steps before run-worker
	seenRunWorker := false
	for _, step := range result.Steps {
		if step.Type == jobs.StepTypeRunWorker {
			seenRunWorker = true
		}
		if step.Type == jobs.StepTypeStopContainers && seenRunWorker {
			t.Error("Stop step found after run-worker step - ordering violation")
		}
	}

	// Verify run-worker step exists
	if !seenRunWorker {
		t.Error("Expected at least one run-worker step")
	}

	// Verify plan has volumes attached
	hasVolumes := false
	for _, step := range result.Steps {
		if step.Type == jobs.StepTypeRunWorker && len(step.VolumeMounts) > 0 {
			hasVolumes = true
			break
		}
	}
	if !hasVolumes {
		t.Log("Warning: No volumes attached to run-worker step")
	}
}

// Test_Integration_PlanShow_JobNotFound validates error handling for unknown jobs.
func Test_Integration_PlanShow_JobNotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan show nonexistent-job --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "show", "nonexistent-job", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected non-zero exit code for nonexistent job")
	}

	stderrOutput := stderr.String()
	t.Logf("Stderr output:\n%s", stderrOutput)

	// Verify error message mentions the job wasn't found
	if !strings.Contains(stderrOutput, "nonexistent-job") {
		t.Error("Expected error to mention the job name 'nonexistent-job'")
	}

	// Verify available jobs are shown
	if !strings.Contains(stderrOutput, "daily-backup") {
		t.Error("Expected available jobs list to include 'daily-backup'")
	}
}

// Test_Integration_PlanShow_YAMLFormat validates YAML output of `bosun plan show <job>`.
func Test_Integration_PlanShow_YAMLFormat(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "joblabels-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run `bosun plan show daily-backup --format yaml --project <project>` for isolation
	cmd := runBosun(ctx, t, bosunBin, "plan", "show", "daily-backup", "--format", "yaml", "--project", stack.Project)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command error: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.Bytes()
	t.Logf("YAML output:\n%s", string(output))

	// Parse YAML - the output is an ExecutionPlan directly
	var result jobs.ExecutionPlan
	if err := yaml.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	// Verify job name
	if result.JobName != "daily-backup" {
		t.Errorf("JobName = %q, want %q", result.JobName, "daily-backup")
	}

	// Verify plan has steps
	if len(result.Steps) == 0 {
		t.Fatal("Expected steps in execution plan")
	}
}
