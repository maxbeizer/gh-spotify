package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-spotify/internal/spotify"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available playback devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		devices, err := client.GetDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			fmt.Println("No devices available.")
			return nil
		}
		for _, d := range devices {
			active := " "
			if d.IsActive {
				active = "▶"
			}
			fmt.Printf(" %s %-30s %-12s Vol: %d%%\n", active, d.Name, d.Type, d.VolumePercent)
		}
		return nil
	},
}

var deviceCmd = &cobra.Command{
	Use:   "device <name>",
	Short: "Switch playback to a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return err
		}
		devices, err := client.GetDevices()
		if err != nil {
			return err
		}

		target := args[0]
		for _, d := range devices {
			if d.Name == target {
				if err := client.TransferPlayback(d.ID); err != nil {
					return err
				}
				fmt.Printf("🎧 Switched playback to %s\n", d.Name)
				return nil
			}
		}
		return fmt.Errorf("device %q not found — run 'gh spotify devices' to see available devices", target)
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	rootCmd.AddCommand(deviceCmd)
}
