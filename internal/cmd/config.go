package cmd

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config subcommand group
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration validation",
		Long: `Operations for validating Bosun configuration and job label syntax.

Use 'bosun config validate' to check that all bosun.* labels on your
containers and volumes are syntactically correct and semantically valid.

Examples:
  bosun config validate             # Validate all labels
  bosun config validate --stopped   # Include stopped containers`,
	}

	// Add subcommands
	cmd.AddCommand(NewValidateCmd())

	return cmd
}
