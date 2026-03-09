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

const miniPollInterval = 4 * time.Second

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
		// Re-render art at new size.
		m.artTrackID = ""

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
		return center(m.width, fmt.Sprintf("⚠ %s", m.err))
	}

	if m.state == nil || m.state.Item == nil {
		return center(m.width, "♫ Nothing playing")
	}

	track := m.state.Item
	var sections []string

	// Album art, centered.
	if m.artCache != "" {
		sections = append(sections, m.artCache)
	}

	// Track info.
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorText).Render(track.Name)
	artist := lipgloss.NewStyle().Foreground(ColorMutedText).Render(track.ArtistNames())
	sections = append(sections, title)
	sections = append(sections, artist)

	// Progress bar.
	sections = append(sections, m.renderProgress())

	// Play state indicator.
	icon := "▶"
	if m.state.IsPlaying {
		icon = "⏸"
	}
	hint := lipgloss.NewStyle().Foreground(ColorMediumGray).Render(
		fmt.Sprintf("%s  spc:play/pause n:next p:prev q:quit", icon),
	)
	sections = append(sections, hint)

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	// Center the whole thing vertically and horizontally.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m MiniPlayer) renderProgress() string {
	if m.state == nil || m.state.Item == nil {
		return ""
	}

	elapsed := time.Duration(m.state.ProgressMS) * time.Millisecond
	total := time.Duration(m.state.Item.Duration) * time.Millisecond

	barWidth := min(m.width-16, 30)
	if barWidth < 5 {
		barWidth = 5
	}

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
	imageURL := images[0].URL // largest
	if len(images) > 1 {
		imageURL = images[1].URL // medium
	}
	trackID := track.ID
	artW := min(m.width-4, 30)
	if artW < 8 {
		artW = 8
	}
	return func() tea.Msg {
		rendered, err := art.RenderFromURL(imageURL, artW)
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

func center(width int, s string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, s)
}
