package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// mockPlanner is a mock implementation of ports.JobPlanner
type mockPlanner struct {
	plan     jobs.ExecutionPlan
	planErr  error
	planCall int
}

func (m *mockPlanner) Plan(ctx context.Context, job jobs.Job) (jobs.ExecutionPlan, error) {
	m.planCall++
	if m.planErr != nil {
		return jobs.ExecutionPlan{}, m.planErr
	}
	return m.plan, nil
}

// mockComposeController is a mock implementation of ports.ComposeController
type mockComposeController struct {
	containers      []ports.StackContainer
	listErr         error
	stopErr         error
	startErr        error
	isRunning       bool
	isRunningErr    error
	stopCallCount   int
	startCallCount  int
	stoppedProjects []string
	startedProjects []string
}

func (m *mockComposeController) ListStackContainers(ctx context.Context, projectName string) ([]ports.StackContainer, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.containers, nil
}

func (m *mockComposeController) IsStackRunning(ctx context.Context, projectName string) (bool, error) {
	if m.isRunningErr != nil {
		return false, m.isRunningErr
	}
	return m.isRunning, nil
}

func (m *mockComposeController) StopStack(ctx context.Context, projectName string, opts ports.StopOptions) error {
	m.stopCallCount++
	m.stoppedProjects = append(m.stoppedProjects, projectName)
	if m.stopErr != nil {
		return m.stopErr
	}
	return nil
}

func (m *mockComposeController) StartStack(ctx context.Context, projectName string, opts ports.StartOptions) error {
	m.startCallCount++
	m.startedProjects = append(m.startedProjects, projectName)
	if m.startErr != nil {
		return m.startErr
	}
	return nil
}

// mockWorkerRunner is a mock implementation of ports.WorkerRunner
type mockWorkerRunner struct {
	result       ports.WorkerResult
	runErr       error
	runCallCount int
	configs      []ports.WorkerConfig
}

func (m *mockWorkerRunner) Run(ctx context.Context, config ports.WorkerConfig) (ports.WorkerResult, error) {
	m.runCallCount++
	m.configs = append(m.configs, config)
	if m.runErr != nil {
		return ports.WorkerResult{}, m.runErr
	}
	return m.result, nil
}

func TestNew(t *testing.T) {
	exec := New(&mockPlanner{}, &mockComposeController{}, &mockWorkerRunner{}, nil, nil)
	if exec == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDryRun_Success(t *testing.T) {
	expectedPlan := jobs.ExecutionPlan{
		JobName: "test-job",
		Steps: []jobs.PlanStep{
			{Type: jobs.StepTypeStopContainers, Description: "Stop containers"},
			{Type: jobs.StepTypeRunWorker, Description: "Run worker"},
		},
	}

	planner := &mockPlanner{plan: expectedPlan}
	compose := &mockComposeController{}
	worker := &mockWorkerRunner{}

	exec := New(planner, compose, worker, nil, nil)

	job := jobs.Job{
		Name:         "test-job",
		TargetStacks: []string{"mystack"},
		WorkerImage:  "worker:v1",
	}

	plan, err := exec.DryRun(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.JobName != expectedPlan.JobName {
		t.Errorf("JobName = %q, want %q", plan.JobName, expectedPlan.JobName)
	}

	if len(plan.Steps) != len(expectedPlan.Steps) {
		t.Errorf("Steps length = %d, want %d", len(plan.Steps), len(expectedPlan.Steps))
	}

	// Verify no side effects - compose should not be called
	if compose.stopCallCount > 0 {
		t.Errorf("StopStack called %d times during dry run, want 0", compose.stopCallCount)
	}
	if compose.startCallCount > 0 {
		t.Errorf("StartStack called %d times during dry run, want 0", compose.startCallCount)
	}
	if worker.runCallCount > 0 {
		t.Errorf("Worker.Run called %d times during dry run, want 0", worker.runCallCount)
	}
}

func TestDryRun_PlannerError(t *testing.T) {
	plannerErr := errors.New("planner failed")
	planner := &mockPlanner{planErr: plannerErr}

	exec := New(planner, &mockComposeController{}, &mockWorkerRunner{}, nil, nil)

	job := jobs.Job{Name: "test-job", TargetStacks: []string{"mystack"}}

	_, err := exec.DryRun(context.Background(), job)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExecute_PlannerError(t *testing.T) {
	plannerErr := errors.New("planner failed")
	planner := &mockPlanner{planErr: plannerErr}
	compose := &mockComposeController{}

	exec := New(planner, compose, &mockWorkerRunner{}, nil, nil)

	job := jobs.Job{
		Name:         "test-job",
		TargetStacks: []string{"mystack"},
		WorkerImage:  "worker:v1",
	}

	result, err := exec.Execute(context.Background(), job, ports.DefaultExecuteOptions())
	if err == nil {
		t.Error("expected error, got nil")
	}

	if result.Run.Status != jobs.RunStatusFailed {
		t.Errorf("Status = %q, want %q", result.Run.Status, jobs.RunStatusFailed)
	}

	// Stack should not have been touched
	if compose.stopCallCount > 0 {
		t.Errorf("StopStack called %d times on planner error, want 0", compose.stopCallCount)
	}
}

func TestDefaultExecuteOptions(t *testing.T) {
	opts := ports.DefaultExecuteOptions()

	if opts.KeepStopped {
		t.Error("KeepStopped = true, want false")
	}
	if opts.KeepFailedWorker {
		t.Error("KeepFailedWorker = true, want false")
	}
	if opts.Quiet {
		t.Error("Quiet = true, want false")
	}
}

func TestExecutionResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   ports.ExecutionResult
		expected bool
	}{
		{
			name: "success - status success and exit 0",
			result: ports.ExecutionResult{
				Run: jobs.JobRun{
					Status:         jobs.RunStatusSuccess,
					WorkerExitCode: 0,
				},
			},
			expected: true,
		},
		{
			name: "failure - status failed",
			result: ports.ExecutionResult{
				Run: jobs.JobRun{
					Status:         jobs.RunStatusFailed,
					WorkerExitCode: 0,
				},
			},
			expected: false,
		},
		{
			name: "failure - non-zero exit code",
			result: ports.ExecutionResult{
				Run: jobs.JobRun{
					Status:         jobs.RunStatusSuccess,
					WorkerExitCode: 1,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepResult_StepSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   ports.StepResult
		expected bool
	}{
		{
			name: "success status",
			result: ports.StepResult{
				Status: jobs.RunStatusSuccess,
			},
			expected: true,
		},
		{
			name: "failed status",
			result: ports.StepResult{
				Status: jobs.RunStatusFailed,
			},
			expected: false,
		},
		{
			name: "pending status",
			result: ports.StepResult{
				Status: jobs.RunStatusPending,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.StepSuccess(); got != tt.expected {
				t.Errorf("StepSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRunStatusConstants(t *testing.T) {
	// Verify run status constants
	if jobs.RunStatusPending != "pending" {
		t.Errorf("RunStatusPending = %q, want %q", jobs.RunStatusPending, "pending")
	}
	if jobs.RunStatusRunning != "running" {
		t.Errorf("RunStatusRunning = %q, want %q", jobs.RunStatusRunning, "running")
	}
	if jobs.RunStatusSuccess != "success" {
		t.Errorf("RunStatusSuccess = %q, want %q", jobs.RunStatusSuccess, "success")
	}
	if jobs.RunStatusFailed != "failed" {
		t.Errorf("RunStatusFailed = %q, want %q", jobs.RunStatusFailed, "failed")
	}
	if jobs.RunStatusCancelled != "cancelled" {
		t.Errorf("RunStatusCancelled = %q, want %q", jobs.RunStatusCancelled, "cancelled")
	}
}

func TestStepTypeConstants(t *testing.T) {
	// Verify step type constants exist
	if jobs.StepTypeStopContainers == "" {
		t.Error("StepTypeStopContainers is empty")
	}
	if jobs.StepTypeStartContainers == "" {
		t.Error("StepTypeStartContainers is empty")
	}
	if jobs.StepTypeRunWorker == "" {
		t.Error("StepTypeRunWorker is empty")
	}
}

func TestJobRun_Fields(t *testing.T) {
	now := time.Now()
	run := jobs.JobRun{
		ID:             "run-123",
		JobName:        "daily-backup",
		StackName:      "mystack",
		Status:         jobs.RunStatusSuccess,
		StartedAt:      now,
		CompletedAt:    now.Add(5 * time.Minute),
		WorkerExitCode: 0,
		Error:          "",
	}

	if run.ID != "run-123" {
		t.Errorf("ID = %q, want %q", run.ID, "run-123")
	}
	if run.JobName != "daily-backup" {
		t.Errorf("JobName = %q, want %q", run.JobName, "daily-backup")
	}
	if run.StackName != "mystack" {
		t.Errorf("StackName = %q, want %q", run.StackName, "mystack")
	}
	if run.Status != jobs.RunStatusSuccess {
		t.Errorf("Status = %q, want %q", run.Status, jobs.RunStatusSuccess)
	}
	if run.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}
	if run.CompletedAt.IsZero() {
		t.Error("CompletedAt should not be zero")
	}
	if run.Error != "" {
		t.Errorf("Error = %q, want empty", run.Error)
	}
	if run.WorkerExitCode != 0 {
		t.Errorf("WorkerExitCode = %d, want %d", run.WorkerExitCode, 0)
	}
}

func TestExecuteOptions_Fields(t *testing.T) {
	opts := ports.ExecuteOptions{
		TimeoutOverride:      10 * time.Minute,
		StopTimeoutOverride:  1 * time.Minute,
		StartTimeoutOverride: 2 * time.Minute,
		KeepStopped:          true,
		KeepFailedWorker:     true,
		Quiet:                true,
	}

	if opts.TimeoutOverride != 10*time.Minute {
		t.Errorf("TimeoutOverride = %v, want %v", opts.TimeoutOverride, 10*time.Minute)
	}
	if opts.StopTimeoutOverride != 1*time.Minute {
		t.Errorf("StopTimeoutOverride = %v, want %v", opts.StopTimeoutOverride, 1*time.Minute)
	}
	if opts.StartTimeoutOverride != 2*time.Minute {
		t.Errorf("StartTimeoutOverride = %v, want %v", opts.StartTimeoutOverride, 2*time.Minute)
	}
	if !opts.KeepStopped {
		t.Error("KeepStopped = false, want true")
	}
	if !opts.KeepFailedWorker {
		t.Error("KeepFailedWorker = false, want true")
	}
	if !opts.Quiet {
		t.Error("Quiet = false, want true")
	}
}

func TestExecutionPlan_Fields(t *testing.T) {
	now := time.Now()
	plan := jobs.ExecutionPlan{
		JobName:   "daily-backup",
		CreatedAt: now,
		Steps: []jobs.PlanStep{
			{Type: jobs.StepTypeStopContainers, Description: "Stop stack"},
			{Type: jobs.StepTypeRunWorker, Description: "Run worker", WorkerImage: "backup:v1"},
			{Type: jobs.StepTypeStartContainers, Description: "Start stack"},
		},
	}

	if plan.JobName != "daily-backup" {
		t.Errorf("JobName = %q, want %q", plan.JobName, "daily-backup")
	}
	if plan.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if len(plan.Steps) != 3 {
		t.Errorf("Steps length = %d, want 3", len(plan.Steps))
	}
	if plan.Steps[1].WorkerImage != "backup:v1" {
		t.Errorf("Steps[1].WorkerImage = %q, want %q", plan.Steps[1].WorkerImage, "backup:v1")
	}
}

func TestPlanStep_Fields(t *testing.T) {
	step := jobs.PlanStep{
		Type:           jobs.StepTypeRunWorker,
		Description:    "Run backup worker",
		ContainerIDs:   []string{"container-1", "container-2"},
		ContainerNames: []string{"mystack-web-1", "mystack-db-1"},
		WorkerImage:    "backup:v1",
		VolumeMounts: []jobs.VolumeAttachment{
			{Name: "pgdata", MountPath: "/data/pg", Mode: "ro"},
		},
		UseCompose:     true,
		ComposeProject: "mystack",
	}

	if step.Type != jobs.StepTypeRunWorker {
		t.Errorf("Type = %q, want %q", step.Type, jobs.StepTypeRunWorker)
	}
	if len(step.ContainerNames) != 2 {
		t.Errorf("ContainerNames length = %d, want 2", len(step.ContainerNames))
	}
	if step.ContainerNames[0] != "mystack-web-1" {
		t.Errorf("ContainerNames[0] = %q, want %q", step.ContainerNames[0], "mystack-web-1")
	}
	if step.Description != "Run backup worker" {
		t.Errorf("Description = %q, want %q", step.Description, "Run backup worker")
	}
	if len(step.ContainerIDs) != 2 {
		t.Errorf("ContainerIDs length = %d, want 2", len(step.ContainerIDs))
	}
	if step.WorkerImage != "backup:v1" {
		t.Errorf("WorkerImage = %q, want %q", step.WorkerImage, "backup:v1")
	}
	if len(step.VolumeMounts) != 1 {
		t.Errorf("VolumeMounts length = %d, want 1", len(step.VolumeMounts))
	}
	if !step.UseCompose {
		t.Error("UseCompose = false, want true")
	}
	if step.ComposeProject != "mystack" {
		t.Errorf("ComposeProject = %q, want %q", step.ComposeProject, "mystack")
	}
}

func TestDryRun_NoSideEffects(t *testing.T) {
	// This is the T024 test: verify dry-run has no side effects
	expectedPlan := jobs.ExecutionPlan{
		JobName: "test-job",
		Steps: []jobs.PlanStep{
			{Type: jobs.StepTypeStopContainers, Description: "Stop stack 'mystack'"},
			{Type: jobs.StepTypeRunWorker, Description: "Run worker 'backup:v1'", WorkerImage: "backup:v1"},
			{Type: jobs.StepTypeStartContainers, Description: "Start stack 'mystack'"},
		},
	}

	planner := &mockPlanner{plan: expectedPlan}
	compose := &mockComposeController{
		containers: []ports.StackContainer{
			{ID: "c1", Name: "mystack-web-1", ServiceName: "web", State: "running"},
			{ID: "c2", Name: "mystack-db-1", ServiceName: "db", State: "running"},
		},
		isRunning: true,
	}
	worker := &mockWorkerRunner{
		result: ports.WorkerResult{
			ExitCode:    0,
			Logs:        "backup completed",
			ContainerID: "worker-123",
			Duration:    30 * time.Second,
		},
	}

	exec := New(planner, compose, worker, nil, nil)

	job := jobs.Job{
		Name:         "test-job",
		TargetStacks: []string{"mystack"},
		WorkerImage:  "backup:v1",
	}

	// Execute dry-run
	plan, err := exec.DryRun(context.Background(), job)
	if err != nil {
		t.Fatalf("DryRun error: %v", err)
	}

	// Verify plan was returned
	if plan.JobName != "test-job" {
		t.Errorf("plan.JobName = %q, want %q", plan.JobName, "test-job")
	}
	if len(plan.Steps) != 3 {
		t.Errorf("plan.Steps length = %d, want 3", len(plan.Steps))
	}

	// CRITICAL: Verify no side effects
	if planner.planCall != 1 {
		t.Errorf("Planner.Plan called %d times, want 1", planner.planCall)
	}

	// ComposeController should NOT have been called for stop/start
	if compose.stopCallCount != 0 {
		t.Errorf("ComposeController.StopStack called %d times during dry-run, want 0", compose.stopCallCount)
	}
	if compose.startCallCount != 0 {
		t.Errorf("ComposeController.StartStack called %d times during dry-run, want 0", compose.startCallCount)
	}

	// WorkerRunner should NOT have been called
	if worker.runCallCount != 0 {
		t.Errorf("WorkerRunner.Run called %d times during dry-run, want 0", worker.runCallCount)
	}

	// Verify lists of actions are empty
	if len(compose.stoppedProjects) != 0 {
		t.Errorf("stoppedProjects = %v, want empty", compose.stoppedProjects)
	}
	if len(compose.startedProjects) != 0 {
		t.Errorf("startedProjects = %v, want empty", compose.startedProjects)
	}
	if len(worker.configs) != 0 {
		t.Errorf("worker.configs = %v, want empty", worker.configs)
	}
}
