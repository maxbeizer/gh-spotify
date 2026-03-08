package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Show currently playing track",
	RunE:  runNow,
}

var jsonFlag bool
var jqFlag string

func init() {
	nowCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	nowCmd.Flags().StringVar(&jqFlag, "jq", "", "jq expression to apply to JSON output")
	rootCmd.AddCommand(nowCmd)
}

func runNow(cmd *cobra.Command, args []string) error {
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

	if jsonFlag {
		data, _ := json.MarshalIndent(state, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	track := state.Item
	progress := time.Duration(state.ProgressMS) * time.Millisecond
	duration := time.Duration(track.Duration) * time.Millisecond
	status := "▶"
	if !state.IsPlaying {
		status = "⏸"
	}

	fmt.Printf("%s %s — %s\n", status, track.Name, track.ArtistNames())
	fmt.Printf("  Album: %s\n", track.Album.Name)
	fmt.Printf("  %s / %s\n", formatDuration(progress), formatDuration(duration))
	return nil
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
