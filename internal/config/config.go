package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const appName = "gh-spotify"

// Config holds the application configuration.
type Config struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// Dir returns the configuration directory path (~/.config/gh-spotify).
func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(configDir, appName), nil
}

// Load reads the config from disk. Returns a zero Config if the file doesn't exist.
func Load() (Config, error) {
	dir, err := Dir()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
