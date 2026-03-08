package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var shuffleCmd = &cobra.Command{
	Use:   "shuffle [on|off]",
	Short: "Toggle shuffle mode",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}

		var state bool
		if len(args) == 0 {
			// Toggle: get current state and flip it.
			ps, err := client.GetPlayerState()
			if err != nil {
				return err
			}
			state = !ps.ShuffleState
		} else {
			switch strings.ToLower(args[0]) {
			case "on":
				state = true
			case "off":
				state = false
			default:
				return fmt.Errorf("usage: gh spotify shuffle [on|off]")
			}
		}

		if err := client.SetShuffle(state); err != nil {
			return err
		}

		if state {
			fmt.Println("🔀 Shuffle on")
		} else {
			fmt.Println("🔀 Shuffle off")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shuffleCmd)
}
