package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const graphHost = "https://graph.facebook.com"

// GraphClient calls Instagram Graph API for container creation with optional params (alt_text, location_id).
type GraphClient struct {
	AccessToken string
	IGID        string
	HTTPClient  *http.Client
}

// NewGraphClient returns a client using the same credentials as the main API client.
func NewGraphClient(accessToken, igID string) *GraphClient {
	return &GraphClient{
		AccessToken: accessToken,
		IGID:        igID,
		HTTPClient:  &http.Client{Timeout: 60 * time.Second},
	}
}

// CreateImageContainerOptions holds optional parameters for image container creation.
type CreateImageContainerOptions struct {
	Caption    string
	AltText    string
	LocationID string
	UserTags   []string // Instagram usernames to tag
}

// CreateImageContainer creates a single image media container with optional caption, alt_text, location_id, user_tags.
// Returns the container ID on success. Validates inputs per implementation_plan.md (location numeric, username format, alt_text length).
func (c *GraphClient) CreateImageContainer(ctx context.Context, imageURL string, opts CreateImageContainerOptions) (string, error) {
	if err := ValidateLocationID(opts.LocationID); err != nil {
		return "", err
	}
	if err := ValidateUserTags(opts.UserTags); err != nil {
		return "", err
	}
	if err := ValidateAltText(opts.AltText, true); err != nil {
		return "", err
	}
	u := fmt.Sprintf("%s/%s/%s/media", graphHost, APIVersion, c.IGID)
	body := map[string]interface{}{"image_url": imageURL}
	if opts.Caption != "" {
		body["caption"] = opts.Caption
	}
	if opts.AltText != "" {
		body["alt_text"] = opts.AltText
	}
	if opts.LocationID != "" {
		body["location_id"] = opts.LocationID
	}
	if len(opts.UserTags) > 0 {
		tags := make([]map[string]string, len(opts.UserTags))
		for i, u := range opts.UserTags {
			tags[i] = map[string]string{"username": u}
		}
		body["user_tags"] = tags
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("instagram api: %s (code %d)", out.Error.Message, out.Error.Code)
	}
	if out.ID == "" {
		return "", fmt.Errorf("instagram api: empty container id")
	}
	return out.ID, nil
}

// ReelContainerOptions holds optional parameters for Reel container creation.
type ReelContainerOptions struct {
	Caption      string
	ShareToFeed  bool
	CoverURL     string
	ThumbOffset  int   // milliseconds; 0 = not set
	Collaborators []string
	AudioName    string
}

// CreateReelContainer creates a Reels media container with optional share_to_feed, cover_url, thumb_offset, collaborators, audio_name (implementation_plan.md §4.1).
func (c *GraphClient) CreateReelContainer(ctx context.Context, videoURL string, opts ReelContainerOptions) (string, error) {
	if err := ValidateUserTags(opts.Collaborators); err != nil {
		return "", err
	}
	u := fmt.Sprintf("%s/%s/%s/media", graphHost, APIVersion, c.IGID)
	body := map[string]interface{}{
		"media_type": "REELS",
		"video_url":  videoURL,
	}
	if opts.Caption != "" {
		body["caption"] = opts.Caption
	}
	if opts.ShareToFeed {
		body["share_to_feed"] = true
	}
	if opts.CoverURL != "" {
		body["cover_url"] = opts.CoverURL
	}
	if opts.ThumbOffset > 0 {
		body["thumb_offset"] = opts.ThumbOffset
	}
	if len(opts.Collaborators) > 0 {
		body["collaborators"] = opts.Collaborators
	}
	if opts.AudioName != "" {
		body["audio_name"] = opts.AudioName
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("instagram api: %s (code %d)", out.Error.Message, out.Error.Code)
	}
	if out.ID == "" {
		return "", fmt.Errorf("instagram api: empty container id")
	}
	return out.ID, nil
}

// CreateStoryContainer creates a Story media container (image or video). Stories expire after 24 hours (implementation_plan.md §3.2).
func (c *GraphClient) CreateStoryContainer(ctx context.Context, mediaURL string, isVideo bool, caption string, userTags []string) (string, error) {
	if err := ValidateUserTags(userTags); err != nil {
		return "", err
	}
	u := fmt.Sprintf("%s/%s/%s/media", graphHost, APIVersion, c.IGID)
	body := map[string]interface{}{
		"media_type": "STORIES",
	}
	if isVideo {
		body["video_url"] = mediaURL
	} else {
		body["image_url"] = mediaURL
	}
	if caption != "" {
		body["caption"] = caption
	}
	if len(userTags) > 0 {
		tags := make([]map[string]string, len(userTags))
		for i, username := range userTags {
			tags[i] = map[string]string{"username": username}
		}
		body["user_tags"] = tags
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("instagram api: %s (code %d)", out.Error.Message, out.Error.Code)
	}
	if out.ID == "" {
		return "", fmt.Errorf("instagram api: empty container id")
	}
	return out.ID, nil
}

// GetContainerStatus returns the status_code of the container (e.g. IN_PROGRESS, FINISHED, ERROR).
func (c *GraphClient) GetContainerStatus(ctx context.Context, containerID string) (ContainerStatus, error) {
	u := fmt.Sprintf("%s/%s/%s?fields=status_code", graphHost, APIVersion, containerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ContainerStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode status: %w", err)
	}
	return out.StatusCode, nil
}

// PublishContainer publishes the container and returns the Instagram media ID.
func (c *GraphClient) PublishContainer(ctx context.Context, containerID string) (string, error) {
	u := fmt.Sprintf("%s/%s/%s/media_publish", graphHost, APIVersion, c.IGID)
	body := map[string]string{"creation_id": containerID}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode publish response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("instagram publish: %s", out.Error.Message)
	}
	if out.ID == "" {
		return "", fmt.Errorf("instagram publish: empty media id")
	}
	return out.ID, nil
}
