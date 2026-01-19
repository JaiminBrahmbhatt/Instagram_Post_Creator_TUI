package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// startTunnel starts a cloudflared tunnel on the specified port.
// It returns the public URL, a cleanup function, and an error if it fails.
func startTunnel(port int) (string, func(), error) {
	// Create a context that we can cancel to kill the process
	ctx, cancel := context.WithCancel(context.Background())
	
	// Construct command: cloudflared tunnel --url http://localhost:PORT
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))

	// cloudflared outputs the URL to stderr, so we pipe it
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return "", nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return "", nil, fmt.Errorf("failed to start cloudflared: %w", err)
	}

	// Channel to receive the detected URL
	urlChan := make(chan string)
	
	// Regex to extract the URL (e.g., https://foo-bar.trycloudflare.com)
	urlRegex := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)

	// Scan stderr in a goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Log tunnel output to our debug log for troubleshooting
			log.Printf("[Cloudflared] %s", line)

			// Check for URL
			if match := urlRegex.FindString(line); match != "" {
				// We found it! Send it and keep scanning (logging)
				// Use a non-blocking send in case we found it multiple times or stopped listening
				select {
				case urlChan <- match:
				default:
				}
			}
		}
	}()

	// Wait for the URL to appear or timeout
	select {
	case url := <-urlChan:
		log.Printf("Tunnel established at: %s", url)
		
		// Ensure trailing slash for simple concatenation
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}

		// Cleanup function
		cleanup := func() {
			log.Println("Shutting down cloudflare tunnel...")
			cancel() // Kills the process via context
			_ = cmd.Wait()
		}

		return url, cleanup, nil

	case <-time.After(15 * time.Second):
		// Timeout
		cancel()
		_ = cmd.Wait()
		return "", nil, fmt.Errorf("timed out waiting for tunnel URL")
	}
}
