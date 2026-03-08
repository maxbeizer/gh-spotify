package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-spotify",
	Short: "Control Spotify from the terminal",
	Long:  "A gh CLI extension with a full TUI for controlling Spotify from the terminal.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
