package cmd

import (
	"github.com/spf13/cobra"
)

// NewLabelsCmd creates the labels subcommand
func NewLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labels",
		Short: "Label operations",
		Long:  "Operations for inspecting and managing Docker labels.",
	}

	// Add subcommands
	cmd.AddCommand(NewSnapshotCmd())

	return cmd
}
