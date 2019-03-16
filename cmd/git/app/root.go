package app

import (
	"github.com/spf13/cobra"
)

// NewGitHubCommand creates a new GitHub command.
func NewGitHubCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "git",
		Short: "Performs git related operations",
	}

	// Add subcommands
	rootCmd.AddCommand(newGlobCloneCommand())
	return rootCmd
}
