package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-spotify/internal/tui"
	"github.com/spf13/cobra"
)

var miniCmd = &cobra.Command{
	Use:   "mini",
	Short: "Compact now-playing display for small terminals",
	Long:  "A minimal, auto-updating now-playing view with album art.\nDesigned for small terminal panes (tmux/cmux sidebars).\n\nKeys: space (play/pause), n (next), p (prev), q (quit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		mini, err := tui.NewMiniPlayer()
		if err != nil {
			return fmt.Errorf("failed to start mini player: %w", err)
		}
		p := tea.NewProgram(mini, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(miniCmd)
}
