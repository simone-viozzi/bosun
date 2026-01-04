package cmd

// TODO(#141): Move adapter wiring to internal/app/factory or internal/bootstrap,
// CLI should only parse args and call app-layer services.
// See smells #16-18 in wip_smell_milestone3

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/simone-viozzi/bosun/internal/adapters/docker/compose"
	"github.com/simone-viozzi/bosun/internal/adapters/docker/worker"
	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/adapters/joblabels"
	"github.com/simone-viozzi/bosun/internal/app/executor"
	"github.com/simone-viozzi/bosun/internal/app/planner"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// NewJobCmd creates the `job` command.
func NewJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Job execution commands",
		Long: `Operations for running backup jobs.

Bosun executes jobs by stopping the target Compose stack, running a worker
container with attached volumes, and restarting the stack.`,
	}

	// Add subcommands
	cmd.AddCommand(NewJobRunCmd())

	return cmd
}

// NewJobRunCmd creates the `job run` subcommand.
func NewJobRunCmd() *cobra.Command {
	var (
		timeout        string
		stopTimeout    string
		startTimeout   string
		keepStopped    bool
		keepFailed     bool
		quiet          bool
		includeStopped bool
		dryRun         bool
		format         string
		projectFilter  string
	)

	cmd := &cobra.Command{
		Use:   "run <job-name>",
		Short: "Execute a backup job",
		Long: `Executes a backup job by name.

The execution flow:
1. Discover job by name from Docker labels
2. Validate worker image is available
3. Stop target Compose stack
4. Run worker container with attached volumes
5. Restart stack (always, unless --keep-stopped)

The stack is ALWAYS restarted after worker execution, even if the worker
fails. This ensures production services remain available.

Use --dry-run to preview the execution plan without making any changes.`,
		Example: `  # Run a job
  bosun job run daily-backup

  # Preview what would happen (dry run)
  bosun job run daily-backup --dry-run

  # Dry run with JSON output
  bosun job run daily-backup --dry-run --format json

  # Run with custom timeout
  bosun job run daily-backup --timeout 2h

  # Keep stack stopped after execution (maintenance mode)
  bosun job run daily-backup --keep-stopped

  # Keep worker container on failure for debugging
  bosun job run daily-backup --keep-failed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			jobName := args[0]

			// Handle dry-run mode
			if dryRun {
				return runDryRun(ctx, jobName, format, includeStopped, projectFilter)
			}

			exitCode, err := runJobRun(ctx, jobName, jobRunOptions{
				timeout:        timeout,
				stopTimeout:    stopTimeout,
				startTimeout:   startTimeout,
				keepStopped:    keepStopped,
				keepFailed:     keepFailed,
				quiet:          quiet,
				includeStopped: includeStopped,
				dryRun:         dryRun,
				format:         format,
				projectFilter:  projectFilter,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			if exitCode != ExitSuccess {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&timeout, "timeout", "", "Worker execution timeout (e.g., 1h, 30m)")
	cmd.Flags().StringVar(&stopTimeout, "stop-timeout", "", "Stack stop timeout (e.g., 30s, 1m)")
	cmd.Flags().StringVar(&startTimeout, "start-timeout", "", "Stack start timeout (e.g., 30s, 1m)")
	cmd.Flags().BoolVar(&keepStopped, "keep-stopped", false, "Skip stack restart after worker")
	cmd.Flags().BoolVar(&keepFailed, "keep-failed", false, "Preserve worker container on failure")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress log output")
	cmd.Flags().BoolVar(&includeStopped, "stopped", false, "Include stopped containers in discovery")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview execution plan without running")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json (for --dry-run)")
	cmd.Flags().StringVar(&projectFilter, "project", "", "Filter jobs by Docker Compose project name")

	return cmd
}

type jobRunOptions struct {
	timeout        string
	stopTimeout    string
	startTimeout   string
	keepStopped    bool
	keepFailed     bool
	quiet          bool
	includeStopped bool
	dryRun         bool
	format         string // "text" or "json"
	projectFilter  string
}

// runJobRun executes the job run command logic.
func runJobRun(ctx context.Context, jobName string, opts jobRunOptions) (int, error) {
	// Set up signal handler for graceful shutdown (FR-024)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Track if we were interrupted
	interrupted := false
	go func() {
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "\n⚠️  Received %v, shutting down gracefully...\n", sig)
			fmt.Fprintf(os.Stderr, "   Stack will be restarted after worker cleanup.\n")
			interrupted = true
			cancel()
		case <-ctx.Done():
			// Context cancelled by other means
		}
	}()

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ExitRuntimeError, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		return ExitRuntimeError, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	// Create selector
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: opts.includeStopped,
	}
	if opts.projectFilter != "" {
		selector.ProjectFilter = []string{opts.projectFilter}
	}

	// Get snapshot
	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		if interrupted {
			return ExitInterrupted, fmt.Errorf("interrupted during discovery")
		}
		return ExitRuntimeError, fmt.Errorf("failed to get Docker snapshot: %w", err)
	}

	// Create job discoverer
	discoverer := joblabels.NewDiscoverer()

	// Discover all jobs
	allJobs, _, err := discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		if interrupted {
			return ExitInterrupted, fmt.Errorf("interrupted during job discovery")
		}
		return ExitRuntimeError, fmt.Errorf("failed to discover jobs: %w", err)
	}

	// Find job by name
	var targetJob *jobs.Job
	for i := range allJobs {
		if allJobs[i].Name == jobName {
			targetJob = &allJobs[i]
			break
		}
	}

	if targetJob == nil {
		return ExitJobNotFound, fmt.Errorf("job not found: %s", jobName)
	}

	// Create planner
	jobPlanner := planner.New()

	// Create compose controller
	composeController := compose.NewController(dockerClient)

	// Create worker runner
	workerRunner := worker.NewRunner(dockerClient)

	// Create executor
	exec := executor.New(discoverer, jobPlanner, composeController, workerRunner, dockerClient)

	// Build execute options
	executeOpts := ports.DefaultExecuteOptions()
	executeOpts.KeepStopped = opts.keepStopped
	executeOpts.KeepFailedWorker = opts.keepFailed
	executeOpts.Quiet = opts.quiet
	if !opts.quiet {
		executeOpts.LogWriter = os.Stdout
	}

	// Parse timeout strings and set overrides (Phase 6: US4)
	if opts.timeout != "" {
		d, err := time.ParseDuration(opts.timeout)
		if err != nil {
			return ExitValidationError, fmt.Errorf("invalid --timeout value %q: %w", opts.timeout, err)
		}
		executeOpts.TimeoutOverride = d
	}
	if opts.stopTimeout != "" {
		d, err := time.ParseDuration(opts.stopTimeout)
		if err != nil {
			return ExitValidationError, fmt.Errorf("invalid --stop-timeout value %q: %w", opts.stopTimeout, err)
		}
		executeOpts.StopTimeoutOverride = d
	}
	if opts.startTimeout != "" {
		d, err := time.ParseDuration(opts.startTimeout)
		if err != nil {
			return ExitValidationError, fmt.Errorf("invalid --start-timeout value %q: %w", opts.startTimeout, err)
		}
		executeOpts.StartTimeoutOverride = d
	}

	// Execute job
	result, err := exec.ExecuteJob(ctx, *targetJob, executeOpts)

	// Print result
	printExecutionResult(result)

	// Check if interrupted (even if execution completed, signal was received)
	if interrupted {
		return ExitInterrupted, fmt.Errorf("execution interrupted by signal")
	}

	// Map error to exit code
	if err != nil {
		return ExitCodeFromError(err), err
	}

	if result.Run.WorkerExitCode != 0 {
		return ExitWorkerFailed, fmt.Errorf("worker failed with exit code %d", result.Run.WorkerExitCode)
	}

	return ExitSuccess, nil
}

// printExecutionResult prints execution result to stdout.
func printExecutionResult(result ports.ExecutionResult) {
	fmt.Printf("\n=== Job Execution: %s ===\n", result.Run.JobName)
	fmt.Printf("Run ID: %s\n", result.Run.ID)
	fmt.Printf("Status: %s\n", result.Run.Status)
	fmt.Printf("Started: %s\n", result.Run.StartedAt.Format("2006-01-02 15:04:05"))
	if !result.Run.CompletedAt.IsZero() {
		fmt.Printf("Completed: %s\n", result.Run.CompletedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration: %s\n", result.Run.CompletedAt.Sub(result.Run.StartedAt))
	}

	fmt.Printf("\n=== Steps ===\n")
	for _, step := range result.StepResults {
		status := "✓"
		if step.Status != "success" {
			status = "✗"
		}
		fmt.Printf("%s %s (%.2fs)\n", status, step.Step.Description, step.Duration.Seconds())
		if step.Error != "" {
			fmt.Printf("  Error: %s\n", step.Error)
		}
	}

	if result.WorkerLogs != "" {
		fmt.Printf("\n=== Worker Logs ===\n")
		fmt.Print(result.WorkerLogs)
	}

	if result.Run.Error != "" {
		fmt.Printf("\n=== Error ===\n")
		fmt.Printf("%s\n", result.Run.Error)
	}
}

// runDryRun executes a dry-run preview of job execution.
func runDryRun(ctx context.Context, jobName, format string, includeStopped bool, projectFilter string) error {
	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
	}

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to Docker: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to Docker: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Create selector
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: includeStopped,
	}
	if projectFilter != "" {
		selector.ProjectFilter = []string{projectFilter}
	}

	// Get snapshot
	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get Docker snapshot: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Create job discoverer
	discoverer := joblabels.NewDiscoverer()

	// Discover all jobs
	allJobs, _, err := discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover jobs: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Find job by name
	var targetJob *jobs.Job
	for i := range allJobs {
		if allJobs[i].Name == jobName {
			targetJob = &allJobs[i]
			break
		}
	}

	if targetJob == nil {
		fmt.Fprintf(os.Stderr, "Error: job not found: %s\n", jobName)
		os.Exit(ExitJobNotFound)
		return nil
	}

	// Create planner and compose controller for dry-run
	jobPlanner := planner.New()
	composeController := compose.NewController(dockerClient)
	workerRunner := worker.NewRunner(dockerClient)

	// Create executor
	exec := executor.New(discoverer, jobPlanner, composeController, workerRunner, dockerClient)

	// Execute dry-run
	plan, err := exec.DryRunJob(ctx, *targetJob)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to generate execution plan: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Set timestamp
	plan.CreatedAt = time.Now()

	// Output based on format
	switch format {
	case "json":
		return printDryRunJSON(plan)
	default:
		printDryRunText(plan)
		return nil
	}
}

// printDryRunText prints the execution plan in text format.
func printDryRunText(plan jobs.ExecutionPlan) {
	fmt.Printf("\n=== Dry Run: %s ===\n", plan.JobName)
	fmt.Printf("Generated: %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nExecution Plan (%d steps):\n", len(plan.Steps))

	for i, step := range plan.Steps {
		fmt.Printf("\n%d. %s\n", i+1, step.Description)
		fmt.Printf("   Type: %s\n", step.Type)

		if len(step.ContainerNames) > 0 {
			fmt.Printf("   Containers: %v\n", step.ContainerNames)
		}
		if step.WorkerImage != "" {
			fmt.Printf("   Worker Image: %s\n", step.WorkerImage)
		}
		if len(step.VolumeMounts) > 0 {
			fmt.Printf("   Volume Mounts:\n")
			for _, mount := range step.VolumeMounts {
				fmt.Printf("     - %s → %s (%s)\n", mount.Name, mount.MountPath, mount.Mode)
			}
		}
		if step.UseComposeStop {
			fmt.Printf("   Uses: docker compose stop %s\n", step.ComposeProject)
		}
	}

	fmt.Printf("\n✓ Dry run complete. No changes were made.\n")
}

// printDryRunJSON prints the execution plan in JSON format.
func printDryRunJSON(plan jobs.ExecutionPlan) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}
