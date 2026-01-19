package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	AccessToken string
	IGID        string
	HTTPClient  *http.Client
}

type ContainerResponse struct {
	ID string `json:"id"`
}

func NewClient(accessToken, igID string) *Client {
	return &Client{
		AccessToken: accessToken,
		IGID:        igID,
		HTTPClient:  &http.Client{},
	}
}

func (c *Client) CreateMediaContainer(imageURL string, isCarouselItem bool) (string, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v24.0/%s/media", c.IGID)

	params := map[string]interface{}{
		"image_url":    imageURL,
		"access_token": c.AccessToken,
	}
	if isCarouselItem {
		params["is_carousel_item"] = true
	}

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

	return res.ID, nil
}

func (c *Client) CreateCarouselContainer(caption string, children []string) (string, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v24.0/%s/media", c.IGID)

	childrenStr := ""
	for i, child := range children {
		childrenStr += child
		if i < len(children)-1 {
			childrenStr += ","
		}
	}

	params := map[string]interface{}{
		"media_type":   "CAROUSEL",
		"caption":      caption,
		"children":     childrenStr,
		"access_token": c.AccessToken,
	}

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

	return res.ID, nil
}

func (c *Client) PublishContainer(containerID string) (string, error) {
	url := fmt.Sprintf("https://graph.instagram.com/v24.0/%s/media_publish", c.IGID)

	params := map[string]string{
		"creation_id":  containerID,
		"access_token": c.AccessToken,
	}

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

	return res.ID, nil
}
