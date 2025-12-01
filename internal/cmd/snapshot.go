package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// NewSnapshotCmd creates the snapshot subcommand
func NewSnapshotCmd() *cobra.Command {
	var (
		includeStopped bool
		projectFilter  string
		stackFilter    string
	)

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Print current label snapshot as JSON",
		Long:  "Captures and prints a snapshot of all Docker entities with Bosun labels as pretty-printed JSON.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runSnapshot(ctx, includeStopped, projectFilter, stackFilter)
		},
	}

	cmd.Flags().BoolVar(&includeStopped, "stopped", false, "Include stopped containers in the snapshot")
	cmd.Flags().StringVar(&projectFilter, "project", "", "Filter by Docker Compose project name (com.docker.compose.project label)")
	cmd.Flags().StringVar(&stackFilter, "stack", "", "Filter by Bosun stack name (bosun.stack label)")

	return cmd
}

func runSnapshot(ctx context.Context, includeStopped bool, projectFilter, stackFilter string) error {
	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w\nIs Docker running?", err)
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
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Print as pretty JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snapshot); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
