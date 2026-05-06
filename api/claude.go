package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	anthropicModel  = "claude-opus-4-7"
	// MaxPhotosPerBatch is the maximum number of images sent to Claude per analysis call.
	MaxPhotosPerBatch = 20
	// CarouselMaxPhotos is Instagram's hard limit for carousel posts.
	CarouselMaxPhotos = 10
	// CarouselMinPhotos is the minimum photos required for a carousel group.
	CarouselMinPhotos = 2
)

var anthropicHTTPClient = &http.Client{Timeout: 120 * time.Second}

// PhotoGroup is a cluster of image paths that Claude considers thematically similar.
type PhotoGroup struct {
	Name   string
	Reason string
	Photos []string // absolute file paths
}

// ClaudeClient calls the Anthropic API for vision-based photo analysis.
type ClaudeClient struct {
	APIKey string
}

// NewClaudeClient creates a client using the stored Anthropic API key.
func NewClaudeClient() *ClaudeClient {
	return &ClaudeClient{APIKey: GetAnthropicKey()}
}

// GroupPhotos sends up to maxPhotosPerBatch images to Claude and returns
// suggested groupings for Instagram carousel posts.
func (c *ClaudeClient) GroupPhotos(photoPaths []string) ([]PhotoGroup, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key not set — add ANTHROPIC_API_KEY to your environment or .env file")
	}

	// Filter to supported image types only (no video)
	var imagePaths []string
	for _, p := range photoPaths {
		switch strings.ToLower(filepath.Ext(p)) {
		case ".jpg", ".jpeg", ".png", ".gif":
			imagePaths = append(imagePaths, p)
		}
	}
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no image files found to group")
	}

	batch := imagePaths
	if len(batch) > MaxPhotosPerBatch {
		batch = batch[:MaxPhotosPerBatch]
	}

	// Build content blocks: one image block per file, then a text prompt.
	type imageSource struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	}
	type contentBlock struct {
		Type   string       `json:"type"`
		Text   string       `json:"text,omitempty"`
		Source *imageSource `json:"source,omitempty"`
	}

	var content []contentBlock
	for _, path := range batch {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content = append(content, contentBlock{
			Type: "image",
			Source: &imageSource{
				Type:      "base64",
				MediaType: extToMediaType(filepath.Ext(path)),
				Data:      base64.StdEncoding.EncodeToString(data),
			},
		})
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("could not read any image files")
	}

	prompt := fmt.Sprintf(
		"I've shared %d photos. Group them into Instagram carousel posts based on "+
			"visual similarity, location, event, or story cohesion.\n\n"+
			"Rules:\n"+
			"- Each group must have %d–%d photos (Instagram carousel hard limits)\n"+
			"- A photo can appear in only one group\n"+
			"- Skip photos that don't fit cleanly with any others\n"+
			"- Never put more than %d photos in a single group\n\n"+
			"Respond with ONLY valid JSON, no explanation before or after:\n"+
			`{"groups":[{"name":"short title","reason":"why these go together","indices":[0,1]}]}`,
		len(content),
		CarouselMinPhotos, CarouselMaxPhotos, CarouselMaxPhotos,
	)
	content = append(content, contentBlock{Type: "text", Text: prompt})

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":      anthropicModel,
		"max_tokens": 1024,
		"messages": []map[string]interface{}{
			{"role": "user", "content": content},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := anthropicHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var apiResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}
	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Extract JSON from the response text (Claude may add a tiny preamble).
	text := apiResp.Content[0].Text
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end < start {
		return nil, fmt.Errorf("no JSON found in Claude response")
	}

	var result struct {
		Groups []struct {
			Name    string `json:"name"`
			Reason  string `json:"reason"`
			Indices []int  `json:"indices"`
		} `json:"groups"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &result); err != nil {
		return nil, fmt.Errorf("parsing grouping JSON: %w", err)
	}

	var groups []PhotoGroup
	for _, g := range result.Groups {
		var paths []string
		for _, idx := range g.Indices {
			if idx >= 0 && idx < len(batch) {
				paths = append(paths, batch[idx])
			}
		}
		// Enforce Instagram carousel limits strictly.
		if len(paths) > CarouselMaxPhotos {
			paths = paths[:CarouselMaxPhotos]
		}
		if len(paths) >= CarouselMinPhotos {
			groups = append(groups, PhotoGroup{
				Name:   g.Name,
				Reason: g.Reason,
				Photos: paths,
			})
		}
	}
	return groups, nil
}

func extToMediaType(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}
