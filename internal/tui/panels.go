package tui

import "github.com/maxbeizer/gh-spotify/internal/spotify"

// PanelSet holds all sub-panel models for the TUI app.
// This provides a convenient way for app.go to integrate the panels
// without directly depending on each panel's constructor.
type PanelSet struct {
	Playlists PlaylistModel
	Search    SearchModel
	Devices   DevicesModel
}

// NewPanelSet creates all panel models with the given client and theme.
// The client may be nil if the user is not authenticated.
func NewPanelSet(client *spotify.Client, theme Theme) PanelSet {
	return PanelSet{
		Playlists: NewPlaylistModel(client, theme),
		Search:    NewSearchModel(client, theme),
		Devices:   NewDevicesModel(client, theme),
	}
}

// SetSize propagates a size change to all panels.
func (ps *PanelSet) SetSize(w, h int) {
	ps.Playlists.SetSize(w, h)
	ps.Search.SetSize(w, h)
	ps.Devices.SetSize(w, h)
}
