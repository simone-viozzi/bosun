package cmd

import (
	"github.com/spf13/cobra"
)

// NewLabelsCmd creates the labels subcommand
func NewLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labels",
		Short: "Docker label inspection",
		Long: `Operations for inspecting Docker labels used by Bosun.

The 'labels snapshot' command captures a JSON dump of all Docker entities
(containers, volumes, networks) with bosun.* labels, useful for debugging
job discovery or understanding your label configuration.

Examples:
  bosun labels snapshot              # Dump all labeled entities as JSON
  bosun labels snapshot --stopped    # Include stopped containers
  bosun labels snapshot --project x  # Filter by Compose project`,
	}

	// Add subcommands
	cmd.AddCommand(NewSnapshotCmd())

	return cmd
}
