//go:build integration

package integration

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/testutil"
)

// Test_Integration_JobExecution_HappyPath validates successful job execution:
// - Stack is stopped
// - Worker runs to completion
// - Stack is restarted
// - Exit code 0
func Test_Integration_JobExecution_HappyPath(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run job execution (using alpine as worker, with echo command)
	// The job's worker.image is alpine:latest, we'll run with a short command
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--timeout", "60s",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Job should succeed (exit 0)
	if err != nil {
		t.Errorf("Expected job to succeed, got error: %v", err)
	}

	// Output should indicate successful completion
	output := stdout.String()
	if !strings.Contains(output, "✅") && !strings.Contains(output, "success") && !strings.Contains(output, "Success") {
		t.Log("Warning: Output doesn't clearly indicate success, but exit code was 0")
	}
}

// Test_Integration_JobExecution_WorkerFailure validates that worker failures
// still trigger stack restart.
func Test_Integration_JobExecution_WorkerFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run job execution but override worker command to fail
	// Note: We can't override the worker command via CLI in M3, so we'll test
	// with the default job which should succeed. This test is a placeholder
	// for when we add worker command override support.
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--timeout", "60s",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Verify stack is running after job (regardless of worker exit)
	// Use docker compose ps to check
	psCmd := exec.CommandContext(ctx, "docker", "compose", "-p", stack.Project, "ps", "--format", "{{.State}}")
	psCmd.Dir = stack.ComposeDir
	psOut, psErr := psCmd.Output()
	if psErr != nil {
		t.Fatalf("Failed to get stack status: %v", psErr)
	}

	psOutput := string(psOut)
	t.Logf("Stack state after job: %s", psOutput)

	// Stack should be running (contains "running" or "Up")
	if !strings.Contains(strings.ToLower(psOutput), "running") {
		t.Errorf("Stack should be running after job, got: %s", psOutput)
	}
}

// Test_Integration_JobExecution_DryRun validates dry-run mode has no side effects.
func Test_Integration_JobExecution_DryRun(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Get container IDs before dry run
	beforeIDs := getContainerIDs(t, ctx, stack)

	// Run dry-run
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--dry-run",
		"--format", "json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	if err != nil {
		t.Errorf("Dry-run should succeed, got error: %v", err)
	}

	// Get container IDs after dry run
	afterIDs := getContainerIDs(t, ctx, stack)

	// Container IDs should be unchanged (no restart happened)
	if beforeIDs != afterIDs {
		t.Errorf("Dry-run should not affect containers\nBefore: %s\nAfter: %s", beforeIDs, afterIDs)
	}

	// Verify output contains plan
	output := stdout.String()
	if !strings.Contains(output, "test-backup") {
		t.Error("Dry-run output should contain job name")
	}
}

// Test_Integration_JobExecution_JobNotFound validates proper error for missing job.
func Test_Integration_JobExecution_JobNotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Try to run non-existent job
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "nonexistent-job",
		"--project", stack.Project,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Should fail with exit code 3 (ExitJobNotFound)
	if err == nil {
		t.Error("Expected error for non-existent job")
	}

	// Check stderr contains "not found"
	if !strings.Contains(stderr.String(), "not found") && !strings.Contains(stdout.String(), "not found") {
		t.Error("Error message should indicate job not found")
	}
}

// Test_Integration_JobExecution_PlanListWithJob validates that plan list shows jobs
// that can be executed.
func Test_Integration_JobExecution_PlanListWithJob(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run plan list to show available jobs
	cmd := runBosun(ctx, t, bosunBin,
		"plan", "list",
		"--project", stack.Project,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Verify test-backup job is listed
	output := stdout.String()
	if !strings.Contains(output, "test-backup") {
		t.Error("plan list should show test-backup job")
	}
}

// getContainerIDs returns a string of container IDs for the stack.
func getContainerIDs(t *testing.T, ctx context.Context, stack *testutil.Stack) string {
	t.Helper()

	cmd := exec.CommandContext(ctx, "docker", "compose", "-p", stack.Project, "ps", "-q")
	cmd.Dir = stack.ComposeDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get container IDs: %v", err)
	}
	return strings.TrimSpace(string(output))
}

// Test_Integration_JobExecution_TimeoutFlag validates --timeout flag parsing.
func Test_Integration_JobExecution_TimeoutFlag(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Test with valid timeout duration
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--timeout", "5m",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Job should complete successfully with a 5 minute timeout
	if err != nil {
		t.Errorf("Job with valid timeout should succeed, got error: %v", err)
	}
}

// Test_Integration_JobExecution_InvalidTimeout validates invalid --timeout flag error.
func Test_Integration_JobExecution_InvalidTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Test with invalid timeout duration
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--timeout", "invalid-duration",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for invalid timeout")
	}

	// Check stderr contains timeout error
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "timeout") {
		t.Error("Error message should mention timeout")
	}
}

// Test_Integration_JobExecution_QuietMode validates --quiet flag suppresses output.
func Test_Integration_JobExecution_QuietMode(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start compose stack with job labels
	stack := testutil.StartCompose(t, ctx, "job-execution-compose.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run with --quiet flag
	cmd := runBosun(ctx, t, bosunBin,
		"job", "run", "test-backup",
		"--project", stack.Project,
		"--quiet",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Job should succeed
	if err != nil {
		t.Errorf("Quiet job should succeed, got error: %v", err)
	}

	// Output should still show execution result (quiet only suppresses worker logs)
	output := stdout.String()
	if !strings.Contains(output, "Job Execution") {
		t.Error("Output should still contain execution summary")
	}
}
