package api

import (
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
	return c.SDK.GetPublishingLimit()
}
