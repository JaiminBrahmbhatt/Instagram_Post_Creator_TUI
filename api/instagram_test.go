package api

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	accessToken := "test_token"
	igID := "test_ig_id"
	client := NewClient(accessToken, igID)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.AccessToken != accessToken {
		t.Errorf("Expected access token %s, got %s", accessToken, client.AccessToken)
	}

	if client.IGID != igID {
		t.Errorf("Expected IGID %s, got %s", igID, client.IGID)
	}

	if client.SDK == nil {
		t.Error("SDK was not initialized")
	}
}
