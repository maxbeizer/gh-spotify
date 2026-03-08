package spotify

import (
	"fmt"
	"net/url"
	"strings"
)

// Search searches Spotify for tracks, albums, artists, and playlists.
func (c *Client) Search(query string, types []string, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{
		"q":     {query},
		"type":  {strings.Join(types, ",")},
		"limit": {fmt.Sprintf("%d", limit)},
	}

	var result SearchResult
	if err := c.Get("/search?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
