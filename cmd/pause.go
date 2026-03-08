package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		if err := client.Pause(); err != nil {
			return err
		}
		fmt.Println("⏸ Playback paused")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
