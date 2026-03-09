package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/art"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

const (
	miniArtWidth     = 8 // ~8 columns wide, ~4 rows tall (half-block chars double vertical res)
	miniPollInterval = 4 * time.Second
)

// MiniPlayer is a compact now-playing display for small terminal windows.
type MiniPlayer struct {
	client     *spotify.Client
	state      *spotify.PlayerState
	err        error
	width      int
	height     int
	artCache   string
	artTrackID string
}

// NewMiniPlayer creates a new mini player.
func NewMiniPlayer() (MiniPlayer, error) {
	client, err := spotify.NewClient()
	if err != nil {
		return MiniPlayer{}, err
	}
	return MiniPlayer{client: client}, nil
}

func (m MiniPlayer) Init() tea.Cmd {
	return m.poll()
}

func (m MiniPlayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			return m, m.togglePlayPause()
		case "n":
			return m, m.doAction(func() error { return m.client.Next() })
		case "p":
			return m, m.doAction(func() error { return m.client.Previous() })
		}

	case playerStateMsg:
		m.state = msg.state
		m.err = msg.err
		var cmds []tea.Cmd
		cmds = append(cmds, m.tick())
		if msg.state != nil && msg.state.Item != nil && msg.state.Item.ID != m.artTrackID {
			cmds = append(cmds, m.renderArt(msg.state.Item))
		}
		return m, tea.Batch(cmds...)

	case artRenderedMsg:
		m.artTrackID = msg.trackID
		m.artCache = msg.art

	case controlActionMsg:
		// Trigger immediate refresh after an action.
		return m, m.poll()
	}

	return m, nil
}

func (m MiniPlayer) View() string {
	if m.width == 0 {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("⚠ %s", m.err)
	}

	if m.state == nil || m.state.Item == nil {
		return "♫ Nothing playing"
	}

	track := m.state.Item

	// Build the right side: title, artist, progress.
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorText).Render(track.Name)
	artist := lipgloss.NewStyle().Foreground(ColorMutedText).Render(track.ArtistNames())

	icon := "▶ "
	if m.state.IsPlaying {
		icon = "⏸ "
	}

	info := lipgloss.JoinVertical(lipgloss.Left,
		title,
		artist,
		icon+m.renderProgress(),
	)

	// Join art (left) + info (right) horizontally.
	if m.artCache != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.artCache+"  ", info)
	}
	return info
}

func (m MiniPlayer) renderProgress() string {
	if m.state == nil || m.state.Item == nil {
		return ""
	}

	elapsed := time.Duration(m.state.ProgressMS) * time.Millisecond
	total := time.Duration(m.state.Item.Duration) * time.Millisecond

	barWidth := 15
	ratio := 0.0
	if m.state.Item.Duration > 0 {
		ratio = float64(m.state.ProgressMS) / float64(m.state.Item.Duration)
	}
	filled := int(ratio * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := lipgloss.NewStyle().Foreground(ColorSpotifyGreen).Render(strings.Repeat("▓", filled)) +
		lipgloss.NewStyle().Foreground(ColorMediumGray).Render(strings.Repeat("░", barWidth-filled))

	return fmt.Sprintf("%s %d:%02d/%d:%02d",
		bar,
		int(elapsed.Minutes()), int(elapsed.Seconds())%60,
		int(total.Minutes()), int(total.Seconds())%60,
	)
}

func (m MiniPlayer) renderArt(track *spotify.Track) tea.Cmd {
	images := track.Album.Images
	if len(images) == 0 {
		return nil
	}
	// Use smallest image for the tiny thumbnail.
	imageURL := images[len(images)-1].URL
	trackID := track.ID
	return func() tea.Msg {
		rendered, err := art.RenderFromURL(imageURL, miniArtWidth)
		if err != nil {
			return artRenderedMsg{trackID: trackID, art: ""}
		}
		return artRenderedMsg{trackID: trackID, art: rendered}
	}
}

func (m MiniPlayer) poll() tea.Cmd {
	return func() tea.Msg {
		state, err := m.client.GetPlayerState()
		return playerStateMsg{state: state, err: err}
	}
}

func (m MiniPlayer) tick() tea.Cmd {
	return tea.Tick(miniPollInterval, func(_ time.Time) tea.Msg {
		state, err := m.client.GetPlayerState()
		return playerStateMsg{state: state, err: err}
	})
}

func (m MiniPlayer) togglePlayPause() tea.Cmd {
	return func() tea.Msg {
		if m.state != nil && m.state.IsPlaying {
			err := m.client.Pause()
			return controlActionMsg{feedback: "⏸", err: err}
		}
		err := m.client.Play()
		return controlActionMsg{feedback: "▶", err: err}
	}
}

func (m MiniPlayer) doAction(fn func() error) tea.Cmd {
	return func() tea.Msg {
		err := fn()
		return controlActionMsg{feedback: "", err: err}
	}
}


