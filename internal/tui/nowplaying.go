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
	pollInterval = 4 * time.Second
	artWidth     = 22
)

// artRenderedMsg carries the result of async album art rendering.
type artRenderedMsg struct {
	trackID string
	art     string
}

// playerStateMsg carries the result of a Spotify player state poll.
type playerStateMsg struct {
	state *spotify.PlayerState
	err   error
}

// NowPlayingModel displays the current Spotify playback state.
type NowPlayingModel struct {
	client     *spotify.Client
	state      *spotify.PlayerState
	err        error
	theme      Theme
	width      int
	artCache   string // rendered ASCII art
	artTrackID string // track ID the cached art corresponds to
}

// NewNowPlayingModel creates a NowPlayingModel. client may be nil if not authenticated.
func NewNowPlayingModel(client *spotify.Client, theme Theme) NowPlayingModel {
	return NowPlayingModel{
		client: client,
		theme:  theme,
	}
}

func (m NowPlayingModel) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return m.pollPlayerState()
}

func (m NowPlayingModel) Update(msg tea.Msg) (NowPlayingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case playerStateMsg:
		m.state = msg.state
		m.err = msg.err
		var cmds []tea.Cmd
		cmds = append(cmds, m.tickPoll())
		// Render art asynchronously when track changes.
		if msg.state != nil && msg.state.Item != nil && msg.state.Item.ID != m.artTrackID {
			trackID := msg.state.Item.ID
			images := msg.state.Item.Album.Images
			if len(images) > 0 {
				imageURL := images[len(images)-1].URL // smallest image
				if len(images) > 1 {
					imageURL = images[1].URL // medium if available
				}
				cmds = append(cmds, func() tea.Msg {
					rendered, err := art.RenderFromURL(imageURL, artWidth)
					if err != nil {
						return artRenderedMsg{trackID: trackID, art: ""}
					}
					return artRenderedMsg{trackID: trackID, art: rendered}
				})
			}
		}
		return m, tea.Batch(cmds...)

	case artRenderedMsg:
		m.artTrackID = msg.trackID
		m.artCache = msg.art
	}
	return m, nil
}

func (m NowPlayingModel) View() string {
	if m.client == nil {
		return m.theme.NowPlaying.Render("⚠  Not authenticated\n\nRun `gh spotify auth` to connect your Spotify account.")
	}

	if m.err != nil {
		return m.theme.NowPlaying.Render(fmt.Sprintf("⚠  Error fetching playback:\n%s", m.err))
	}

	if m.state == nil || m.state.Item == nil {
		return m.theme.NowPlaying.Render("♫ Nothing is playing\n\nStart playback in Spotify or run `gh spotify play`.")
	}

	s := m.state
	track := s.Item

	// Shuffle & repeat indicators
	shuffleIcon := "🔀 off"
	if s.ShuffleState {
		shuffleIcon = "🔀 on"
	}

	repeatIcon := "🔁 " + s.RepeatState

	// Progress bar
	progressBar := m.renderProgress(s.ProgressMS, track.Duration)

	// Track info lines — no play/pause icon here (it's in the controls bar).
	lines := []string{
		track.Name,
		track.ArtistNames(),
		track.Album.Name,
		"",
		progressBar,
		"",
		fmt.Sprintf("%s   %s   🔊 %d%%", shuffleIcon, repeatIcon, s.Device.VolumePercent),
	}

	info := strings.Join(lines, "\n")

	// Place album art to the left of track info when available.
	var panel string
	if m.artCache != "" {
		panel = lipgloss.JoinHorizontal(lipgloss.Top, m.artCache+"  ", info)
	} else {
		panel = info
	}

	return m.theme.NowPlaying.Render(panel)
}

func (m NowPlayingModel) renderProgress(progressMS, durationMS int) string {
	elapsed := time.Duration(progressMS) * time.Millisecond
	total := time.Duration(durationMS) * time.Millisecond

	ratio := 0.0
	if durationMS > 0 {
		ratio = float64(progressMS) / float64(durationMS)
	}

	barWidth := 30
	filled := int(ratio * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := m.theme.ProgressBar.Render(strings.Repeat("█", filled)) +
		strings.Repeat("░", barWidth-filled)

	return fmt.Sprintf("%s %s / %s",
		bar,
		formatDuration(elapsed),
		formatDuration(total),
	)
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func (m NowPlayingModel) pollPlayerState() tea.Cmd {
	return func() tea.Msg {
		state, err := m.client.GetPlayerState()
		return playerStateMsg{state: state, err: err}
	}
}

func (m NowPlayingModel) tickPoll() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return tea.Tick(pollInterval, func(_ time.Time) tea.Msg {
		state, err := m.client.GetPlayerState()
		return playerStateMsg{state: state, err: err}
	})
}
