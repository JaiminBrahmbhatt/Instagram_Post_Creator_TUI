package api

import (
	"os"
	"testing"
)

func TestGetPublicURLPrefix(t *testing.T) {
	// Clean up environment
	oldEnv := os.Getenv("PUBLIC_URL_PREFIX")
	defer os.Setenv("PUBLIC_URL_PREFIX", oldEnv)
	
	// Ensure tunnel is stopped
	StopTunnel()

	// 1. Test fallback when nothing is set
	os.Setenv("PUBLIC_URL_PREFIX", "")
	prefix := getPublicURLPrefix()
	if prefix != "https://example.com/" {
		t.Errorf("Expected default example.com, got %s", prefix)
	}

	// 2. Test environment variable
	os.Setenv("PUBLIC_URL_PREFIX", "http://env-test.com/")
	prefix = getPublicURLPrefix()
	if prefix != "http://env-test.com/" {
		t.Errorf("Expected http://env-test.com/, got %s", prefix)
	}
}
