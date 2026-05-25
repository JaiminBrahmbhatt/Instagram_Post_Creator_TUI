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

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

var anthropicHTTPClient = &http.Client{Timeout: 120 * time.Second}

// ClaudeClient calls the Anthropic API for vision-based photo analysis.
type ClaudeClient struct {
	APIKey string
	Model  string
}

func (c *ClaudeClient) DisplayName() string {
	return "Claude / " + c.Model
}

// NewClaudeClient creates a client using the stored Anthropic API key and
// defaults to the most capable Claude model.
func NewClaudeClient() *ClaudeClient {
	return &ClaudeClient{APIKey: GetAnthropicKey(), Model: ModelClaudeOpus}
}

// GroupPhotos sends up to MaxPhotosPerBatch images to Claude and returns
// suggested groupings for Instagram carousel posts.
func (c *ClaudeClient) GroupPhotos(photoPaths []string) ([]PhotoGroup, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key not set — add ANTHROPIC_API_KEY to your environment or .env file")
	}

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
	var validBatch []string
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
		validBatch = append(validBatch, path)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("could not read any image files")
	}

	content = append(content, contentBlock{Type: "text", Text: buildGroupingPrompt(len(content))})

	model := c.Model
	if model == "" {
		model = ModelClaudeOpus
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":      model,
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

	return parseGroupingResponse(apiResp.Content[0].Text, validBatch)
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
