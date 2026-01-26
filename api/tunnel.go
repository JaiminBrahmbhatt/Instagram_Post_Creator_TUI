package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.ngrok.com/ngrok/v2"
)

var (
	tunnelURL string
	tunnelMu  sync.RWMutex
	listener  ngrok.EndpointListener
)

// StartTunnel starts an ngrok tunnel using the v2 API and serves the provided handler.
func StartTunnel(ctx context.Context, authToken string, handler http.Handler) (string, error) {
	// Fast check with read lock
	tunnelMu.RLock()
	if tunnelURL != "" {
		url := tunnelURL
		tunnelMu.RUnlock()
		return url, nil
	}
	tunnelMu.RUnlock()

	if authToken == "" {
		return "", fmt.Errorf("ngrok auth token is required")
	}

	log.Println("[Ngrok] Setting up tunnel...")
	
	// Set the token in environment as expected by the ngrok-go DefaultAgent
	os.Setenv("NGROK_AUTHTOKEN", authToken)

	// Create a timeout context for the connection process
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Println("[Ngrok] Calling ngrok.Listen...")
	
	opts := []ngrok.EndpointOption{}
	if domain := GetNgrokDomain(); domain != "" {
		log.Printf("[Ngrok] Using static domain: %s", domain)
		opts = append(opts, ngrok.WithURL(domain))
	}

	l, err := ngrok.Listen(connectCtx, opts...)
	if err != nil {
		log.Printf("[Ngrok] Listen failed: %v", err)
		return "", fmt.Errorf("failed to start ngrok tunnel: %w", err)
	}
	log.Printf("[Ngrok] Tunnel established at %s", l.URL())

	// Update state
	tunnelMu.Lock()
	listener = l
	tunnelURL = l.URL().String()
	tunnelMu.Unlock()

	go func() {
		log.Println("[Ngrok] Starting HTTP server on tunnel...")
		if err := http.Serve(l, handler); err != nil && err != http.ErrServerClosed {
			log.Printf("[Ngrok] Tunnel server error: %v", err)
		}
	}()

	return tunnelURL, nil
}

// StopTunnel closes the active ngrok tunnel.
func StopTunnel() error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if listener != nil {
		log.Println("[Ngrok] Stopping tunnel...")
		err := listener.Close()
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
