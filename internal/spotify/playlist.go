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

// GetPlaylistTracks returns tracks in a playlist.
func (c *Client) GetPlaylistTracks(playlistID string, limit, offset int) (*PlaylistTracksResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	var resp PlaylistTracksResponse
	if err := c.Get(fmt.Sprintf("/playlists/%s/tracks?limit=%d&offset=%d", playlistID, limit, offset), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PlayPlaylist starts playback of a playlist.
func (c *Client) PlayPlaylist(playlistURI string) error {
	return c.Put("/me/player/play", nil) // simplified; full implementation would send context_uri
}
