package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		if err := client.Next(); err != nil {
			return err
		}
		fmt.Println("⏭ Skipped to next track")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}
