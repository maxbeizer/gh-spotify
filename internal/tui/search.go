package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

// searchResultsMsg is sent when search results arrive.
type searchResultsMsg struct {
	result *spotify.SearchResult
	err    error
}

// searchResultItem is a flattened search result entry for cursor navigation.
type searchResultItem struct {
	label    string
	category string // "Tracks", "Artists", "Albums", "Playlists"
	uri      string // Spotify URI for playable items
}

// SearchModel is the search panel.
type SearchModel struct {
	client  *spotify.Client
	theme   Theme
	width   int
	height  int
	input   textinput.Model
	results []searchResultItem
	cursor  int
	loading bool
	err     error
	queried bool // true after first search
}

// NewSearchModel creates a new search panel.
func NewSearchModel(client *spotify.Client, theme Theme) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search Spotify..."
	ti.CharLimit = 128
	ti.Width = 38

	return SearchModel{
		client: client,
		theme:  theme,
		input:  ti,
	}
}

// Init focuses the text input.
func (m SearchModel) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return textinput.Blink
}

// SetSize updates the panel dimensions.
func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Focus gives focus to the search input.
func (m *SearchModel) Focus() {
	m.input.Focus()
}

// Blur removes focus from the search input.
func (m *SearchModel) Blur() {
	m.input.Blur()
}

// InputFocused reports whether the search text input has focus.
func (m SearchModel) InputFocused() bool {
	return m.input.Focused()
}

func (m SearchModel) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.client.Search(query, []string{"track", "artist", "album", "playlist"}, 10)
		if err != nil {
			return searchResultsMsg{err: err}
		}
		return searchResultsMsg{result: result}
	}
}

// Update handles messages for the search panel.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if m.client == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case searchResultsMsg:
		m.loading = false
		m.queried = true
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.results = flattenResults(msg.result)
		// Start cursor on first playable item, not a section header.
		m.cursor = 0
		if len(m.results) > 0 && m.results[0].label == "" {
			m.cursor = nextNonHeader(m.results, -1, 1)
		}

	case tea.KeyMsg:
		if m.input.Focused() {
			switch msg.String() {
			case "enter":
				q := m.input.Value()
				if q == "" {
					return m, nil
				}
				m.loading = true
				m.input.Blur()
				return m, m.doSearch(q)
			case "esc":
				if len(m.results) > 0 {
					m.input.Blur()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// Navigate results when input is blurred.
		switch msg.String() {
		case "up", "k":
			m.cursor = nextNonHeader(m.results, m.cursor, -1)
		case "down", "j":
			m.cursor = nextNonHeader(m.results, m.cursor, 1)
		case "enter":
			// Queue the selected item as next.
			if m.cursor >= 0 && m.cursor < len(m.results) && m.results[m.cursor].uri != "" {
				uri := m.results[m.cursor].uri
				return m, func() tea.Msg {
					err := m.client.AddToQueue(uri)
					if err == nil {
						m.client.Play()
					}
					return controlActionMsg{feedback: "➕ Queued next", err: err}
				}
			}
		case "P":
			// Play the selected item immediately (Shift+P).
			if m.cursor >= 0 && m.cursor < len(m.results) && m.results[m.cursor].uri != "" {
				uri := m.results[m.cursor].uri
				return m, func() tea.Msg {
					err := m.client.PlayTrack(uri)
					return controlActionMsg{feedback: "▶ Playing now", err: err}
				}
			}
		case "/":
			m.input.Focus()
			return m, textinput.Blink
		}
	}

	// Pass through to textinput when focused.
	if m.input.Focused() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the search panel.
func (m SearchModel) View() string {
	if m.client == nil {
		return m.theme.UnselectedItem.Render("Not authenticated — run gh spotify auth")
	}

	s := m.theme.SearchInput.Render(m.input.View()) + "\n\n"

	if m.loading {
		s += m.theme.UnselectedItem.Render("Searching...")
		return s
	}
	if m.err != nil {
		s += m.theme.UnselectedItem.Render(fmt.Sprintf("Error: %s", m.err))
		return s
	}
	if !m.queried {
		s += m.theme.HelpBar.Render("Type a query and press Enter to search")
		return s
	}
	if len(m.results) == 0 {
		s += m.theme.UnselectedItem.Render("No results found")
		return s
	}

	maxVisible := m.height - 6
	start, end := scrollWindow(len(m.results), m.cursor, maxVisible)

	for i := start; i < end; i++ {
		item := m.results[i]
		if item.category != "" && item.label == "" {
			// Section header.
			header := lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText).
				PaddingLeft(2).
				Render(item.category)
			s += "\n" + header + "\n"
			continue
		}
		if i == m.cursor {
			s += m.theme.SelectedItem.Render("▸ "+item.label) + "\n"
		} else {
			s += m.theme.UnselectedItem.Render("  "+item.label) + "\n"
		}
	}

	s += "\n" + m.theme.HelpBar.Render("/: search  ↑↓: navigate  enter: queue next  P: play now")
	return s
}

// flattenResults converts a SearchResult into a flat list with section headers.
func flattenResults(r *spotify.SearchResult) []searchResultItem {
	var items []searchResultItem

	if len(r.Tracks.Items) > 0 {
		items = append(items, searchResultItem{category: fmt.Sprintf("Tracks (%d)", r.Tracks.Total)})
		for _, t := range r.Tracks.Items {
			items = append(items, searchResultItem{
				label:    fmt.Sprintf("%s — %s", t.Name, t.ArtistNames()),
				category: "Tracks",
				uri:      t.URI,
			})
		}
	}
	if len(r.Artists.Items) > 0 {
		items = append(items, searchResultItem{category: fmt.Sprintf("Artists (%d)", r.Artists.Total)})
		for _, a := range r.Artists.Items {
			items = append(items, searchResultItem{
				label:    a.Name,
				category: "Artists",
				uri:      a.URI,
			})
		}
	}
	if len(r.Albums.Items) > 0 {
		items = append(items, searchResultItem{category: fmt.Sprintf("Albums (%d)", r.Albums.Total)})
		for _, a := range r.Albums.Items {
			items = append(items, searchResultItem{
				label:    a.Name,
				category: "Albums",
				uri:      a.URI,
			})
		}
	}
	if len(r.Playlists.Items) > 0 {
		var playlistItems []searchResultItem
		for _, p := range r.Playlists.Items {
			if p.Name == "" || p.TrackCount() == 0 {
				continue
			}
			playlistItems = append(playlistItems, searchResultItem{
				label:    fmt.Sprintf("%s (%d tracks)", p.Name, p.TrackCount()),
				category: "Playlists",
				uri:      p.URI,
			})
		}
		if len(playlistItems) > 0 {
			items = append(items, searchResultItem{category: fmt.Sprintf("Playlists (%d)", len(playlistItems))})
			items = append(items, playlistItems...)
		}
	}

	return items
}

// nextNonHeader moves the cursor skipping section header rows.
func nextNonHeader(items []searchResultItem, cur, dir int) int {
	next := cur + dir
	for next >= 0 && next < len(items) {
		if items[next].label != "" {
			return next
		}
		next += dir
	}
	return cur
}

// scrollWindow returns the visible window [start, end) for a list.
func scrollWindow(total, cursor, maxVisible int) (int, int) {
	if maxVisible <= 0 || maxVisible >= total {
		return 0, total
	}
	start := cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
}
