package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var repeatCmd = &cobra.Command{
	Use:   "repeat [off|track|context]",
	Short: "Set repeat mode",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		var mode string
		if len(args) == 0 {
			// Cycle: off → context → track → off.
			ps, err := client.GetPlayerState()
			if err != nil {
				return err
			}
			switch ps.RepeatState {
			case "off":
				mode = "context"
			case "context":
				mode = "track"
			default:
				mode = "off"
			}
		} else {
			mode = strings.ToLower(args[0])
			if mode != "off" && mode != "track" && mode != "context" {
				return fmt.Errorf("usage: gh spotify repeat [off|track|context]")
			}
		}

		if err := client.SetRepeat(mode); err != nil {
			return err
		}

		icons := map[string]string{"off": "➡️", "context": "🔁", "track": "🔂"}
		fmt.Printf("%s Repeat: %s\n", icons[mode], mode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(repeatCmd)
}
