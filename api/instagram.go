package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MediaType string

const (
	MediaTypeCarousel MediaType = "CAROUSEL"
	MediaTypeImage    MediaType = "IMAGE"
	MediaTypeReels    MediaType = "REELS"
	MediaTypeStories  MediaType = "STORIES"
	MediaTypeVideo    MediaType = "VIDEO"
)

// SupportedExtensions defines the file extensions allowed for media upload.
// This is the source of truth for the entire application.
// Note: Keep extensions lowercase.
var SupportedExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov"}

type ContainerStatus string

const (
	ContainerStatusExpired    ContainerStatus = "EXPIRED"
	ContainerStatusError      ContainerStatus = "ERROR"
	ContainerStatusFinished   ContainerStatus = "FINISHED"
	ContainerStatusInProgress ContainerStatus = "IN_PROGRESS"
	ContainerStatusPublished  ContainerStatus = "PUBLISHED"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	IGID        string
}

// Request Types

type MediaCreationRequest struct {
	AccessToken    string    `json:"access_token"`
	Caption        string    `json:"caption,omitempty"`
	MediaType      MediaType `json:"media_type,omitempty"`
	ImageURL       string    `json:"image_url,omitempty"`
	VideoURL       string    `json:"video_url,omitempty"`
	IsCarouselItem bool      `json:"is_carousel_item,omitempty"`
	Children       string    `json:"children,omitempty"` // Comma-separated list of IDs
}

type MediaPublishRequest struct {
	AccessToken string `json:"access_token"`
	CreationID  string `json:"creation_id"`
}

// Response Types

type ContainerResponse struct {
	ID string `json:"id"`
}

type ContainerStatusResponse struct {
	ID         string          `json:"id"`
	StatusCode ContainerStatus `json:"status_code"` // FINISHED, IN_PROGRESS, ERROR, EXPIRED, PUBLISHED
}

type LimitResponse struct {
	Data []PublishingLimit `json:"data"`
}

type PublishingLimit struct {
	Config struct {
		QuotaDuration int `json:"quota_duration"`
		QuotaTotal    int `json:"quota_total"`
	} `json:"config"`
	QuotaUsage int `json:"quota_usage"`
}

func NewClient(accessToken, igID string) *Client {
	return &Client{
		AccessToken: accessToken,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		IGID:        igID,
	}
}

func (c *Client) CreateCarouselContainer(caption string, children []string) (string, error) {
	childrenStr := ""
	for i, child := range children {
		if i > 0 {
			childrenStr += ","
		}
		childrenStr += child
	}

	req := MediaCreationRequest{
		AccessToken: c.AccessToken,
		Caption:     caption,
		Children:    childrenStr,
		MediaType:   MediaTypeCarousel,
	}
	return c.makePostRequest("media", req)
}

func (c *Client) CreateMediaContainer(mediaURL, caption string, mediaType MediaType, isCarouselItem bool) (string, error) {
	req := MediaCreationRequest{
		AccessToken:    c.AccessToken,
		Caption:        caption,
		MediaType:      mediaType,
		IsCarouselItem: isCarouselItem,
	}

	if mediaType == MediaTypeVideo {
		req.VideoURL = mediaURL
	} else {
		req.ImageURL = mediaURL
	}

	return c.makePostRequest("media", req)
}

func (c *Client) GetContainerStatus(containerID string) (ContainerStatus, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v20.0/%s?fields=status_code&access_token=%s", containerID, c.AccessToken)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var res ContainerStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.StatusCode, nil
}

func (c *Client) GetPublishingLimit() (*PublishingLimit, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v20.0/%s/content_publishing_limit?fields=config,quota_usage&access_token=%s", c.IGID, c.AccessToken)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var res LimitResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, fmt.Errorf("no limit data returned")
	}

	return &res.Data[0], nil
}

func (c *Client) PublishContainer(containerID string) (string, error) {
	req := MediaPublishRequest{
		AccessToken: c.AccessToken,
		CreationID:  containerID,
	}
	return c.makePostRequest("media_publish", req)
}

func (c *Client) WaitForContainer(containerID string) error {
	// Poll for up to 5 minutes
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for container %s to be ready", containerID)
		case <-ticker.C:
			status, err := c.GetContainerStatus(containerID)
			if err != nil {
				return err
			}
			if status == ContainerStatusFinished {
				return nil
			}
			if status == ContainerStatusError || status == ContainerStatusExpired {
				return fmt.Errorf("container %s failed with status: %s", containerID, status)
			}
			// IN_PROGRESS, continue waiting
		}
	}
}

// makePostRequest handles the common logic for Instagram Graph API POST requests
func (c *Client) makePostRequest(endpoint string, payload any) (string, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v20.0/%s/%s", c.IGID, endpoint)
	log.Printf("API Request: POST %s", url)

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode into a map first to check for error field
	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if errObj, ok := raw["error"].(map[string]interface{}); ok {
		msg := "API Error"
		if m, ok := errObj["message"].(string); ok {
			msg = m
		}
		return "", fmt.Errorf("%s (code: %v)", msg, errObj["code"])
	}

	id, ok := raw["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("API request failed (no ID returned in response: %v)", raw)
	}

	return id, nil
}
