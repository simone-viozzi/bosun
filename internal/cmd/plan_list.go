package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/adapters/joblabels"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Exit codes for plan commands
const (
	exitCodeOK            = 0
	exitCodeValidationErr = 1
	exitCodeDockerUnavail = 2
	exitCodeInternalErr   = 3
)

// NewPlanListCmd creates the `plan list` subcommand.
func NewPlanListCmd() *cobra.Command {
	var (
		format         string
		includeStopped bool
		stackFilter    string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List discovered backup jobs",
		Long: `Lists all backup jobs discovered from Docker container and volume labels.

Jobs are discovered from containers with bosun.job.enabled=true labels.
Multiple containers with the same bosun.job.name are merged into a single job.

Output formats:
  text  - Human-readable table format (default)
  json  - Machine-readable JSON
  yaml  - Machine-readable YAML`,
		Example: `  # List all jobs
  bosun plan list

  # List jobs in JSON format
  bosun plan list --format json

  # List jobs including stopped containers
  bosun plan list --stopped

  # List jobs for a specific stack
  bosun plan list --stack myapp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			exitCode, err := runPlanList(ctx, format, includeStopped, stackFilter)
			if err != nil {
				if exitCode == exitCodeDockerUnavail {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					fmt.Fprintln(os.Stderr, "Is Docker running?")
				} else {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
			if exitCode != exitCodeOK {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json, yaml")
	cmd.Flags().BoolVar(&includeStopped, "stopped", false, "Include stopped containers in discovery")
	cmd.Flags().StringVar(&stackFilter, "stack", "", "Filter jobs by stack name")

	return cmd
}

// runPlanList executes the plan list command logic.
func runPlanList(ctx context.Context, format string, includeStopped bool, stackFilter string) (int, error) {
	// Validate format
	format = strings.ToLower(format)
	if format != "text" && format != "json" && format != "yaml" {
		return exitCodeInternalErr, fmt.Errorf("invalid format %q: must be text, json, or yaml", format)
	}

	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		return exitCodeDockerUnavail, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	// Create selector
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: includeStopped,
	}

	// Get snapshot
	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		return exitCodeDockerUnavail, fmt.Errorf("failed to get Docker snapshot: %w", err)
	}

	// Discover jobs
	discoverer := joblabels.NewDiscoverer()
	foundJobs, validationErrors, err := discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		return exitCodeInternalErr, fmt.Errorf("failed to discover jobs: %w", err)
	}

	// Apply stack filter if specified
	if stackFilter != "" {
		foundJobs = filterJobsByStack(foundJobs, stackFilter)
	}

	// Sort jobs by name for consistent output
	sort.Slice(foundJobs, func(i, j int) bool {
		return foundJobs[i].Name < foundJobs[j].Name
	})

	// Output results
	switch format {
	case "text":
		renderTextOutput(foundJobs, validationErrors)
	case "json":
		if err := renderJSONOutput(foundJobs, validationErrors); err != nil {
			return exitCodeInternalErr, fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := renderYAMLOutput(foundJobs, validationErrors); err != nil {
			return exitCodeInternalErr, fmt.Errorf("failed to encode YAML: %w", err)
		}
	}

	// Return validation error exit code if there were validation errors
	if len(validationErrors) > 0 {
		return exitCodeValidationErr, nil
	}

	return exitCodeOK, nil
}

// filterJobsByStack filters jobs to only those with matching stack names.
func filterJobsByStack(allJobs []jobs.Job, stackFilter string) []jobs.Job {
	var filtered []jobs.Job
	for _, job := range allJobs {
		for _, stack := range job.TargetStacks {
			if stack == stackFilter {
				filtered = append(filtered, job)
				break
			}
		}
	}
	return filtered
}

// renderTextOutput renders jobs in a human-readable table format.
func renderTextOutput(foundJobs []jobs.Job, validationErrors []ports.ValidationError) {
	if len(foundJobs) == 0 && len(validationErrors) == 0 {
		fmt.Println("No jobs discovered.")
		fmt.Println("")
		fmt.Println("To define a backup job, add these labels to a container:")
		fmt.Println("  bosun.job.enabled: \"true\"")
		fmt.Println("  bosun.job.name: \"my-backup\"")
		return
	}

	if len(foundJobs) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSCHEDULE\tCONTAINERS\tVOLUMES\tSTACKS")
		fmt.Fprintln(w, "----\t--------\t----------\t-------\t------")

		for _, job := range foundJobs {
			stacks := "-"
			if len(job.TargetStacks) > 0 {
				stacks = strings.Join(job.TargetStacks, ", ")
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
				job.Name,
				job.Schedule,
				len(job.TargetContainers),
				len(job.AttachVolumes),
				stacks,
			)
		}
		w.Flush()
	}

	if len(validationErrors) > 0 {
		fmt.Println("")
		fmt.Printf("⚠️  %d validation error(s):\n", len(validationErrors))
		for _, ve := range validationErrors {
			fmt.Printf("  - %s %q (%s): %s\n", ve.EntityKind, ve.EntityName, ve.Field, ve.Message)
		}
	}
}

// planListOutput is the structured output for JSON/YAML formats.
type planListOutput struct {
	Jobs             []jobs.Job              `json:"jobs" yaml:"jobs"`
	ValidationErrors []ports.ValidationError `json:"validationErrors,omitempty" yaml:"validationErrors,omitempty"`
}

// renderJSONOutput renders jobs as JSON.
func renderJSONOutput(foundJobs []jobs.Job, validationErrors []ports.ValidationError) error {
	output := planListOutput{
		Jobs:             foundJobs,
		ValidationErrors: validationErrors,
	}
	// Ensure empty slices are serialized as [] not null
	if output.Jobs == nil {
		output.Jobs = []jobs.Job{}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// renderYAMLOutput renders jobs as YAML.
func renderYAMLOutput(foundJobs []jobs.Job, validationErrors []ports.ValidationError) error {
	output := planListOutput{
		Jobs:             foundJobs,
		ValidationErrors: validationErrors,
	}
	// Ensure empty slices are serialized as [] not null
	if output.Jobs == nil {
		output.Jobs = []jobs.Job{}
	}

	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(output)
}
