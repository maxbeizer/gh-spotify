# gh-spotify

A [GitHub CLI](https://cli.github.com/) extension for controlling Spotify from the terminal.

## Installation

```bash
gh extension install maxbeizer/gh-spotify
```

## Setup

1. Create a Spotify app at [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard)
2. Check **Web API** under "Which API/SDKs are you planning to use?"
3. Set the redirect URI to `http://127.0.0.1:8818/callback`
4. Run `gh spotify auth` and enter your Client ID

## Commands

| Command | Description |
|---------|-------------|
| `gh spotify auth` | Authenticate with Spotify |
| `gh spotify now` | Show currently playing track |
| `gh spotify play` | Resume playback |
| `gh spotify pause` | Pause playback |
| `gh spotify next` | Skip to next track |
| `gh spotify prev` | Go to previous track |
| `gh spotify search <query>` | Search tracks, albums, artists |
| `gh spotify playlist` | Browse your playlists |
| `gh spotify queue [uri]` | View or add to queue |
| `gh spotify devices` | List playback devices |
| `gh spotify device <name>` | Switch playback device |
| `gh spotify volume <0-100>` | Set volume |
| `gh spotify shuffle [on\|off]` | Toggle shuffle |
| `gh spotify repeat [off\|track\|context]` | Set repeat mode |
| `gh spotify art` | ASCII album art |

## Requirements

- [GitHub CLI](https://cli.github.com/) 2.0+
- Spotify Premium account (required for playback control)
- A Spotify Developer app (free to create)

## License

MIT
