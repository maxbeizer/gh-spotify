package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

// PlayContext starts playback of a context (playlist, album, artist URI).
func (c *Client) PlayContext(contextURI string, shuffle bool) error {
	if err := c.SetShuffle(shuffle); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{
		"context_uri": contextURI,
	})
	return c.Put("/me/player/play", bytes.NewReader(body))
}

// GetPlayerState returns the current playback state.
func (c *Client) GetPlayerState() (*PlayerState, error) {
	var state PlayerState
	if err := c.Get("/me/player", &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// GetCurrentlyPlaying returns the currently playing track.
func (c *Client) GetCurrentlyPlaying() (*PlayerState, error) {
	var state PlayerState
	if err := c.Get("/me/player/currently-playing", &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// Play resumes playback on the active device.
func (c *Client) Play() error {
	return c.Put("/me/player/play", nil)
}

// Pause pauses playback.
func (c *Client) Pause() error {
	return c.Put("/me/player/pause", nil)
}

// Next skips to the next track.
func (c *Client) Next() error {
	return c.Post("/me/player/next", nil)
}

// Previous goes back to the previous track.
func (c *Client) Previous() error {
	return c.Post("/me/player/previous", nil)
}

// GetDevices returns available playback devices.
func (c *Client) GetDevices() ([]Device, error) {
	var resp DevicesResponse
	if err := c.Get("/me/player/devices", &resp); err != nil {
		return nil, err
	}
	return resp.Devices, nil
}

// TransferPlayback transfers playback to the given device.
func (c *Client) TransferPlayback(deviceID string) error {
	body, _ := json.Marshal(map[string]any{
		"device_ids": []string{deviceID},
	})
	return c.Put("/me/player", bytes.NewReader(body))
}

// SetVolume sets the playback volume (0-100).
func (c *Client) SetVolume(percent int) error {
	if percent < 0 || percent > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}
	return c.Put(fmt.Sprintf("/me/player/volume?volume_percent=%d", percent), nil)
}

// SetShuffle enables or disables shuffle.
func (c *Client) SetShuffle(state bool) error {
	return c.Put(fmt.Sprintf("/me/player/shuffle?state=%t", state), nil)
}

// SetRepeat sets the repeat mode (off, track, context).
func (c *Client) SetRepeat(state string) error {
	return c.Put(fmt.Sprintf("/me/player/repeat?state=%s", url.QueryEscape(state)), nil)
}

// AddToQueue adds a track URI to the playback queue.
func (c *Client) AddToQueue(uri string) error {
	return c.Post(fmt.Sprintf("/me/player/queue?uri=%s", url.QueryEscape(uri)), nil)
}
