package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for bosun
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bosun",
		Short: "Bosun - A Docker-label-driven job orchestrator",
		Long: `Bosun is a Docker-label-driven job orchestrator for single hosts.

It discovers jobs defined via container labels, generates execution plans
that safely stop containers, run worker containers with attached volumes,
and start everything back up.

Bosun uses a label-based approach: add bosun.job.* labels to your containers
to define jobs, then use 'bosun plan list' to see discovered jobs and
'bosun plan show <job>' to preview the execution plan.`,
		SilenceErrors: true, // Errors are handled by main.go, not by Cobra
	}

	// Add subcommands
	cmd.AddCommand(NewLabelsCmd())
	cmd.AddCommand(NewConfigCmd())
	cmd.AddCommand(NewPlanCmd())
	cmd.AddCommand(NewJobCmd())    // M3: Job execution
	cmd.AddCommand(NewDaemonCmd()) // M4: Long-lived daemon

	return cmd
}
