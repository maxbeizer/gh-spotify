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

You'll need a Spotify Client ID. Get one for free at:
  https://developer.spotify.com/dashboard

When creating your app, set the redirect URI to http://127.0.0.1/callback
(the port will be auto-configured).`,
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
		fmt.Println("Create one at: https://developer.spotify.com/dashboard")
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

	redirectURI := addr + "/callback"
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
