package api

import (
	"context"
	"net/http"
	"testing"
)

func TestStartTunnelEmptyToken(t *testing.T) {
	_, err := StartTunnel(context.Background(), "", http.NewServeMux())
	if err == nil {
		t.Error("Expected error for empty auth token, got nil")
	}
}

func TestGetTunnelURL(t *testing.T) {
	// Should be empty initially
	url := GetTunnelURL()
	if url != "" {
		t.Errorf("Expected empty URL initially, got %s", url)
	}
}

func TestStopTunnelIdempotent(t *testing.T) {
	// Ensure it's stopped
	StopTunnel()

	// Call again
	err := StopTunnel()
	if err != nil {
		t.Errorf("StopTunnel should be safe to call multiple times, got: %v", err)
	}
}
