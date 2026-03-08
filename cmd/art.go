package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/art"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var artCmd = &cobra.Command{
	Use:   "art",
	Short: "Display ASCII album art of the current track",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		state, err := client.GetCurrentlyPlaying()
		if err != nil {
			return err
		}

		if state.Item == nil {
			fmt.Println("Nothing is currently playing.")
			return nil
		}

		track := state.Item

		// Pick the largest album image (Spotify returns them largest-first).
		if len(track.Album.Images) == 0 {
			fmt.Println("No album art available for the current track.")
			return nil
		}
		imageURL := track.Album.Images[0].URL

		rendered, err := art.RenderFromURL(imageURL, 40)
		if err != nil {
			return fmt.Errorf("rendering album art: %w", err)
		}

		fmt.Print(rendered)
		fmt.Printf("\n%s — %s\n", track.Name, track.ArtistNames())
		fmt.Printf("%s\n", track.Album.Name)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(artCmd)
}
