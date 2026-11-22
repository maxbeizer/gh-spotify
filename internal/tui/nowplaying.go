package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/maxbeizer/gh-spotify/internal/art"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

const (
	pollInterval = 4 * time.Second

	// artWidth is in columns; the art is artWidth/2 rows tall, matching the
	// five info lines next to it.
	artWidth = 10
	artGap   = 2
	// minInfoWidth is the narrowest info column worth showing art beside.
	minInfoWidth = 20
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

// View renders the now-playing card to fit within width x height cells. As
// space shrinks it drops the border, then the album art, then lines of text.
func (m NowPlayingModel) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	var lines []string
	showArt := false
	switch {
	case m.client == nil:
		lines = []string{"⚠  Not authenticated", "Run `gh spotify auth` to connect your Spotify account."}
	case m.err != nil:
		lines = []string{"⚠  Error fetching playback:", m.err.Error()}
	case m.state == nil || m.state.Item == nil:
		lines = []string{"♫ Nothing is playing", "Start playback in Spotify or run `gh spotify play`."}
	default:
		showArt = m.artCache != ""
	}

	const borderW, borderH, padW = 2, 2, 2
	bordered := height >= artWidth/2+borderH && width > borderW+padW+minInfoWidth
	innerW := width
	if bordered {
		innerW = width - borderW - padW
	}
	innerH := height
	if bordered {
		innerH = height - borderH
	}

	if showArt && (innerH < artWidth/2 || innerW < artWidth+artGap+minInfoWidth) {
		showArt = false
	}
	infoW := innerW
	if showArt {
		infoW = innerW - artWidth - artGap
	}

	if lines == nil {
		lines = m.trackLines(infoW, innerH)
	}
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, infoW, "…")
	}

	body := strings.Join(lines, "\n")
	if showArt {
		artBlock := lipgloss.NewStyle().MarginRight(artGap).Render(m.artCache)
		body = lipgloss.JoinHorizontal(lipgloss.Center, artBlock, body)
	}

	if !bordered {
		return body
	}
	return m.theme.NowPlaying.Width(width - borderW).Render(body)
}

// trackLines returns the info lines for the current track, most important
// first so callers can trim from the end when space is tight.
func (m NowPlayingModel) trackLines(width, height int) []string {
	track := m.state.Item
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorText).Render(track.Name)
	progress := m.renderProgress(m.state.ProgressMS, track.Duration, width)

	if height < 5 {
		// Too short for the full layout: keep the title, progress, and artist.
		return []string{title, progress, track.ArtistNames()}
	}
	return []string{
		title,
		track.ArtistNames(),
		lipgloss.NewStyle().Foreground(ColorMutedText).Render(track.Album.Name),
		"",
		progress,
	}
}

func (m NowPlayingModel) renderProgress(progressMS, durationMS, width int) string {
	elapsed := time.Duration(progressMS) * time.Millisecond
	total := time.Duration(durationMS) * time.Millisecond
	times := fmt.Sprintf(" %s / %s", formatDuration(elapsed), formatDuration(total))

	barWidth := min(40, width-lipgloss.Width(times))
	if barWidth < 5 {
		return strings.TrimSpace(times)
	}

	ratio := 0.0
	if durationMS > 0 {
		ratio = float64(progressMS) / float64(durationMS)
	}
	filled := min(barWidth, int(ratio*float64(barWidth)))

	bar := m.theme.ProgressBar.Render(strings.Repeat("█", filled)) +
		strings.Repeat("░", barWidth-filled)

	return bar + times
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
