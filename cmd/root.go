package cmd

import (
	"fmt"
	"os"

	"github.com/nhan-ng/alfred/cmd/git/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "My personal butler.",
}

func init() {
	rootCmd.AddCommand(app.NewGitHubCommand())
}

// Execute executes
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
