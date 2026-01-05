package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/adapters/joblabels"
	"github.com/simone-viozzi/bosun/internal/app/planner"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// NewPlanShowCmd creates the `plan show` subcommand.
func NewPlanShowCmd() *cobra.Command {
	var (
		format         string
		includeStopped bool
		projectFilter  string
		stackFilter    string
	)

	cmd := &cobra.Command{
		Use:   "show <job-name>",
		Short: "Show execution plan for a job",
		Long: `Shows the execution plan for a specific job.

The plan displays the exact steps Bosun would take to execute the job:
1. Stop containers (if any are targeted)
2. Run the worker container with attached volumes
3. Restart containers

The plan is deterministic - running this command multiple times with
the same job configuration will produce identical output.`,
		Example: `  # Show plan for a job
  bosun plan show daily-job

  # Show plan in JSON format
  bosun plan show daily-job --format json

  # Include stopped containers in discovery
  bosun plan show daily-job --stopped

  # Show plan for job in specific project
  bosun plan show daily-job --project myapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			jobName := args[0]
			exitCode, err := runPlanShow(ctx, jobName, format, includeStopped, projectFilter, stackFilter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			if exitCode != ExitSuccess {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json, yaml")
	cmd.Flags().BoolVar(&includeStopped, "stopped", false, "Include stopped containers in discovery")
	cmd.Flags().StringVar(&projectFilter, "project", "", "Filter by Docker Compose project name")
	cmd.Flags().StringVar(&stackFilter, "stack", "", "Filter by bosun.stack label value")

	return cmd
}

// runPlanShow executes the plan show command logic.
func runPlanShow(ctx context.Context, jobName string, format string, includeStopped bool, projectFilter, stackFilter string) (int, error) {
	// Validate format
	format = strings.ToLower(format)
	if format != "text" && format != "json" && format != "yaml" {
		return ExitRuntimeError, fmt.Errorf("invalid format %q: must be text, json, or yaml", format)
	}

	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		return ExitRuntimeError, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	// Create selector with filters
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: includeStopped,
	}
	if projectFilter != "" {
		selector.ProjectFilter = []string{projectFilter}
	}
	if stackFilter != "" {
		selector.StackFilter = []string{stackFilter}
	}

	// Get snapshot
	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		return ExitRuntimeError, fmt.Errorf("failed to get Docker snapshot: %w", err)
	}

	// Discover jobs
	discoverer := joblabels.NewDiscoverer()
	foundJobs, _, err := discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		return ExitRuntimeError, fmt.Errorf("failed to discover jobs: %w", err)
	}

	// Find the requested job
	var targetJob *jobs.Job
	for i := range foundJobs {
		if foundJobs[i].Name == jobName {
			targetJob = &foundJobs[i]
			break
		}
	}

	if targetJob == nil {
		// Job not found - show available jobs and suggest bosun plan list
		return ExitValidationError, formatJobNotFoundError(jobName, foundJobs)
	}

	// Generate execution plan
	p := planner.New()
	plan, err := p.Plan(ctx, *targetJob)
	if err != nil {
		// Check for specific errors
		if err == ports.ErrOrphanedDependents {
			return ExitValidationError, formatOrphanedDependentsError(jobName, err)
		}
		return ExitRuntimeError, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Output results
	switch format {
	case "text":
		renderPlanTextOutput(plan, *targetJob)
	case "json":
		if err := renderPlanJSONOutput(plan); err != nil {
			return ExitRuntimeError, fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := renderPlanYAMLOutput(plan); err != nil {
			return ExitRuntimeError, fmt.Errorf("failed to encode YAML: %w", err)
		}
	}

	return ExitSuccess, nil
}

// formatJobNotFoundError creates a helpful error message when a job is not found.
func formatJobNotFoundError(jobName string, availableJobs []jobs.Job) error {
	if len(availableJobs) == 0 {
		return fmt.Errorf("job %q not found\n\nNo jobs discovered. To define a job, add these labels to a container:\n  bosun.job.enabled: \"true\"\n  bosun.job.name: \"my-job\"\n\nRun 'bosun plan list' to see discovered jobs", jobName)
	}

	// Sort jobs by name for consistent output
	sort.Slice(availableJobs, func(i, j int) bool {
		return availableJobs[i].Name < availableJobs[j].Name
	})

	names := make([]string, len(availableJobs))
	for i, j := range availableJobs {
		names[i] = j.Name
	}

	return fmt.Errorf("job %q not found\n\nAvailable jobs:\n  %s\n\nRun 'bosun plan list' to see all discovered jobs", jobName, strings.Join(names, "\n  "))
}

// formatOrphanedDependentsError creates a helpful error for orphan dependent errors.
func formatOrphanedDependentsError(jobName string, err error) error {
	return fmt.Errorf("cannot generate plan for job %q: %w (stopping the targeted containers would leave dependent containers running; either include the dependent containers in the job, or configure them to be stopped separately)", jobName, err)
}

// renderPlanTextOutput renders an execution plan in human-readable format.
func renderPlanTextOutput(plan jobs.ExecutionPlan, job jobs.Job) {
	fmt.Printf("Execution Plan: %s\n", plan.JobName)
	fmt.Printf("Schedule: %s\n", job.Schedule)
	fmt.Printf("Worker Image: %s\n", job.WorkerImage)
	fmt.Println("")

	if len(job.TargetStacks) > 0 {
		fmt.Printf("Target Stacks: %s\n", strings.Join(job.TargetStacks, ", "))
	}
	if len(job.TargetContainers) > 0 {
		fmt.Printf("Target Containers: %d\n", len(job.TargetContainers))
	}
	if len(job.AttachVolumes) > 0 {
		fmt.Printf("Attached Volumes: %d\n", len(job.AttachVolumes))
	}
	fmt.Println("")

	fmt.Printf("Steps (%d):\n", len(plan.Steps))
	fmt.Println(strings.Repeat("-", 60))

	for i, step := range plan.Steps {
		fmt.Printf("\n%d. [%s] %s\n", i+1, step.Type, step.Description)

		// Show details based on step type
		switch step.Type {
		case jobs.StepTypeStopContainers:
			if step.UseComposeStop {
				fmt.Printf("   Method: docker compose stop (project: %s)\n", step.ComposeProject)
			} else {
				fmt.Printf("   Method: docker stop\n")
			}
			if len(step.ContainerIDs) > 0 && len(step.ContainerIDs) <= 5 {
				fmt.Printf("   Containers: %s\n", strings.Join(step.ContainerIDs, ", "))
			}

		case jobs.StepTypeRunWorker:
			fmt.Printf("   Image: %s\n", step.WorkerImage)
			if len(step.VolumeMounts) > 0 {
				fmt.Println("   Volume Mounts:")
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				for _, vol := range step.VolumeMounts {
					_, _ = fmt.Fprintf(w, "     - %s\t→ %s\t(%s)\n", vol.Name, vol.MountPath, vol.Mode)
				}
				_ = w.Flush()
			}

		case jobs.StepTypeStartContainers:
			fmt.Printf("   (Start step - not yet implemented)\n")
		}
	}

	fmt.Println("")
}

// renderPlanJSONOutput renders an execution plan as JSON.
func renderPlanJSONOutput(plan jobs.ExecutionPlan) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}

// renderPlanYAMLOutput renders an execution plan as YAML.
func renderPlanYAMLOutput(plan jobs.ExecutionPlan) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(plan)
}

func init() {
	// Register show command with plan command group
}
