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
}

// New creates a new Executor with dependency injection.
func New(
	planner ports.JobPlanner,
	compose ports.ComposeController,
	worker ports.WorkerRunner,
	docker *client.Client,
) *Executor {
	return &Executor{
		planner: planner,
		compose: compose,
		worker:  worker,
		docker:  docker,
	}
}

// Execute runs a job (implements ports.JobExecutor).
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

	// TODO: DESIGN ISSUE - Plan is generated but NOT used to drive execution.
	// Decision #4 in wip_smell_milestone3 states "Plan-is-source-of-truth" but this
	// executor hardcodes stop→worker→start sequence instead of interpreting plan.Steps.
	// The plan is only used to populate result.Plan for display purposes.
	// Fix: Implement step interpreter loop: for _, step := range plan.Steps { executeStep(step) }
	// See: smell #23 in wip_smell_milestone3, issue #142 (marked FIXED but not actually fixed)

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

	// Track if stack was stopped (for cleanup)
	stackStopped := false

	// Defer: ALWAYS attempt to restart stack
	defer func() {
		if stackStopped && !opts.KeepStopped {
			startOpts := ports.DefaultStartOptions()
			if opts.StartTimeoutOverride > 0 {
				startOpts.Timeout = opts.StartTimeoutOverride
			}

			// Use background context for cleanup (not cancelled ctx)
			cleanupCtx, cancel := context.WithTimeout(context.Background(), startOpts.Timeout*2)
			defer cancel()

			startStepStart := time.Now()
			startErr := e.compose.StartStack(cleanupCtx, run.StackName, startOpts)

			stepResult := ports.StepResult{
				Step: jobs.PlanStep{
					Type:        jobs.StepTypeStartContainers,
					Description: fmt.Sprintf("Start stack '%s'", run.StackName),
				},
				StartedAt: startStepStart,
				Duration:  time.Since(startStepStart),
			}

			if startErr != nil {
				stepResult.Status = jobs.RunStatusFailed
				stepResult.Error = fmt.Sprintf("failed to restart stack: %v", startErr)
			} else {
				stepResult.Status = jobs.RunStatusSuccess
			}

			result.StepResults = append(result.StepResults, stepResult)
		}
	}()

	// Step 3: Stop stack
	stopOpts := ports.DefaultStopOptions()
	if opts.StopTimeoutOverride > 0 {
		stopOpts.Timeout = opts.StopTimeoutOverride
	}

	stopStepStart := time.Now()
	stopErr := e.compose.StopStack(ctx, run.StackName, stopOpts)

	stopStepResult := ports.StepResult{
		Step: jobs.PlanStep{
			Type:        jobs.StepTypeStopContainers,
			Description: fmt.Sprintf("Stop stack '%s'", run.StackName),
		},
		StartedAt: stopStepStart,
		Duration:  time.Since(stopStepStart),
	}

	if stopErr != nil {
		stopStepResult.Status = jobs.RunStatusFailed
		stopStepResult.Error = fmt.Sprintf("failed to stop stack: %v", stopErr)
		result.StepResults = append(result.StepResults, stopStepResult)

		run.Status = jobs.RunStatusFailed
		run.Error = stopErr.Error()
		run.CompletedAt = time.Now()
		result.Run = run
		return result, stopErr
	}

	stopStepResult.Status = jobs.RunStatusSuccess
	result.StepResults = append(result.StepResults, stopStepResult)
	stackStopped = true

	// Step 4: Run worker
	workerConfig := e.buildWorkerConfig(job, runID, opts)
	workerStepStart := time.Now()
	workerResult, workerErr := e.worker.Run(ctx, workerConfig)

	workerStepResult := ports.StepResult{
		Step: jobs.PlanStep{
			Type:        jobs.StepTypeRunWorker,
			Description: fmt.Sprintf("Run worker '%s'", job.WorkerImage),
		},
		StartedAt: workerStepStart,
		Duration:  time.Since(workerStepStart),
		Details:   workerResult.ContainerID,
	}

	run.WorkerExitCode = workerResult.ExitCode
	result.WorkerLogs = workerResult.Logs

	if workerErr != nil {
		workerStepResult.Status = jobs.RunStatusFailed
		workerStepResult.Error = fmt.Sprintf("worker execution error: %v", workerErr)
		result.StepResults = append(result.StepResults, workerStepResult)

		run.Status = jobs.RunStatusFailed
		run.Error = workerErr.Error()
		run.CompletedAt = time.Now()
		result.Run = run
		return result, workerErr
	}

	if workerResult.ExitCode != 0 {
		workerStepResult.Status = jobs.RunStatusFailed
		workerStepResult.Error = fmt.Sprintf("worker exited with code %d", workerResult.ExitCode)
	} else {
		workerStepResult.Status = jobs.RunStatusSuccess
	}
	result.StepResults = append(result.StepResults, workerStepResult)

	// Update final status
	if workerResult.ExitCode == 0 {
		run.Status = jobs.RunStatusSuccess
	} else {
		run.Status = jobs.RunStatusFailed
		run.Error = fmt.Sprintf("worker failed with exit code %d", workerResult.ExitCode)
	}
	run.CompletedAt = time.Now()
	result.Run = run

	// Note: Stack restart happens in defer

	return result, nil
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
