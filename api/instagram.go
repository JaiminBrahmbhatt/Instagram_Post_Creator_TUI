package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	IGID        string
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
		Caption:   caption,
		Children:  childrenStr,
		MediaType: MediaTypeCarousel,
	}
	return c.makePostRequest("media", req)
}

func (c *Client) CreateMediaContainer(mediaURL, caption string, mediaType MediaType, isCarouselItem bool) (string, error) {
	req := MediaCreationRequest{
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
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s?fields=status_code", APIVersion, containerID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
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
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s/content_publishing_limit?fields=config,quota_usage", APIVersion, c.IGID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
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
		CreationID: containerID,
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
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s/%s", APIVersion, c.IGID, endpoint)
	log.Printf("API Request: POST %s", url)

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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
