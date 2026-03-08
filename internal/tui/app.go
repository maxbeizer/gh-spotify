package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

// Tab represents a top-level navigation panel.
type Tab int

const (
	TabNowPlaying Tab = iota
	TabPlaylists
	TabSearch
	TabDevices
)

func (t Tab) String() string {
	switch t {
	case TabNowPlaying:
		return "Now Playing"
	case TabPlaylists:
		return "Playlists"
	case TabSearch:
		return "Search"
	case TabDevices:
		return "Devices"
	default:
		return ""
	}
}

var allTabs = []Tab{TabNowPlaying, TabPlaylists, TabSearch, TabDevices}

// App is the top-level bubbletea model.
type App struct {
	activeTab  Tab
	width      int
	height     int
	theme      Theme
	showHelp   bool
	nowPlaying NowPlayingModel
	controls   ControlsModel
	panels     PanelSet
}

// NewApp creates a new App model.
func NewApp() App {
	theme := DefaultTheme()

	// Attempt to create a Spotify client; nil is okay if not authenticated.
	var client *spotify.Client
	c, err := spotify.NewClient()
	if err == nil {
		client = c
	}

	return App{
		activeTab:  TabNowPlaying,
		theme:      theme,
		nowPlaying: NewNowPlayingModel(client, theme),
		controls:   NewControlsModel(client, theme),
		panels:     NewPanelSet(client, theme),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.nowPlaying.Init(),
		a.controls.Init(),
		a.panels.Playlists.Init(),
		a.panels.Search.Init(),
		a.panels.Devices.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentHeight := a.height - TabBarHeight - 2
		if contentHeight < 1 {
			contentHeight = 1
		}
		a.panels.SetSize(msg.Width, contentHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			prev := a.activeTab
			a.activeTab = Tab((int(a.activeTab) + 1) % len(allTabs))
			a.handleTabSwitch(prev, a.activeTab)
		case "shift+tab":
			prev := a.activeTab
			a.activeTab = Tab((int(a.activeTab) - 1 + len(allTabs)) % len(allTabs))
			a.handleTabSwitch(prev, a.activeTab)
		case "?":
			a.showHelp = !a.showHelp
		}
	}

	// Always route to now-playing and controls so they receive state updates.
	var cmd tea.Cmd
	a.nowPlaying, cmd = a.nowPlaying.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	a.controls, cmd = a.controls.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Always route non-key messages to all panels (so async data loads land).
	// Only route key messages to the active panel.
	_, isKey := msg.(tea.KeyMsg)

	if !isKey || a.activeTab == TabPlaylists {
		a.panels.Playlists, cmd = a.panels.Playlists.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if !isKey || a.activeTab == TabSearch {
		a.panels.Search, cmd = a.panels.Search.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if !isKey || a.activeTab == TabDevices {
		a.panels.Devices, cmd = a.panels.Devices.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Tab bar
	sections = append(sections, a.renderTabBar())

	// Active panel content
	sections = append(sections, a.renderPanel())

	// Help or status bar
	if a.showHelp {
		sections = append(sections, a.renderHelp())
	} else {
		sections = append(sections, a.renderStatusBar())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a App) renderTabBar() string {
	var tabs []string
	for _, t := range allTabs {
		if t == a.activeTab {
			tabs = append(tabs, a.theme.ActiveTab.Render(t.String()))
		} else {
			tabs = append(tabs, a.theme.InactiveTab.Render(t.String()))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
}

const helpHeight = 12 // lines used by the help overlay

func (a App) renderPanel() string {
	statusBarHeight := 1
	reserveForHelp := statusBarHeight
	if a.showHelp {
		reserveForHelp = helpHeight
	}
	contentHeight := a.height - TabBarHeight - reserveForHelp
	if contentHeight < 1 {
		contentHeight = 1
	}

	var content string
	switch a.activeTab {
	case TabNowPlaying:
		content = a.nowPlaying.View() + "\n" + a.controls.View()
	case TabPlaylists:
		content = a.panels.Playlists.View()
	case TabSearch:
		content = a.panels.Search.View()
	case TabDevices:
		content = a.panels.Devices.View()
	}

	return lipgloss.NewStyle().
		Width(a.width).
		Height(contentHeight).
		Render(content)
}

func (a App) renderStatusBar() string {
	status := fmt.Sprintf(" %s", a.activeTab.String())
	help := "spc:play/pause n:next p:prev ?:help "
	available := a.width - lipgloss.Width(status)
	if available < lipgloss.Width(help) {
		help = "?:help "
	}
	gap := max(0, a.width-lipgloss.Width(status)-lipgloss.Width(help))
	line := status + strings.Repeat(" ", gap) + help
	// Truncate to terminal width to prevent wrapping.
	if lipgloss.Width(line) > a.width {
		line = line[:a.width]
	}
	return a.theme.StatusBar.Render(line)
}

func (a App) renderHelp() string {
	help := []string{
		"",
		"  Navigation                     Playback (any tab)",
		"  ───────────                    ──────────────────",
		"  tab/shift+tab  switch panels   space  play/pause",
		"  ?              toggle help      n      next track",
		"  q/ctrl+c       quit             p      prev track",
		"                                  +/-    volume",
		"                                  s      shuffle",
		"                                  r      repeat",
		"",
	}
	return a.theme.HelpBar.Width(a.width).Render(strings.Join(help, "\n"))
}

// handleTabSwitch manages focus transitions between tabs.
func (a *App) handleTabSwitch(from, to Tab) {
	if from == TabSearch {
		a.panels.Search.Blur()
	}
	if to == TabSearch {
		a.panels.Search.Focus()
	}
}
