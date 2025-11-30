package cmd

import (
	"github.com/spf13/cobra"
)

// NewPlanCmd creates the plan command group for job planning operations.
func NewPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Job planning operations",
		Long:  "Operations for viewing, validating, and previewing backup job execution plans.",
	}

	// Add subcommands
	cmd.AddCommand(NewPlanListCmd())
	cmd.AddCommand(NewPlanShowCmd())

	return cmd
}
