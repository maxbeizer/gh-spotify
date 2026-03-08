package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// CallbackResult holds the result of the OAuth callback.
type CallbackResult struct {
	Code  string
	State string
	Err   error
}

// DefaultCallbackPort is the fixed port for the OAuth callback server.
// Users register http://127.0.0.1:8818/callback in the Spotify dashboard.
const DefaultCallbackPort = "8818"

// RedirectURI is the full redirect URI to register in the Spotify dashboard.
const RedirectURI = "http://127.0.0.1:" + DefaultCallbackPort + "/callback"

// StartCallbackServer starts a localhost HTTP server on the fixed callback port
// and waits for the OAuth callback. It returns the listener address and a channel
// that will receive the callback result.
func StartCallbackServer(ctx context.Context) (string, <-chan CallbackResult, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+DefaultCallbackPort)
	if err != nil {
		return "", nil, fmt.Errorf("starting callback server: %w", err)
	}

	resultCh := make(chan CallbackResult, 1)
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			resultCh <- CallbackResult{Err: fmt.Errorf("spotify auth error: %s — %s", errParam, q.Get("error_description"))}
			fmt.Fprint(w, "<html><body><h2>Authentication failed.</h2><p>You can close this tab.</p></body></html>")
			return
		}

		resultCh <- CallbackResult{
			Code:  q.Get("code"),
			State: q.Get("state"),
		}
		fmt.Fprint(w, "<html><body><h2>Success!</h2><p>You can close this tab and return to your terminal.</p></body></html>")
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultCh <- CallbackResult{Err: fmt.Errorf("callback server error: %w", err)}
		}
	}()

	// Shut down when context is cancelled or result is received.
	go func() {
		select {
		case <-ctx.Done():
			server.Shutdown(context.Background())
		}
	}()

	addr := fmt.Sprintf("http://%s", listener.Addr().String())
	return addr, resultCh, nil
}
