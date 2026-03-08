package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var artCmd = &cobra.Command{
	Use:   "art",
	Short: "Display ASCII album art of the current track",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement album art renderer (Phase 4)
		fmt.Println("Album art rendering is not yet implemented.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(artCmd)
}
