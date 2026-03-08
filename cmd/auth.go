package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/maxbeizer/gh-spotify/internal/auth"
	"github.com/maxbeizer/gh-spotify/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Spotify",
	Long: `Run the OAuth2 PKCE authentication flow to connect your Spotify account.

Before running this command, create a free Spotify app:

  1. Go to https://developer.spotify.com/dashboard
  2. Click "Create app"
  3. Enter any name and description
  4. Under "Redirect URIs", enter exactly:
       http://127.0.0.1:8818/callback
     (Use 127.0.0.1, not "localhost" — Spotify rejects localhost as insecure)
  5. Under "Which API/SDKs are you planning to use?", check "Web API"
  6. Agree to the terms and click "Save"
  7. Copy your Client ID from the app's settings page

Then run 'gh spotify auth' and paste your Client ID when prompted.

Note: Spotify Premium is required for playback control commands.`,
	RunE: runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config or start fresh.
	cfg, _ := config.Load()

	if cfg.ClientID == "" {
		fmt.Println("To authenticate, you need a Spotify Client ID.")
		fmt.Println()
		fmt.Println("If you haven't already, create a Spotify app:")
		fmt.Println("  1. Go to https://developer.spotify.com/dashboard")
		fmt.Println("  2. Click 'Create app'")
		fmt.Printf("  3. Set redirect URI to: %s\n", auth.RedirectURI)
		fmt.Println("     (Use 127.0.0.1, NOT localhost — Spotify rejects localhost)")
		fmt.Println("  4. Check 'Web API' under API/SDKs")
		fmt.Println("  5. Save, then copy your Client ID from settings")
		fmt.Println()
		fmt.Print("Enter your Spotify Client ID: ")
		clientID, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading client ID: %w", err)
		}
		cfg.ClientID = strings.TrimSpace(clientID)
	} else {
		fmt.Printf("Using existing Client ID: %s...%s\n", cfg.ClientID[:4], cfg.ClientID[len(cfg.ClientID)-4:])
	}

	// Generate PKCE values.
	verifier, err := auth.GenerateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generating code verifier: %w", err)
	}
	challenge := auth.CodeChallenge(verifier)

	state, err := auth.GenerateState()
	if err != nil {
		return fmt.Errorf("generating state: %w", err)
	}

	// Start callback server.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	addr, resultCh, err := auth.StartCallbackServer(ctx)
	if err != nil {
		return err
	}
	_ = addr // server is running; we use the fixed RedirectURI

	redirectURI := auth.RedirectURI
	cfg.RedirectURI = redirectURI

	// Build auth URL and open browser.
	authURL := auth.AuthorizeURL(cfg.ClientID, redirectURI, challenge, state)
	fmt.Printf("\nOpening browser for Spotify authorization...\n")
	fmt.Printf("If it doesn't open, visit:\n  %s\n\n", authURL)
	openBrowser(authURL)

	// Wait for callback.
	fmt.Println("Waiting for authorization...")
	select {
	case result := <-resultCh:
		cancel()
		if result.Err != nil {
			return result.Err
		}
		if result.State != state {
			return fmt.Errorf("state mismatch — possible CSRF attack")
		}

		// Exchange code for token.
		fmt.Println("Exchanging authorization code for token...")
		token, err := auth.ExchangeCode(cfg.ClientID, result.Code, verifier, redirectURI)
		if err != nil {
			return err
		}

		// Save config and token.
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		if err := auth.SaveToken(token); err != nil {
			return fmt.Errorf("saving token: %w", err)
		}

		fmt.Println("✓ Authentication successful! You can now use gh spotify.")
		return nil

	case <-ctx.Done():
		return fmt.Errorf("authorization timed out (5 minutes)")
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
