package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

// devicesLoadedMsg is sent when devices finish loading.
type devicesLoadedMsg struct {
	devices []spotify.Device
	err     error
}

// transferCompleteMsg is sent after playback transfer completes.
type transferCompleteMsg struct {
	err error
}

// DevicesModel is the device selector panel.
type DevicesModel struct {
	client  *spotify.Client
	theme   Theme
	width   int
	height  int
	devices []spotify.Device
	cursor  int
	loading bool
	err     error
	status  string // transient status message
}

// NewDevicesModel creates a new device selector panel.
func NewDevicesModel(client *spotify.Client, theme Theme) DevicesModel {
	return DevicesModel{
		client: client,
		theme:  theme,
	}
}

// Init starts loading devices.
func (m DevicesModel) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	m.loading = true
	return m.loadDevices
}

func (m DevicesModel) loadDevices() tea.Msg {
	devices, err := m.client.GetDevices()
	if err != nil {
		return devicesLoadedMsg{err: err}
	}
	return devicesLoadedMsg{devices: devices}
}

// SetSize updates the panel dimensions.
func (m *DevicesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages for the devices panel.
func (m DevicesModel) Update(msg tea.Msg) (DevicesModel, tea.Cmd) {
	if m.client == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case devicesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.devices = msg.devices
		if m.cursor >= len(m.devices) {
			m.cursor = max(len(m.devices)-1, 0)
		}

	case transferCompleteMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Transfer failed: %s", msg.err)
		} else {
			m.status = "Playback transferred"
		}
		// Refresh device list after transfer.
		return m, m.loadDevices

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.devices) {
				d := m.devices[m.cursor]
				if d.IsActive {
					m.status = "Already active"
					return m, nil
				}
				m.status = "Transferring..."
				return m, m.transferPlayback(d.ID)
			}
		case "r":
			m.loading = true
			m.status = ""
			return m, m.loadDevices
		}
	}

	return m, nil
}

func (m DevicesModel) transferPlayback(deviceID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.TransferPlayback(deviceID)
		return transferCompleteMsg{err: err}
	}
}

// View renders the devices panel.
func (m DevicesModel) View() string {
	if m.client == nil {
		return m.theme.UnselectedItem.Render("Not authenticated — run gh spotify auth")
	}
	if m.loading {
		return m.theme.UnselectedItem.Render("Loading devices...")
	}
	if m.err != nil {
		return m.theme.UnselectedItem.Render(fmt.Sprintf("Error: %s", m.err))
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		PaddingLeft(2).
		Render("Devices")

	s := header + "\n\n"

	if len(m.devices) == 0 {
		s += m.theme.UnselectedItem.Render("  No devices available")
		s += "\n\n" + m.theme.HelpBar.Render("r: refresh")
		return s
	}

	for i, d := range m.devices {
		active := "  "
		if d.IsActive {
			active = "▶ "
		}

		line := fmt.Sprintf("%s%s  ·  %s  ·  vol %d%%", active, d.Name, d.Type, d.VolumePercent)

		if i == m.cursor {
			s += m.theme.SelectedItem.Render("▸ "+line) + "\n"
		} else {
			s += m.theme.UnselectedItem.Render("  "+line) + "\n"
		}
	}

	s += "\n" + m.theme.HelpBar.Render("enter: transfer playback  r: refresh")

	if m.status != "" {
		s += "\n" + m.theme.StatusBar.Render(m.status)
	}

	return s
}
