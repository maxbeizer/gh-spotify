package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/maxbeizer/gh-spotify/internal/auth"
	"github.com/maxbeizer/gh-spotify/internal/config"
)

const baseURL = "https://api.spotify.com/v1"

// Client is a Spotify Web API client with automatic token refresh.
type Client struct {
	httpClient *http.Client
	token      *auth.Token
	clientID   string
}

// NewClient creates a new Spotify API client. It loads the token and config
// from disk and refreshes the token if expired.
func NewClient() (*Client, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("no client ID configured — run 'gh spotify auth' first")
	}

	c := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
		clientID:   cfg.ClientID,
	}

	if token.IsExpired() {
		if err := c.refreshToken(); err != nil {
			return nil, fmt.Errorf("token expired and refresh failed: %w — run 'gh spotify auth'", err)
		}
	}

	return c, nil
}

// refreshToken refreshes the access token and persists it.
func (c *Client) refreshToken() error {
	newToken, err := auth.Refresh(c.clientID, c.token)
	if err != nil {
		return err
	}
	c.token = newToken
	return auth.SaveToken(newToken)
}

// Do performs an authenticated HTTP request. It retries once on 401 after refreshing the token.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if err := c.refreshToken(); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		return c.httpClient.Do(req)
	}

	return resp, nil
}

// Get performs a GET request to the given Spotify API path and decodes the response.
func (c *Client) Get(path string, v any) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error (%d): %s", resp.StatusCode, string(body))
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

// Put performs a PUT request to the given Spotify API path with no body.
func (c *Client) Put(path string, body io.Reader) error {
	req, err := http.NewRequest("PUT", baseURL+path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Post performs a POST request to the given Spotify API path.
func (c *Client) Post(path string, body io.Reader) error {
	req, err := http.NewRequest("POST", baseURL+path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
