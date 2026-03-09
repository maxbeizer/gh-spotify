package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-spotify/internal/tui"
	"github.com/spf13/cobra"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "gh-spotify",
	Short: "Control Spotify from the terminal",
	Long:  "A gh CLI extension with a full TUI for controlling Spotify from the terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tui.NewApp()
		opts := []tea.ProgramOption{tea.WithAltScreen()}
		if debugMode {
			f, err := os.OpenFile("gh-spotify-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to open debug log: %w", err)
			}
			defer f.Close()
			opts = append(opts, tea.WithOutput(os.Stderr))
			// Also enable key debug in the app.
			app.SetDebug(f)
			fmt.Fprintln(os.Stderr, "Debug mode: logging to gh-spotify-debug.log")
		}
		p := tea.NewProgram(app, opts...)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Log key events to gh-spotify-debug.log")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
