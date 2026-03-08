package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tracks, albums, artists, and playlists",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		result, err := client.Search(query, []string{"track", "artist", "album", "playlist"}, 10)
		if err != nil {
			return err
		}

		if len(result.Tracks.Items) > 0 {
			fmt.Println("🎵 Tracks:")
			for i, t := range result.Tracks.Items {
				fmt.Printf("  %d. %s — %s\n", i+1, t.Name, t.ArtistNames())
			}
			fmt.Println()
		}

		if len(result.Artists.Items) > 0 {
			fmt.Println("🎤 Artists:")
			for i, a := range result.Artists.Items {
				fmt.Printf("  %d. %s\n", i+1, a.Name)
			}
			fmt.Println()
		}

		if len(result.Albums.Items) > 0 {
			fmt.Println("💿 Albums:")
			for i, a := range result.Albums.Items {
				fmt.Printf("  %d. %s\n", i+1, a.Name)
			}
			fmt.Println()
		}

		if len(result.Playlists.Items) > 0 {
			fmt.Println("📋 Playlists:")
			for i, p := range result.Playlists.Items {
				fmt.Printf("  %d. %s (%d tracks)\n", i+1, p.Name, p.TrackCount())
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
