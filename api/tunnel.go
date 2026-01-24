package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

var (
	tunnelURL string
	tunnelMu  sync.RWMutex
	session   ngrok.Session
)

// StartTunnel starts an ngrok tunnel and serves the provided handler.
func StartTunnel(ctx context.Context, authToken string, handler http.Handler) (string, error) {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if tunnelURL != "" {
		return tunnelURL, nil
	}

	if authToken == "" {
		return "", fmt.Errorf("ngrok auth token is required")
	}

	sess, err := ngrok.Connect(ctx,
		ngrok.WithAuthtoken(authToken),
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to ngrok: %w", err)
	}
	session = sess

	tun, err := sess.Listen(ctx,
		config.HTTPEndpoint(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to start ngrok tunnel: %w", err)
	}

	tunnelURL = tun.URL()

	go func() {
		if err := http.Serve(tun, handler); err != nil {
			// In a real app, we'd handle this error better
		}
	}()

	return tunnelURL, nil
}

// StopTunnel closes the active ngrok tunnel.
func StopTunnel() error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if session != nil {
		err := session.Close()
		session = nil
		tunnelURL = ""
		return err
	}
	return nil
}

// GetTunnelURL returns the current tunnel URL.
func GetTunnelURL() string {
	tunnelMu.RLock()
	defer tunnelMu.RUnlock()
	return tunnelURL
}
