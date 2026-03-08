package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

const feedbackDuration = 2 * time.Second

// controlFeedbackMsg clears the transient feedback message.
type controlFeedbackMsg struct{}

// controlActionMsg carries the result of a control action (e.g. play/pause).
type controlActionMsg struct {
	feedback string
	err      error
}

// ControlsModel renders playback controls and handles keybindings.
type ControlsModel struct {
	client   *spotify.Client
	state    *spotify.PlayerState
	feedback string
	theme    Theme
}

// NewControlsModel creates a ControlsModel. client may be nil if not authenticated.
func NewControlsModel(client *spotify.Client, theme Theme) ControlsModel {
	return ControlsModel{
		client: client,
		theme:  theme,
	}
}

func (m ControlsModel) Init() tea.Cmd {
	return nil
}

func (m ControlsModel) Update(msg tea.Msg) (ControlsModel, tea.Cmd) {
	switch msg := msg.(type) {

	case playerStateMsg:
		// Keep a local copy of state for rendering and toggle logic.
		if msg.err == nil {
			m.state = msg.state
		}

	case controlActionMsg:
		if msg.err != nil {
			m.feedback = fmt.Sprintf("⚠ %s", msg.err)
		} else {
			m.feedback = msg.feedback
		}
		return m, tea.Tick(feedbackDuration, func(_ time.Time) tea.Msg {
			return controlFeedbackMsg{}
		})

	case controlFeedbackMsg:
		m.feedback = ""

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ControlsModel) handleKey(msg tea.KeyMsg) (ControlsModel, tea.Cmd) {
	if m.client == nil {
		return m, nil
	}

	switch msg.String() {
	case " ":
		return m, m.togglePlayPause()
	case "n":
		return m, m.doAction("⏭ Next", func() error { return m.client.Next() })
	case "p":
		return m, m.doAction("⏮ Previous", func() error { return m.client.Previous() })
	case "+", "=":
		return m, m.adjustVolume(5)
	case "-":
		return m, m.adjustVolume(-5)
	case "s":
		return m, m.toggleShuffle()
	case "r":
		return m, m.cycleRepeat()
	}

	return m, nil
}

func (m ControlsModel) View() string {
	if m.client == nil {
		return ""
	}

	playIcon := "▶"
	shuffleIcon := "🔀"
	repeatIcon := "🔁"
	volBar := ""

	if m.state != nil {
		if m.state.IsPlaying {
			playIcon = "⏸"
		}
		if m.state.ShuffleState {
			shuffleIcon = "🔀✓"
		}
		switch m.state.RepeatState {
		case "track":
			repeatIcon = "🔂"
		case "context":
			repeatIcon = "🔁✓"
		}
		volBar = renderVolume(m.state.Device.VolumePercent)
	}

	controls := fmt.Sprintf("  ⏮  %s  ⏭   %s  %s  %s", playIcon, shuffleIcon, repeatIcon, volBar)

	if m.feedback != "" {
		controls += "  " + m.feedback
	}

	return m.theme.ControlsBar.Render(controls)
}

func renderVolume(pct int) string {
	filled := pct / 10
	if filled > 10 {
		filled = 10
	}
	empty := 10 - filled
	return fmt.Sprintf("Vol: %s%s %d%%",
		strings.Repeat("█", filled),
		strings.Repeat("░", empty),
		pct,
	)
}

// --- action helpers ---

func (m ControlsModel) togglePlayPause() tea.Cmd {
	return func() tea.Msg {
		if m.state != nil && m.state.IsPlaying {
			err := m.client.Pause()
			return controlActionMsg{feedback: "⏸ Paused", err: err}
		}
		err := m.client.Play()
		return controlActionMsg{feedback: "▶ Playing", err: err}
	}
}

func (m ControlsModel) doAction(label string, fn func() error) tea.Cmd {
	return func() tea.Msg {
		err := fn()
		return controlActionMsg{feedback: label, err: err}
	}
}

func (m ControlsModel) adjustVolume(delta int) tea.Cmd {
	return func() tea.Msg {
		vol := 50
		if m.state != nil {
			vol = m.state.Device.VolumePercent
		}
		vol += delta
		if vol < 0 {
			vol = 0
		}
		if vol > 100 {
			vol = 100
		}
		err := m.client.SetVolume(vol)
		return controlActionMsg{feedback: fmt.Sprintf("🔊 Volume %d%%", vol), err: err}
	}
}

func (m ControlsModel) toggleShuffle() tea.Cmd {
	return func() tea.Msg {
		newState := true
		if m.state != nil {
			newState = !m.state.ShuffleState
		}
		err := m.client.SetShuffle(newState)
		label := "🔀 Shuffle on"
		if !newState {
			label = "🔀 Shuffle off"
		}
		return controlActionMsg{feedback: label, err: err}
	}
}

func (m ControlsModel) cycleRepeat() tea.Cmd {
	return func() tea.Msg {
		current := "off"
		if m.state != nil {
			current = m.state.RepeatState
		}
		next := "off"
		switch current {
		case "off":
			next = "context"
		case "context":
			next = "track"
		case "track":
			next = "off"
		}
		err := m.client.SetRepeat(next)
		return controlActionMsg{feedback: fmt.Sprintf("🔁 Repeat: %s", next), err: err}
	}
}
