package spotify

// Image represents a Spotify image.
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// Artist represents a simplified Spotify artist.
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// Album represents a simplified Spotify album.
type Album struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	URI    string  `json:"uri"`
	Images []Image `json:"images"`
}

// Track represents a Spotify track.
type Track struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	URI      string   `json:"uri"`
	Artists  []Artist `json:"artists"`
	Album    Album    `json:"album"`
	Duration int      `json:"duration_ms"`
}

// Device represents a Spotify playback device.
type Device struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	IsRestricted  bool   `json:"is_restricted"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	VolumePercent int    `json:"volume_percent"`
}

// PlayerState represents the current playback state.
type PlayerState struct {
	Device       Device `json:"device"`
	ShuffleState bool   `json:"shuffle_state"`
	RepeatState  string `json:"repeat_state"`
	IsPlaying    bool   `json:"is_playing"`
	Item         *Track `json:"item"`
	ProgressMS   int    `json:"progress_ms"`
}

// DevicesResponse wraps the devices list response.
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// Playlist represents a Spotify playlist.
type Playlist struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	URI         string  `json:"uri"`
	Images      []Image `json:"images"`
	Tracks      struct {
		Total int `json:"total"`
	} `json:"tracks"`
}

// PlaylistsResponse wraps the playlists list response.
type PlaylistsResponse struct {
	Items []Playlist `json:"items"`
	Total int        `json:"total"`
}

// PlaylistTracksResponse wraps the playlist tracks response.
type PlaylistTracksResponse struct {
	Items []struct {
		Track Track `json:"track"`
	} `json:"items"`
	Total int `json:"total"`
}

// SearchResult represents Spotify search results.
type SearchResult struct {
	Tracks struct {
		Items []Track `json:"items"`
		Total int     `json:"total"`
	} `json:"tracks"`
	Albums struct {
		Items []Album `json:"items"`
		Total int     `json:"total"`
	} `json:"albums"`
	Artists struct {
		Items []Artist `json:"items"`
		Total int      `json:"total"`
	} `json:"artists"`
	Playlists struct {
		Items []Playlist `json:"items"`
		Total int        `json:"total"`
	} `json:"playlists"`
}

// ArtistNames returns a comma-separated string of artist names for a track.
func (t Track) ArtistNames() string {
	names := make([]string, len(t.Artists))
	for i, a := range t.Artists {
		names[i] = a.Name
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
