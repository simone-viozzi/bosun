package cmd

import (
	"github.com/spf13/cobra"
)

// NewPlanCmd creates the plan command group for job planning operations.
func NewPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Job planning and preview",
		Long: `Operations for discovering, listing, and previewing job execution plans.

Bosun discovers jobs from containers labeled with bosun.job.* labels,
then generates execution plans that coordinate container stops and worker execution.

Examples:
  bosun plan list                  # List all discovered jobs
  bosun plan show daily-job        # Preview execution plan for a job
  bosun plan list --project myapp  # List jobs for a specific Compose project`,
	}

	// Add subcommands
	cmd.AddCommand(NewPlanListCmd())
	cmd.AddCommand(NewPlanShowCmd())

	return cmd
}
