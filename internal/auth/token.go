package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxbeizer/gh-spotify/internal/config"
)

const tokenEndpoint = "https://accounts.spotify.com/api/token"

// Token represents stored OAuth tokens.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired reports whether the access token has expired (with 60s buffer).
func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-60 * time.Second))
}

// ExchangeCode exchanges an authorization code for tokens.
func ExchangeCode(clientID, code, codeVerifier, redirectURI string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {codeVerifier},
	}

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token exchange failed (%d): %s — %s", resp.StatusCode, errResp.Error, errResp.Description)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

// Refresh uses the refresh token to get a new access token.
func Refresh(clientID string, token *Token) (*Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
		"client_id":     {clientID},
	}

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	var newToken Token
	if err := json.NewDecoder(resp.Body).Decode(&newToken); err != nil {
		return nil, fmt.Errorf("decoding refresh response: %w", err)
	}

	// Spotify may not return a new refresh token; keep the old one.
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = token.RefreshToken
	}
	newToken.ExpiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
	return &newToken, nil
}

// tokenPath returns the path to the token file.
func tokenPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

// LoadToken reads the stored token from disk.
func LoadToken() (*Token, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("not authenticated — run 'gh spotify auth' first")
	}
	if err != nil {
		return nil, fmt.Errorf("reading token: %w", err)
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &token, nil
}

// SaveToken writes the token to disk.
func SaveToken(token *Token) error {
	path, err := tokenPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

// AuthorizeURL builds the Spotify authorization URL for PKCE.
func AuthorizeURL(clientID, redirectURI, codeChallenge, state string) string {
	scopes := []string{
		"user-read-playback-state",
		"user-modify-playback-state",
		"user-read-currently-playing",
		"playlist-read-private",
		"user-library-read",
	}

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"scope":                 {strings.Join(scopes, " ")},
		"redirect_uri":          {redirectURI},
		"code_challenge_method": {"S256"},
		"code_challenge":        {codeChallenge},
		"state":                 {state},
	}
	return "https://accounts.spotify.com/authorize?" + params.Encode()
}
