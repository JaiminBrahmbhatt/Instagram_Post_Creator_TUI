package api

import (
	"log"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	serviceName        = "instagram-post-creator"
	ngrokTokenKey      = "NGROK_AUTH_TOKEN"
	ngrokDomainKey     = "NGROK_DOMAIN"
	anthropicKeyName   = "ANTHROPIC_API_KEY"
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

// GetNgrokToken retrieves the Ngrok auth token from the keyring.
func GetNgrokToken() string {
	return GetCredential(ngrokTokenKey)
}

// SaveNgrokToken saves the Ngrok auth token to the keyring.
func SaveNgrokToken(token string) error {
	return SetCredential(ngrokTokenKey, token)
}

// GetNgrokDomain retrieves the Ngrok static domain from the keyring.
func GetNgrokDomain() string {
	return GetCredential(ngrokDomainKey)
}

// SaveNgrokDomain saves the Ngrok static domain to the keyring.
func SaveNgrokDomain(domain string) error {
	return SetCredential(ngrokDomainKey, domain)
}

// GetAnthropicKey retrieves the Anthropic API key from keyring or environment.
func GetAnthropicKey() string {
	return GetCredential(anthropicKeyName)
}

// SaveAnthropicKey persists the Anthropic API key to the system keyring.
func SaveAnthropicKey(key string) error {
	return SetCredential(anthropicKeyName, key)
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
