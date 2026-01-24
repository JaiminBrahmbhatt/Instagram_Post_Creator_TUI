package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.ngrok.com/ngrok/v2"
)

var (
	tunnelURL string
	tunnelMu  sync.RWMutex
	listener  ngrok.EndpointListener
	agent     ngrok.Agent
)

// StartTunnel starts an ngrok tunnel using the v2 API and serves the provided handler.
func StartTunnel(ctx context.Context, authToken string, handler http.Handler) (string, error) {
	// Fast check with read lock
	tunnelMu.RLock()
	if tunnelURL != "" {
		defer tunnelMu.RUnlock()
		return tunnelURL, nil
	}
	tunnelMu.RUnlock()

	if authToken == "" {
		return "", fmt.Errorf("ngrok auth token is required")
	}

	// Connect to ngrok (Slow Network Operation - No Lock held here)
	a, err := ngrok.NewAgent(ngrok.WithAuthtoken(authToken))
	if err != nil {
		return "", fmt.Errorf("failed to create ngrok agent: %w", err)
	}

	if err := a.Connect(ctx); err != nil {
		return "", fmt.Errorf("failed to connect ngrok agent: %w", err)
	}

	l, err := a.Listen(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start ngrok tunnel: %w", err)
	}

	// Update state (Fast Operation - Lock held here)
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	agent = a
	listener = l
	tunnelURL = l.URL().String()

	go func() {
		if err := http.Serve(l, handler); err != nil {
			// In a real app, we'd handle this error better
		}
	}()

	return tunnelURL, nil
}

// StopTunnel closes the active ngrok tunnel and disconnects the agent.
func StopTunnel() error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if agent != nil {
		err := agent.Disconnect()
		agent = nil
		listener = nil
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