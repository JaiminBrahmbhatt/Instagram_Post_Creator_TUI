package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	instagram "github.com/JaiminBrahmbhatt/insta-go-sdk"
)

type Client struct {
	AccessToken string
	IGID        string
	SDK         *instagram.SDK
}

func NewClient(accessToken, igID string) *Client {
	config := instagram.Config{
		AccessToken: accessToken,
		IGUserID:    igID,
	}
	sdk, _ := instagram.NewSDK(config)

	return &Client{
		AccessToken: accessToken,
		IGID:        igID,
		SDK:         sdk,
	}
}

func (c *Client) GetPublishingLimit() (*PublishingLimit, error) {
	url := fmt.Sprintf("https://graph.instagram.com/%s/%s/content_publishing_limit?fields=config,quota_usage&access_token=%s", APIVersion, c.IGID, c.AccessToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Using a simple client for this single request
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
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
