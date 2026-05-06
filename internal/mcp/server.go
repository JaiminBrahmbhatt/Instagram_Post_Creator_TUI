// Package mcp implements a minimal MCP (Model Context Protocol) server
// over stdin/stdout JSON-RPC, exposing Instagram Post Creator functionality
// as tools that Claude Code (and other MCP clients) can call.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
)

// Server handles MCP JSON-RPC messages from stdin and writes responses to stdout.
type Server struct {
	app *app.App
}

// NewServer creates an MCP server wrapping the given app.
func NewServer(a *app.App) *Server {
	return &Server{app: a}
}

// Run reads newline-delimited JSON-RPC requests from stdin until EOF.
func (s *Server) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	enc := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			enc.Encode(errResponse(nil, -32700, "parse error"))
			continue
		}
		result, err := s.dispatch(req)
		if err != nil {
			enc.Encode(errResponse(req.ID, -32603, err.Error()))
		} else {
			enc.Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result})
		}
	}
}

// ─── JSON-RPC types ────────────────────────────────────────────────────────

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func errResponse(id interface{}, code int, msg string) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

// ─── Dispatch ──────────────────────────────────────────────────────────────

func (s *Server) dispatch(req rpcRequest) (interface{}, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize()
	case "notifications/initialized":
		return nil, nil
	case "tools/list":
		return map[string]interface{}{"tools": toolDefs()}, nil
	case "tools/call":
		return s.handleToolCall(req.Params)
	default:
		return nil, fmt.Errorf("method not found: %s", req.Method)
	}
}

func (s *Server) handleInitialize() (interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
		"serverInfo":      map[string]interface{}{"name": "instagram-post-creator", "version": "1.0.0"},
	}, nil
}

// ─── Tool definitions ──────────────────────────────────────────────────────

type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

func toolDefs() []toolDef {
	obj := func(props map[string]interface{}, required ...string) interface{} {
		schema := map[string]interface{}{"type": "object", "properties": props}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	}
	str := func(desc string) interface{} {
		return map[string]interface{}{"type": "string", "description": desc}
	}
	num := func(desc string) interface{} {
		return map[string]interface{}{"type": "number", "description": desc}
	}

	return []toolDef{
		{
			Name:        "list_photos",
			Description: "List image files in the configured photos directory.",
			InputSchema: obj(map[string]interface{}{
				"subdirectory": str("Optional subdirectory relative to the photos root."),
			}),
		},
		{
			Name:        "ai_group_photos",
			Description: "Use Claude vision to cluster photos into suggested Instagram carousel groups (reads ANTHROPIC_API_KEY).",
			InputSchema: obj(map[string]interface{}{}),
		},
		{
			Name:        "create_post",
			Description: "Save an Instagram carousel post as a draft or schedule it for immediate publishing.",
			InputSchema: obj(
				map[string]interface{}{
					"photo_paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Absolute paths to the image files to include.",
					},
					"caption": str("Post caption (max 2200 chars)."),
					"status":  map[string]interface{}{"type": "string", "enum": []string{"draft", "scheduled"}, "description": "draft (default) or scheduled."},
				},
				"photo_paths", "caption",
			),
		},
		{
			Name:        "list_posts",
			Description: "List recent posts and their status.",
			InputSchema: obj(map[string]interface{}{
				"limit":  num("Maximum number of results (default 20)."),
				"status": str("Filter by status: draft, scheduled, publishing, published, failed."),
			}),
		},
	}
}

// ─── Tool call dispatcher ──────────────────────────────────────────────────

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolCall(raw json.RawMessage) (interface{}, error) {
	var p callParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	switch p.Name {
	case "list_photos":
		return s.toolListPhotos(p.Arguments)
	case "ai_group_photos":
		return s.toolAIGroupPhotos()
	case "create_post":
		return s.toolCreatePost(p.Arguments)
	case "list_posts":
		return s.toolListPosts(p.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", p.Name)
	}
}

// ─── Tool implementations ──────────────────────────────────────────────────

func (s *Server) photosDir() (string, error) {
	dir, err := s.app.DB.GetSetting("photos_dir")
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("photos directory not configured — run the TUI and complete setup first")
	}
	return dir, nil
}

func (s *Server) toolListPhotos(raw json.RawMessage) (interface{}, error) {
	var args struct {
		Subdirectory string `json:"subdirectory"`
	}
	json.Unmarshal(raw, &args)

	root, err := s.photosDir()
	if err != nil {
		return nil, err
	}
	dir := root
	if args.Subdirectory != "" {
		dir = filepath.Join(root, args.Subdirectory)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type fileInfo struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Size int64  `json:"size_bytes"`
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !slices.Contains(api.SupportedExtensions, strings.ToLower(filepath.Ext(e.Name()))) {
			continue
		}
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		files = append(files, fileInfo{
			Name: e.Name(),
			Path: filepath.Join(dir, e.Name()),
			Size: size,
		})
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d photos in %s\n\n", len(files), dir))
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("  %s  (%d bytes)\n  path: %s\n\n", f.Name, f.Size, f.Path))
	}
	return textResult(sb.String()), nil
}

func (s *Server) toolAIGroupPhotos() (interface{}, error) {
	root, err := s.photosDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".jpg", ".jpeg", ".png", ".gif":
			paths = append(paths, filepath.Join(root, e.Name()))
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no image files found in %s", root)
	}

	groups, err := api.NewClaudeClient().GroupPhotos(paths)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Claude identified %d carousel groups (analyzed up to %d photos):\n\n",
		len(groups), api.MaxPhotosPerBatch))
	for i, g := range groups {
		sb.WriteString(fmt.Sprintf("Group %d: %s\n", i+1, g.Name))
		sb.WriteString(fmt.Sprintf("  Reason: %s\n", g.Reason))
		sb.WriteString(fmt.Sprintf("  Photos (%d):\n", len(g.Photos)))
		for _, p := range g.Photos {
			sb.WriteString(fmt.Sprintf("    %s\n", p))
		}
		sb.WriteString("\n")
	}
	return textResult(sb.String()), nil
}

func (s *Server) toolCreatePost(raw json.RawMessage) (interface{}, error) {
	var args struct {
		PhotoPaths []string `json:"photo_paths"`
		Caption    string   `json:"caption"`
		Status     string   `json:"status"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	if len(args.PhotoPaths) == 0 {
		return nil, fmt.Errorf("photo_paths must not be empty")
	}
	if len(args.Caption) > 2200 {
		return nil, fmt.Errorf("caption exceeds 2200 character limit")
	}

	status := db.StatusDraft
	scheduleTime := ""
	if args.Status == "scheduled" {
		status = db.StatusScheduled
		scheduleTime = "+0 minutes"
	}

	id, err := s.app.DB.SavePost(args.Caption, args.PhotoPaths, scheduleTime, status)
	if err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Post created: ID %d, status %s, %d photos.", id, status, len(args.PhotoPaths))), nil
}

func (s *Server) toolListPosts(raw json.RawMessage) (interface{}, error) {
	var args struct {
		Limit  int    `json:"limit"`
		Status string `json:"status"`
	}
	json.Unmarshal(raw, &args)
	if args.Limit <= 0 {
		args.Limit = 20
	}

	var (
		posts []db.Post
		err   error
	)
	if args.Status != "" {
		posts, err = s.app.DB.GetPostsByStatus(db.PostStatus(args.Status), args.Limit, 0)
	} else {
		posts, err = s.app.DB.GetPosts(args.Limit, 0)
	}
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	if len(posts) == 0 {
		sb.WriteString("No posts found.\n")
	}
	for _, p := range posts {
		sb.WriteString(fmt.Sprintf("ID %-4d  %-12s  %d photos  %s\n",
			p.ID, p.Status, p.MediaCount, truncate(p.Caption, 60)))
	}
	return textResult(sb.String()), nil
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func textResult(text string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
