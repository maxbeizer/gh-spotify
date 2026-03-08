# gh-spotify Copilot Instructions

## Project Overview

`gh-spotify` is a GitHub CLI extension that provides a full TUI for controlling Spotify from the terminal. It is written in Go and uses the Spotify Web API with OAuth2 PKCE authentication.

## Architecture

- **CLI framework:** `cobra` via `cmd/` package
- **TUI framework:** `bubbletea` + `lipgloss` + `bubbles` via `internal/tui/`
- **Spotify API client:** `internal/spotify/` — typed HTTP client with auto-refresh
- **Auth:** `internal/auth/` — OAuth2 PKCE flow, token storage/refresh
- **Config:** `internal/config/` — XDG-compliant config at `~/.config/gh-spotify/`

## Conventions

- Follow `gh` CLI conventions: support `--json` and `--jq` flags on commands that output data
- Use `cobra` for all CLI commands
- All Spotify API interactions go through `internal/spotify/client.go`
- Token and config files live in `~/.config/gh-spotify/`
- Error messages should be user-friendly; suggest `gh spotify auth` when auth fails
- Spotify Premium is required for playback control — detect and warn gracefully

## Testing

- Unit tests alongside source files (`*_test.go`)
- Mock the Spotify HTTP client for tests using interfaces
- Test PKCE generation, token refresh logic, and command output formatting

## Dependencies

- `github.com/spf13/cobra` — CLI
- `github.com/charmbracelet/bubbletea` — TUI
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components
- `github.com/cli/go-gh/v2` — gh CLI integration
