package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Browse your playlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		resp, err := client.GetPlaylists(50, 0)
		if err != nil {
			return err
		}

		if len(resp.Items) == 0 {
			fmt.Println("No playlists found.")
			return nil
		}

		fmt.Println("📋 Your Playlists:")
		for i, p := range resp.Items {
			fmt.Printf("  %d. %s (%d tracks)\n", i+1, p.Name, p.TrackCount())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)
}
