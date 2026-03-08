package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

// playlistsLoadedMsg is sent when playlists finish loading.
type playlistsLoadedMsg struct {
	playlists []spotify.Playlist
	err       error
}

// PlaylistModel is the playlist browser panel.
type PlaylistModel struct {
	client    *spotify.Client
	theme     Theme
	width     int
	height    int
	playlists []spotify.Playlist
	cursor    int
	loading   bool
	err       error
}

// NewPlaylistModel creates a new playlist browser panel.
func NewPlaylistModel(client *spotify.Client, theme Theme) PlaylistModel {
	return PlaylistModel{
		client: client,
		theme:  theme,
	}
}

// Init starts loading playlists.
func (m PlaylistModel) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return m.loadPlaylists
}

func (m PlaylistModel) loadPlaylists() tea.Msg {
	resp, err := m.client.GetPlaylists(50, 0)
	if err != nil {
		return playlistsLoadedMsg{err: err}
	}
	return playlistsLoadedMsg{playlists: resp.Items}
}

// SetSize updates the panel dimensions.
func (m *PlaylistModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages for the playlist panel.
func (m PlaylistModel) Update(msg tea.Msg) (PlaylistModel, tea.Cmd) {
	if m.client == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case playlistsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.playlists = msg.playlists
		m.cursor = 0

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.playlists)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.playlists) {
				p := m.playlists[m.cursor]
				return m, m.playPlaylist(p.URI, false)
			}
		case "S":
			if m.cursor < len(m.playlists) {
				p := m.playlists[m.cursor]
				return m, m.playPlaylist(p.URI, true)
			}
		case "r":
			m.loading = true
			return m, m.loadPlaylists
		}
	}

	return m, nil
}

func (m PlaylistModel) playPlaylist(uri string, shuffle bool) tea.Cmd {
	return func() tea.Msg {
		err := m.client.PlayContext(uri, shuffle)
		label := "▶ Playing"
		if shuffle {
			label = "🔀 Shuffle playing"
		}
		return controlActionMsg{feedback: label, err: err}
	}
}

// View renders the playlist panel.
func (m PlaylistModel) View() string {
	if m.client == nil {
		return m.theme.UnselectedItem.Render("Not authenticated — run gh spotify auth")
	}
	if m.loading {
		return m.theme.UnselectedItem.Render("Loading...")
	}
	if m.err != nil {
		return m.theme.UnselectedItem.Render(fmt.Sprintf("Error: %s", m.err))
	}
	if len(m.playlists) == 0 {
		return m.theme.UnselectedItem.Render("No playlists found")
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		PaddingLeft(2).
		Render("Your Playlists")

	s := header + "\n\n"

	maxVisible := m.height - 6
	start, end := scrollWindow(len(m.playlists), m.cursor, maxVisible)

	for i := start; i < end; i++ {
		p := m.playlists[i]
		label := fmt.Sprintf("%s (%d tracks)", p.Name, p.TrackCount())
		if i == m.cursor {
			s += m.theme.SelectedItem.Render("▸ "+label) + "\n"
		} else {
			s += m.theme.UnselectedItem.Render("  "+label) + "\n"
		}
	}

	s += "\n" + m.theme.HelpBar.Render("enter: play  S: shuffle play  r: refresh")
	return s
}
