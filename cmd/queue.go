package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue [track-uri]",
	Short: "View or add to the playback queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			if err := client.AddToQueue(args[0]); err != nil {
				return err
			}
			fmt.Println("✓ Added to queue")
			return nil
		}

		// Show currently playing as a simple queue view.
		state, err := client.GetCurrentlyPlaying()
		if err != nil {
			return err
		}
		if state.Item == nil {
			fmt.Println("Nothing is currently playing.")
			return nil
		}
		fmt.Printf("Now playing: %s — %s\n", state.Item.Name, state.Item.ArtistNames())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(queueCmd)
}
