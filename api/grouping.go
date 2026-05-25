package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Backend and model constants.
const (
	BackendClaude = "claude"
	BackendOllama = "ollama"

	DefaultOllamaURL = "http://localhost:11434"

	// MaxPhotosPerBatch is the maximum number of images sent per analysis call.
	MaxPhotosPerBatch = 20
	// CarouselMaxPhotos is Instagram's hard limit for carousel posts.
	CarouselMaxPhotos = 10
	// CarouselMinPhotos is the minimum photos required for a carousel group.
	CarouselMinPhotos = 2

	// Claude model IDs
	ModelClaudeOpus   = "claude-opus-4-7"
	ModelClaudeSonnet = "claude-sonnet-4-6"
	ModelClaudeHaiku  = "claude-haiku-4-5-20251001"

	// Ollama vision model IDs
	ModelGemma3    = "gemma3:4b"
	ModelLlava     = "llava"
	ModelLlavaPhi3 = "llava-phi3"
	ModelMoondream = "moondream"
)

// PhotoGroup is a cluster of image paths considered thematically similar.
type PhotoGroup struct {
	Name   string
	Reason string
	Photos []string // absolute file paths
}

// GroupingBackend abstracts over photo clustering backends (Claude, Ollama, …).
type GroupingBackend interface {
	GroupPhotos(photoPaths []string) ([]PhotoGroup, error)
	DisplayName() string
}

// NewGroupingBackend creates a GroupingBackend from plain settings strings.
// Passing empty strings for model or ollamaURL uses sensible defaults.
func NewGroupingBackend(backend, model, ollamaURL string) GroupingBackend {
	switch backend {
	case BackendOllama:
		if model == "" {
			model = ModelGemma3
		}
		if ollamaURL == "" {
			ollamaURL = DefaultOllamaURL
		}
		return &OllamaClient{BaseURL: ollamaURL, Model: model}
	default: // "claude" or unset
		if model == "" {
			model = ModelClaudeOpus
		}
		return &ClaudeClient{APIKey: GetAnthropicKey(), Model: model}
	}
}

// buildGroupingPrompt returns the shared grouping prompt for n images.
func buildGroupingPrompt(n int) string {
	return fmt.Sprintf(
		"I've shared %d photos. Group them into Instagram carousel posts based on "+
			"visual similarity, location, event, or story cohesion.\n\n"+
			"Rules:\n"+
			"- Each group must have %d–%d photos (Instagram carousel hard limits)\n"+
			"- A photo can appear in only one group\n"+
			"- Skip photos that don't fit cleanly with any others\n"+
			"- Never put more than %d photos in a single group\n\n"+
			"Respond with ONLY valid JSON, no explanation before or after:\n"+
			`{"groups":[{"name":"short title","reason":"why these go together","indices":[0,1]}]}`,
		n, CarouselMinPhotos, CarouselMaxPhotos, CarouselMaxPhotos,
	)
}

// parseGroupingResponse extracts PhotoGroups from a JSON response string,
// mapping indices back to the original batch of file paths.
func parseGroupingResponse(text string, batch []string) ([]PhotoGroup, error) {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end < start {
		return nil, fmt.Errorf("no JSON found in response")
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
