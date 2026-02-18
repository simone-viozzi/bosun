package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/simone-viozzi/bosun/internal/app"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// NewJobListCmd creates the `bosun job list` command.
// Shows all currently scheduled jobs with their status, schedule,
// last run time, next run time, and overlap policy.
func NewJobListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs discovered from Docker labels",
		Long: `Lists all jobs discovered from Docker labels, showing:
  - Job name and cron schedule
  - Overlap policy (queue or skip)
  - Enabled status

Discovers jobs by reading Docker container labels.`,
		Example: `  # List jobs as text table
  bosun job list

  # List jobs as JSON
  bosun job list --format json

  # List jobs as YAML
  bosun job list --format yaml`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runJobList(cmd.Context(), format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json, or yaml")

	return cmd
}

// jobListEntry is the output representation of a discovered job.
type jobListEntry struct {
	Name          string `json:"name"   yaml:"name"`
	Schedule      string `json:"schedule"   yaml:"schedule"`
	OverlapPolicy string `json:"overlapPolicy" yaml:"overlapPolicy"`
	Enabled       bool   `json:"enabled"  yaml:"enabled"`
	Stacks        string `json:"stacks"   yaml:"stacks"`
}

// runJobList discovers jobs from Docker labels and renders them.
func runJobList(ctx context.Context, format string, w io.Writer) error {
	svc, err := app.Bootstrap(ctx, app.BootstrapOptions{})
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}
	defer func() { _ = svc.Close() }()

	snapshot, err := svc.LabelSource.Snapshot(ctx, ports.Selector{
		Prefixes: []string{dlabels.DefaultLabelPrefix},
	})
	if err != nil {
		return fmt.Errorf("label snapshot failed: %w", err)
	}

	discovered, _, err := svc.Discoverer.DiscoverJobs(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("job discovery failed: %w", err)
	}

	entries := make([]jobListEntry, 0, len(discovered))
	for _, j := range discovered {
		stacks := "-"
		if len(j.TargetStacks) > 0 {
			stacks = strings.Join(j.TargetStacks, ", ")
		}
		entries = append(entries, jobListEntry{
			Name:          j.Name,
			Schedule:      j.Schedule,
			OverlapPolicy: string(j.OverlapPolicy),
			Enabled:       j.Enabled,
			Stacks:        stacks,
		})
	}

	// Sort by name for deterministic output.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	switch format {
	case "json":
		return renderJSON(w, entries)
	case "yaml":
		return renderYAML(w, entries)
	case "text":
		renderTextTable(w, entries)
		return nil
	default:
		return fmt.Errorf("unsupported format: %q (use text, json, or yaml)", format)
	}
}

// renderTextTable writes a formatted text table to w.
func renderTextTable(w io.Writer, entries []jobListEntry) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "No jobs discovered.")
		return
	}

	// Column headers and widths.
	headers := []string{"NAME", "SCHEDULE", "OVERLAP", "ENABLED", "STACKS"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	// Compute column widths from data.
	rows := make([][]string, len(entries))
	for i, e := range entries {
		enabled := "true"
		if !e.Enabled {
			enabled = "false"
		}
		row := []string{e.Name, e.Schedule, e.OverlapPolicy, enabled, e.Stacks}
		rows[i] = row
		for j, cell := range row {
			if len(cell) > widths[j] {
				widths[j] = len(cell)
			}
		}
	}

	// Print header.
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(w, "  ")
		}
		_, _ = fmt.Fprintf(w, "%-*s", widths[i], h)
	}
	_, _ = fmt.Fprintln(w)

	// Print rows.
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(w, "  ")
			}
			_, _ = fmt.Fprintf(w, "%-*s", widths[i], cell)
		}
		_, _ = fmt.Fprintln(w)
	}
}

// renderJSON writes entries as JSON to w.
func renderJSON(w io.Writer, entries []jobListEntry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// renderYAML writes entries as YAML to w.
func renderYAML(w io.Writer, entries []jobListEntry) error {
	return yaml.NewEncoder(w).Encode(entries)
}
