package tui

import "github.com/charmbracelet/lipgloss"

// Spotify-inspired color palette.
const (
	ColorSpotifyGreen = lipgloss.Color("#1DB954")
	ColorBackground   = lipgloss.Color("#191414")
	ColorText         = lipgloss.Color("#FFFFFF")
	ColorMutedText    = lipgloss.Color("#B3B3B3")
	ColorDarkGray     = lipgloss.Color("#282828")
	ColorMediumGray   = lipgloss.Color("#404040")
)

// Layout constants.
const (
	PanelMinWidth  = 40
	PanelMaxWidth  = 120
	DefaultPadding = 1
	TabBarHeight   = 3
)

// Theme holds all the styles used throughout the TUI.
type Theme struct {
	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style

	NowPlaying  lipgloss.Style
	ProgressBar lipgloss.Style
	ControlsBar lipgloss.Style

	SelectedItem   lipgloss.Style
	UnselectedItem lipgloss.Style

	SearchInput lipgloss.Style
	HelpBar     lipgloss.Style
	StatusBar   lipgloss.Style
}

// DefaultTheme returns the default Spotify-green theme.
func DefaultTheme() Theme {
	return Theme{
		ActiveTab: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSpotifyGreen).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorSpotifyGreen).
			Padding(0, 2),

		InactiveTab: lipgloss.NewStyle().
			Foreground(ColorMutedText).
			Padding(0, 2),

		NowPlaying: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSpotifyGreen).
			Padding(1, 2),

		ProgressBar: lipgloss.NewStyle().
			Foreground(ColorSpotifyGreen),

		ControlsBar: lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1),

		SelectedItem: lipgloss.NewStyle().
			Foreground(ColorSpotifyGreen).
			Bold(true).
			PaddingLeft(2),

		UnselectedItem: lipgloss.NewStyle().
			Foreground(ColorMutedText).
			PaddingLeft(2),

		SearchInput: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSpotifyGreen).
			Padding(0, 1).
			Width(40),

		HelpBar: lipgloss.NewStyle().
			Foreground(ColorMutedText).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorDarkGray).
			Padding(0, 1),
	}
}
