package api

import (
	"log"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "insta-auto-post"
)

// GetCredential retrieves a credential from environment variables first,
// then falls back to the system keyring.
func GetCredential(key string) string {
	if val := os.Getenv(key); val != "" {
		go syncToKeyring(key, val)
		return val
	}

	secret, err := keyring.Get(serviceName, key)
	if err != nil {
		return ""
	}

	return secret
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
