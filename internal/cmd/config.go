package cmd

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config subcommand group
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration operations",
		Long:  "Operations for validating and inspecting Bosun configuration.",
	}

	// Add subcommands
	cmd.AddCommand(NewValidateCmd())

	return cmd
}
