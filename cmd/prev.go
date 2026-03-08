package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Go to previous track",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		if err := client.Previous(); err != nil {
			return err
		}
		fmt.Println("⏮ Previous track")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(prevCmd)
}
