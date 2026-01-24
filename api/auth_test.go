package api

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestNgrokTokenStorage(t *testing.T) {
	// Initialize mock keyring
	keyring.MockInit()

	expectedToken := "test-ngrok-token-123"

	// Test SaveNgrokToken
	err := SaveNgrokToken(expectedToken)
	if err != nil {
		t.Fatalf("SaveNgrokToken failed: %v", err)
	}

	// Test GetNgrokToken
	token := GetNgrokToken()
	if token != expectedToken {
		t.Errorf("GetNgrokToken returned %s, expected %s", token, expectedToken)
	}
}

func TestNgrokTokenEmpty(t *testing.T) {
	keyring.MockInit()

	// Ensure it returns empty string if not set
	token := GetNgrokToken()
	if token != "" {
		t.Errorf("Expected empty token, got %s", token)
	}
}
