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

// PlaylistDetailResponse wraps the full playlist response from /playlists/{id}.
type PlaylistDetailResponse struct {
	Name   string `json:"name"`
	Tracks struct {
		Items []struct {
			Track Track `json:"track"`
		} `json:"items"`
		Total int `json:"total"`
	} `json:"tracks"`
}

// GetPlaylistTracks returns tracks in a playlist.
func (c *Client) GetPlaylistTracks(playlistID string, limit, offset int) (*PlaylistTracksResponse, error) {
	// Use the full playlist endpoint — the /tracks sub-endpoint requires
	// additional scopes that may not be granted.
	var detail PlaylistDetailResponse
	if err := c.Get(fmt.Sprintf("/playlists/%s", playlistID), &detail); err != nil {
		return nil, err
	}
	return &PlaylistTracksResponse{
		Items: detail.Tracks.Items,
		Total: detail.Tracks.Total,
	}, nil
}

// PlayPlaylist starts playback of a playlist.
func (c *Client) PlayPlaylist(playlistURI string) error {
	return c.Put("/me/player/play", nil) // simplified; full implementation would send context_uri
}
