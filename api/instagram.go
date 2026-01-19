package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MediaType string

const (
	MediaTypeCarousel MediaType = "CAROUSEL"
	MediaTypeReels    MediaType = "REELS"
	MediaTypeStories  MediaType = "STORIES"
	MediaTypeVideo    MediaType = "VIDEO"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	IGID        string
}

type ContainerResponse struct {
	ID string `json:"id"`
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
		HTTPClient:  &http.Client{},
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

	return c.makePostRequest("media", map[string]any{
		"access_token": c.AccessToken,
		"caption":      caption,
		"children":     childrenStr,
		"media_type":   MediaTypeCarousel,
	})
}

func (c *Client) CreateMediaContainer(imageURL string, isCarouselItem bool) (string, error) {
	params := map[string]any{
		"access_token": c.AccessToken,
		"image_url":    imageURL,
	}
	if isCarouselItem {
		params["is_carousel_item"] = true
	}
	return c.makePostRequest("media", params)
}

func (c *Client) GetPublishingLimit() (*PublishingLimit, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v24.0/%s/content_publishing_limit?fields=config,quota_usage&access_token=%s", c.IGID, c.AccessToken)

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
	return c.makePostRequest("media_publish", map[string]any{
		"access_token": c.AccessToken,
		"creation_id":  containerID,
	})
}

// makePostRequest handles the common logic for Instagram Graph API POST requests
func (c *Client) makePostRequest(endpoint string, params map[string]any) (string, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v24.0/%s/%s", c.IGID, endpoint)

	data, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res ContainerResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.ID == "" {
		return "", fmt.Errorf("API request failed (no ID returned)")
	}

	return res.ID, nil
}
