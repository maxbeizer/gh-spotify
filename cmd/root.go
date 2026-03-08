package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-spotify/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-spotify",
	Short: "Control Spotify from the terminal",
	Long:  "A gh CLI extension with a full TUI for controlling Spotify from the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tui.NewApp()
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
