package api

import (
	"log"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "instagram-post-creator"
)

// GetCredential retrieves a credential from the system keyring first,
// then falls back to environment variables.
func GetCredential(key string) string {
	// Try keyring first
	if secret, err := keyring.Get(serviceName, key); err == nil && secret != "" {
		return secret
	}

	// Fallback to environment/dotenv
	if val := os.Getenv(key); val != "" {
		go syncToKeyring(key, val)
		return val
	}

	return ""
}

// SetCredential saves a credential to the system keyring.
func SetCredential(key, value string) error {
	return keyring.Set(serviceName, key, value)
}

func syncToKeyring(key, value string) {
	// Try to see if it exists already
	existing, _ := keyring.Get(serviceName, key)
	if existing != value {
		if err := keyring.Set(serviceName, key, value); err != nil {
			log.Printf("Warning: failed to sync %s to keyring: %v", key, err)
		}
	}
}
