// mcp-server exposes the Instagram Post Creator as an MCP tool server.
// It communicates over stdin/stdout using the MCP JSON-RPC protocol,
// allowing Claude Code and other MCP clients to list photos, group them
// with Claude AI, and create posts without launching the TUI.
//
// Usage:
//
//	./post-creator-mcp
//
// Add to Claude Code via .claude/mcp.json:
//
//	{"mcpServers":{"instagram":{"command":"./post-creator-mcp"}}}
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/mcp"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	f, err := os.OpenFile("mcp-debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	a, err := app.New(app.Config{DBPath: "post_creator.db"})
	if err != nil {
		log.Fatalf("initialising app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Skip tunnel — MCP server does not publish posts directly.
	a.StartInfra(ctx, app.Config{SkipTunnel: true})
	defer a.Shutdown(context.Background())

	mcp.NewServer(a).Run()
}
