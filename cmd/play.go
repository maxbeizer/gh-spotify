package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Resume playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		if err := client.Play(); err != nil {
			return err
		}
		fmt.Println("▶ Playback resumed")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}
