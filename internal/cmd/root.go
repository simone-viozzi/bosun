package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for bosun
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bosun",
		Short: "Bosun - Docker label management tool",
		Long:  "Bosun is a CLI tool for managing and inspecting Docker labels.",
	}

	// Add subcommands
	cmd.AddCommand(NewLabelsCmd())

	return cmd
}
