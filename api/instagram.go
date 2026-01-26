package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	params := map[string]string{
		"caption":    caption,
		"children":   childrenStr,
		"media_type": string(MediaTypeCarousel),
	}
	return c.makePostRequest("media", params)
}

func (c *Client) CreateMediaContainer(mediaURL, caption string, mediaType MediaType, isCarouselItem bool) (string, error) {
	params := map[string]string{
		"caption": caption,
	}

	// Only include media_type for videos (VIDEO, REELS, STORIES)
	// For images, media_type should NOT be sent
	if mediaType == MediaTypeVideo {
		params["media_type"] = string(mediaType)
		params["video_url"] = mediaURL
	} else {
		// For images, just send image_url without media_type
		params["image_url"] = mediaURL
	}

	// Only include is_carousel_item if it's true
	// Instagram API rejects the request if is_carousel_item=false is sent
	if isCarouselItem {
		params["is_carousel_item"] = "true"
	}

	return c.makePostRequest("media", params)
}

func (c *Client) GetContainerStatus(containerID string) (ContainerStatus, error) {
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s?fields=status_code&access_token=%s", APIVersion, containerID, c.AccessToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

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
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s/content_publishing_limit?fields=config,quota_usage&access_token=%s", APIVersion, c.IGID, c.AccessToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

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
	params := map[string]string{
		"creation_id": containerID,
	}
	return c.makePostRequest("media_publish", params)
}

func (c *Client) WaitForContainer(containerID string) error {
	// Poll for up to 5 minutes
	ticker := time.NewTicker(5 * time.Second)
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

// makePostRequest handles the common logic for Instagram Graph API POST requests using query parameters
func (c *Client) makePostRequest(endpoint string, params map[string]string) (string, error) {
	baseURL := fmt.Sprintf("https://graph.instagram.com/%s/%s/%s", APIVersion, c.IGID, endpoint)

	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Add(k, v)
		}
	}
	// Always add access token if not present
	if values.Get("access_token") == "" {
		values.Add("access_token", c.AccessToken)
	}

	fullURL := baseURL + "?" + values.Encode()
	log.Printf("API Request: POST %s", baseURL)
	log.Printf("DEBUG - Parameters map: %+v", params)
	log.Printf("DEBUG - URL values: %s", values.Encode())

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode response
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
