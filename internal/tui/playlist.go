package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

// Playlist panel view states.
const (
	playlistViewList   = iota // browsing playlists
	playlistViewTracks        // viewing tracks in a selected playlist
)

// playlistsLoadedMsg is sent when playlists finish loading.
type playlistsLoadedMsg struct {
	playlists []spotify.Playlist
	err       error
}

// tracksLoadedMsg is sent when playlist tracks finish loading.
type tracksLoadedMsg struct {
	tracks []spotify.Track
	err    error
}

// PlaylistModel is the playlist browser panel.
type PlaylistModel struct {
	client    *spotify.Client
	theme     Theme
	width     int
	height    int
	view      int
	playlists []spotify.Playlist
	tracks    []spotify.Track
	cursor    int
	selected  *spotify.Playlist
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
	m.loading = true
	return m.loadPlaylists
}

func (m PlaylistModel) loadPlaylists() tea.Msg {
	resp, err := m.client.GetPlaylists(50, 0)
	if err != nil {
		return playlistsLoadedMsg{err: err}
	}
	return playlistsLoadedMsg{playlists: resp.Items}
}

func (m PlaylistModel) loadTracks(playlistID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.GetPlaylistTracks(playlistID, 50, 0)
		if err != nil {
			return tracksLoadedMsg{err: err}
		}
		tracks := make([]spotify.Track, len(resp.Items))
		for i, item := range resp.Items {
			tracks[i] = item.Track
		}
		return tracksLoadedMsg{tracks: tracks}
	}
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

	case tracksLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.tracks = msg.tracks
		m.cursor = 0

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.cursor = min(m.cursor+1, m.listLen()-1)
		case "enter":
			if m.view == playlistViewList && m.cursor < len(m.playlists) {
				m.selected = &m.playlists[m.cursor]
				m.view = playlistViewTracks
				m.loading = true
				m.cursor = 0
				return m, m.loadTracks(m.selected.ID)
			}
		case "esc":
			if m.view == playlistViewTracks {
				m.view = playlistViewList
				m.tracks = nil
				m.selected = nil
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m PlaylistModel) listLen() int {
	if m.view == playlistViewTracks {
		return max(len(m.tracks), 1)
	}
	return max(len(m.playlists), 1)
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

	if m.view == playlistViewTracks {
		return m.renderTracks()
	}
	return m.renderPlaylists()
}

func (m PlaylistModel) renderPlaylists() string {
	if len(m.playlists) == 0 {
		return m.theme.UnselectedItem.Render("No playlists found")
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		PaddingLeft(2).
		Render("Your Playlists")

	lines := header + "\n\n"

	visible := m.visibleRange(len(m.playlists))
	for i := visible.start; i < visible.end; i++ {
		p := m.playlists[i]
		label := fmt.Sprintf("%s (%d tracks)", p.Name, p.Tracks.Total)
		if i == m.cursor {
			lines += m.theme.SelectedItem.Render("▸ "+label) + "\n"
		} else {
			lines += m.theme.UnselectedItem.Render("  "+label) + "\n"
		}
	}

	return lines
}

func (m PlaylistModel) renderTracks() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		PaddingLeft(2).
		Render(fmt.Sprintf("♫ %s", m.selected.Name))

	hint := m.theme.HelpBar.Render("esc: back")
	lines := header + "\n" + hint + "\n\n"

	if len(m.tracks) == 0 {
		lines += m.theme.UnselectedItem.Render("  No tracks")
		return lines
	}

	visible := m.visibleRange(len(m.tracks))
	for i := visible.start; i < visible.end; i++ {
		t := m.tracks[i]
		label := fmt.Sprintf("%s — %s", t.Name, t.ArtistNames())
		if i == m.cursor {
			lines += m.theme.SelectedItem.Render("▸ "+label) + "\n"
		} else {
			lines += m.theme.UnselectedItem.Render("  "+label) + "\n"
		}
	}

	return lines
}

type visibleWindow struct{ start, end int }

func (m PlaylistModel) visibleRange(total int) visibleWindow {
	maxVisible := m.height - 6
	if maxVisible <= 0 || maxVisible >= total {
		return visibleWindow{0, total}
	}
	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return visibleWindow{start, end}
}
