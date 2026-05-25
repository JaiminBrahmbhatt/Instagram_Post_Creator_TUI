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

var ollamaHTTPClient = &http.Client{Timeout: 180 * time.Second}

// OllamaClient calls a local Ollama server for vision-based photo grouping.
type OllamaClient struct {
	BaseURL string
	Model   string
}

func (c *OllamaClient) DisplayName() string {
	return "Ollama / " + c.Model
}

// GroupPhotos sends images to the Ollama /api/chat endpoint and returns
// suggested groupings for Instagram carousel posts.
func (c *OllamaClient) GroupPhotos(photoPaths []string) ([]PhotoGroup, error) {
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

	var images []string
	var validBatch []string
	for _, path := range batch {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		images = append(images, base64.StdEncoding.EncodeToString(data))
		validBatch = append(validBatch, path)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("could not read any image files")
	}

	prompt := buildGroupingPrompt(len(images))

	reqBody, err := json.Marshal(map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
				"images":  images,
			},
		},
		"stream": false,
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("model not found — run: ollama pull %s", c.Model)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var apiResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if apiResp.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", apiResp.Error)
	}

	return parseGroupingResponse(apiResp.Message.Content, validBatch)
}
