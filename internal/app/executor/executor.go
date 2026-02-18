package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// Executor orchestrates job execution.
type Executor struct {
	planner ports.JobPlanner
	compose ports.ComposeController
	worker  ports.WorkerRunner
	docker  *client.Client
	events  ports.EventEmitter
}

// New creates a new Executor with dependency injection.
func New(
	planner ports.JobPlanner,
	compose ports.ComposeController,
	worker ports.WorkerRunner,
	docker *client.Client,
	events ports.EventEmitter,
) *Executor {
	if events == nil {
		events = noopEventEmitter{}
	}
	return &Executor{
		planner: planner,
		compose: compose,
		worker:  worker,
		docker:  docker,
		events:  events,
	}
}

// Execute runs a job (implements ports.JobExecutor).
// Implements the step interpreter pattern: the executor iterates over plan.Steps
// and executes each step based on its type. This ensures plan-is-source-of-truth.
func (e *Executor) Execute(ctx context.Context, job jobs.Job, opts ports.ExecuteOptions) (ports.ExecutionResult, error) {
	// Initialize result
	runID := uuid.New().String()
	run := jobs.JobRun{
		ID:             runID,
		JobName:        job.Name,
		StackName:      job.TargetStacks[0], // M3: Single stack only
		Status:         jobs.RunStatusPending,
		StartedAt:      time.Now(),
		WorkerExitCode: -1,
	}

	result := ports.ExecutionResult{
		Run:         run,
		StepResults: []ports.StepResult{},
	}

	// Step 1: Generate execution plan
	plan, err := e.planner.Plan(ctx, job)
	if err != nil {
		run.Status = jobs.RunStatusFailed
		run.Error = fmt.Sprintf("plan generation failed: %v", err)
		run.CompletedAt = time.Now()
		result.Run = run
		return result, err
	}
	result.Plan = plan

	// Step 2: Pre-validate worker image (fail fast)
	if err := e.validateImage(ctx, job.WorkerImage); err != nil {
		run.Status = jobs.RunStatusFailed
		run.Error = fmt.Sprintf("worker image validation failed: %v", err)
		run.CompletedAt = time.Now()
		result.Run = run
		return result, jobs.ErrImageNotFound
	}

	// Update status to running
	run.Status = jobs.RunStatusRunning
	result.Run = run

	// Create execution context for step interpreter
	execCtx := &stepExecutionContext{
		runID:          runID,
		job:            job,
		opts:           opts,
		stackName:      run.StackName,
		stackStopped:   false,
		workerExitCode: -1,
	}

	// Step 3: Execute plan steps using step interpreter pattern
	// Plan-is-source-of-truth: the planner defines the sequence, executor interprets it
	var lastErr error
	for i, step := range plan.Steps {
		stepResult, err := e.executeStep(ctx, step, execCtx)
		result.StepResults = append(result.StepResults, stepResult)

		// Capture worker logs if this was a worker step
		if step.Type == jobs.StepTypeRunWorker && execCtx.workerLogs != "" {
			result.WorkerLogs = execCtx.workerLogs
		}

		if err != nil {
			lastErr = err
			// On failure, execute remaining start steps for cleanup
			// This ensures stack restart even on worker failure
			for j := i + 1; j < len(plan.Steps); j++ {
				remainingStep := plan.Steps[j]
				if remainingStep.Type == jobs.StepTypeStartContainers && execCtx.stackStopped && !opts.KeepStopped {
					cleanupResult, _ := e.executeStep(ctx, remainingStep, execCtx)
					result.StepResults = append(result.StepResults, cleanupResult)
				}
			}
			break
		}
	}

	// Update worker exit code from context
	run.WorkerExitCode = execCtx.workerExitCode

	// Determine final status
	if lastErr != nil {
		run.Status = jobs.RunStatusFailed
		run.Error = lastErr.Error()
	} else if execCtx.workerExitCode != 0 {
		run.Status = jobs.RunStatusFailed
		run.Error = fmt.Sprintf("worker failed with exit code %d", execCtx.workerExitCode)
	} else {
		run.Status = jobs.RunStatusSuccess
	}
	run.CompletedAt = time.Now()
	result.Run = run

	return result, lastErr
}

// stepExecutionContext holds state shared across step executions.
type stepExecutionContext struct {
	runID          string
	job            jobs.Job
	opts           ports.ExecuteOptions
	stackName      string
	stackStopped   bool
	workerExitCode int
	workerLogs     string
}

// executeStep executes a single plan step and returns the result.
// This is the core of the step interpreter pattern.
func (e *Executor) executeStep(ctx context.Context, step jobs.PlanStep, execCtx *stepExecutionContext) (ports.StepResult, error) {
	stepStart := time.Now()

	stepResult := ports.StepResult{
		Step:      step,
		StartedAt: stepStart,
	}

	var err error

	switch step.Type {
	case jobs.StepTypeStopContainers:
		err = e.executeStopStep(ctx, step, execCtx)

	case jobs.StepTypeRunWorker:
		err = e.executeWorkerStep(ctx, step, execCtx, &stepResult)

	case jobs.StepTypeStartContainers:
		err = e.executeStartStep(ctx, step, execCtx)

	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	stepResult.Duration = time.Since(stepStart)

	if err != nil {
		stepResult.Status = jobs.RunStatusFailed
		stepResult.Error = err.Error()
		return stepResult, err
	}

	stepResult.Status = jobs.RunStatusSuccess
	return stepResult, nil
}

// executeStopStep handles StepTypeStopContainers.
func (e *Executor) executeStopStep(ctx context.Context, step jobs.PlanStep, execCtx *stepExecutionContext) error {
	stopOpts := ports.DefaultStopOptions()
	if execCtx.opts.StopTimeoutOverride > 0 {
		stopOpts.Timeout = execCtx.opts.StopTimeoutOverride
	}

	var err error
	if step.UseCompose && step.ComposeProject != "" {
		err = e.compose.StopStack(ctx, step.ComposeProject, stopOpts)
	} else {
		err = e.compose.StopStack(ctx, execCtx.stackName, stopOpts)
	}

	if err != nil {
		return fmt.Errorf("failed to stop stack: %w", err)
	}

	execCtx.stackStopped = true
	return nil
}

// executeWorkerStep handles StepTypeRunWorker.
// Note: step parameter reserved for future per-step config (timeout, retry policy).
func (e *Executor) executeWorkerStep(ctx context.Context, _ jobs.PlanStep, execCtx *stepExecutionContext, stepResult *ports.StepResult) error {
	workerConfig := e.buildWorkerConfig(execCtx.job, execCtx.runID, execCtx.opts)
	workerResult, workerErr := e.worker.Run(ctx, workerConfig)

	stepResult.Details = workerResult.ContainerID
	execCtx.workerExitCode = workerResult.ExitCode
	execCtx.workerLogs = workerResult.Logs

	if workerErr != nil {
		return fmt.Errorf("worker execution error: %w", workerErr)
	}

	if workerResult.ExitCode != 0 {
		stepResult.Status = jobs.RunStatusFailed
		stepResult.Error = fmt.Sprintf("worker exited with code %d", workerResult.ExitCode)
		// Note: non-zero exit is not an error from executor perspective
		// The step completes, but with failed status
	}

	return nil
}

// executeStartStep handles StepTypeStartContainers.
// Note: ctx parameter unused because cleanup uses background context to ensure completion.
func (e *Executor) executeStartStep(_ context.Context, step jobs.PlanStep, execCtx *stepExecutionContext) error {
	// Skip start if stack was never stopped or if user requested to keep stopped
	if !execCtx.stackStopped || execCtx.opts.KeepStopped {
		return nil
	}

	startOpts := ports.DefaultStartOptions()
	if execCtx.opts.StartTimeoutOverride > 0 {
		startOpts.Timeout = execCtx.opts.StartTimeoutOverride
	}

	// Use background context for cleanup (not cancelled ctx)
	cleanupCtx, cancel := context.WithTimeout(context.Background(), startOpts.Timeout*2)
	defer cancel()

	var err error
	if step.UseCompose && step.ComposeProject != "" {
		err = e.compose.StartStack(cleanupCtx, step.ComposeProject, startOpts)
	} else {
		err = e.compose.StartStack(cleanupCtx, execCtx.stackName, startOpts)
	}

	if err != nil {
		return fmt.Errorf("failed to restart stack: %w", err)
	}

	return nil
}

// DryRun returns the execution plan without executing (implements ports.JobExecutor).
func (e *Executor) DryRun(ctx context.Context, job jobs.Job) (jobs.ExecutionPlan, error) {
	plan, err := e.planner.Plan(ctx, job)
	if err != nil {
		return jobs.ExecutionPlan{}, err
	}

	return plan, nil
}

// validateImage checks if the worker image is available.
func (e *Executor) validateImage(ctx context.Context, image string) error {
	// Try to inspect image locally
	_, err := e.docker.ImageInspect(ctx, image)
	if err == nil {
		return nil // Image exists locally
	}

	// Image not found locally - could try to pull, but for M3 we just fail
	return fmt.Errorf("image %s not found locally (pull not implemented in M3): %w", image, jobs.ErrImageNotFound)
}

// buildWorkerConfig constructs worker configuration from job and options.
func (e *Executor) buildWorkerConfig(job jobs.Job, runID string, opts ports.ExecuteOptions) ports.WorkerConfig {
	config := ports.DefaultWorkerConfig()
	config.Image = job.WorkerImage
	config.RunID = runID
	config.JobName = job.Name
	config.StackName = job.TargetStacks[0]
	config.DryRun = false
	config.KeepOnFailure = opts.KeepFailedWorker

	// Apply timeout override
	if opts.TimeoutOverride > 0 {
		config.Timeout = opts.TimeoutOverride
	}

	// Wire log streaming (Phase 5: US3)
	if !opts.Quiet && opts.LogWriter != nil {
		config.LogWriter = opts.LogWriter
	}

	// Convert volume attachments to mounts
	config.Mounts = make([]ports.VolumeMount, len(job.AttachVolumes))
	for i, vol := range job.AttachVolumes {
		config.Mounts[i] = ports.VolumeMount{
			Source:   vol.Name, // Fixed: use Name field
			Target:   vol.MountPath,
			ReadOnly: vol.Mode == "ro", // Fixed: use Mode field
		}
	}

	// TODO: Add worker env vars from labels (bosun.job.worker.env.*)
	// For M3 MVP, we skip this - can be added in Phase 5-7

	return config
}
