package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.ngrok.com/ngrok/v2"
)

var (
	tunnelURL string
	tunnelMu  sync.RWMutex
	listener  ngrok.EndpointListener
)

func StartTunnel(ctx context.Context, authToken string, handler http.Handler) (string, error) {
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

	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(authToken))
	if err != nil {
		log.Printf("[Ngrok] Failed to create agent: %v", err)
		return "", fmt.Errorf("failed to create ngrok agent: %w", err)
	}

	opts := []ngrok.EndpointOption{}
	if domain := GetNgrokDomain(); domain != "" {
		log.Printf("[Ngrok] Using static domain: %s", domain)
		opts = append(opts, ngrok.WithURL(domain))
	}

	l, err := agent.Listen(ctx, opts...)
	if err != nil {
		log.Printf("[Ngrok] Listen failed: %v", err)
		return "", fmt.Errorf("failed to start ngrok tunnel: %w", err)
	}
	log.Printf("[Ngrok] Tunnel established at %s", l.URL())

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

func GetTunnelURL() string {
	tunnelMu.RLock()
	defer tunnelMu.RUnlock()
	return tunnelURL
}
