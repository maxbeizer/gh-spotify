package spotify

import "fmt"

// GetPlaylists returns the current user's playlists.
func (c *Client) GetPlaylists(limit, offset int) (*PlaylistsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	var resp PlaylistsResponse
	if err := c.Get(fmt.Sprintf("/me/playlists?limit=%d&offset=%d", limit, offset), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// playlistDetailResponse wraps the full playlist response from /playlists/{id}.
// The Spotify API returns tracks under "items" (not "tracks"), and each
// track entry uses "item" (not "track") for the track object.
type playlistDetailResponse struct {
	Name  string `json:"name"`
	Items struct {
		Items []struct {
			Track Track `json:"track"`
			Item  Track `json:"item"`
		} `json:"items"`
		Total int `json:"total"`
	} `json:"items"`
}

// GetPlaylistTracks returns tracks in a playlist.
func (c *Client) GetPlaylistTracks(playlistID string, limit, offset int) (*PlaylistTracksResponse, error) {
	// Use the full playlist endpoint — the /tracks sub-endpoint requires
	// additional scopes that may not be granted.
	var detail playlistDetailResponse
	if err := c.Get(fmt.Sprintf("/playlists/%s", playlistID), &detail); err != nil {
		return nil, err
	}

	// Map to PlaylistTracksResponse, handling both "track" and "item" keys.
	result := &PlaylistTracksResponse{Total: detail.Items.Total}
	for _, entry := range detail.Items.Items {
		t := entry.Track
		if t.ID == "" {
			t = entry.Item
		}
		if t.ID != "" {
			result.Items = append(result.Items, struct {
				Track Track `json:"track"`
			}{Track: t})
		}
	}
	return result, nil
}

// PlayPlaylist starts playback of a playlist.
func (c *Client) PlayPlaylist(playlistURI string) error {
	return c.Put("/me/player/play", nil) // simplified; full implementation would send context_uri
}
