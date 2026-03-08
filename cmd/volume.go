package cmd

import (
	"fmt"
	"strconv"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var volumeCmd = &cobra.Command{
	Use:   "volume <0-100>",
	Short: "Set playback volume",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vol, err := strconv.Atoi(args[0])
		if err != nil || vol < 0 || vol > 100 {
			return fmt.Errorf("volume must be a number between 0 and 100")
		}

		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		if err := client.SetVolume(vol); err != nil {
			return err
		}
		fmt.Printf("🔊 Volume set to %d%%\n", vol)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
}
