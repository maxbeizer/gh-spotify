# gh-spotify

A [GitHub CLI](https://cli.github.com/) extension for controlling Spotify from the terminal.

## Installation

```bash
gh extension install maxbeizer/gh-spotify
```

## Setup

### 1. Create a Spotify Developer App (free)

1. Go to [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard) and log in
2. Click **"Create app"**
3. Fill in any name and description
4. Under **Redirect URIs**, enter exactly:
   ```
   http://127.0.0.1:8818/callback
   ```
   > ⚠️ You must use `127.0.0.1`, **not** `localhost`. Spotify rejects `localhost` as insecure.
5. Under **"Which API/SDKs are you planning to use?"**, check **Web API**
6. Agree to the terms and click **Save**

### 2. Authenticate

```bash
gh spotify auth
```

Paste your **Client ID** (found on your app's settings page) when prompted. The command will open your browser to authorize the app.

> **Note:** Spotify Premium is required for playback control (play, pause, next, etc.). Free-tier accounts can use read-only commands like `now` and `devices`.

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
