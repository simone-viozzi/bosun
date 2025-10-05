package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"github.com/spf13/cobra"
)

// NewSnapshotCmd creates the snapshot subcommand
func NewSnapshotCmd() *cobra.Command {
	var includeStopped bool

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Print current label snapshot as JSON",
		Long:  "Captures and prints a snapshot of all Docker entities with Bosun labels as pretty-printed JSON.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runSnapshot(ctx, includeStopped)
		},
	}

	cmd.Flags().BoolVar(&includeStopped, "stopped", false, "Include stopped containers in the snapshot")

	return cmd
}

func runSnapshot(ctx context.Context, includeStopped bool) error {
	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w\nIs Docker running?", err)
	}

	// Create selector with default prefix
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: includeStopped,
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
