package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// SDK provides a high-level interface for Instagram Content Publishing API
// It handles all the edge cases and quirks of the Instagram Graph API
type SDK struct {
	accessToken string
	igUserID    string
	httpClient  *http.Client
	apiVersion  string
}

// Config holds the configuration for the Instagram SDK
type Config struct {
	AccessToken string        // Instagram Access Token
	IGUserID    string        // Instagram User ID
	APIVersion  string        // API version (default: v19.0)
	HTTPTimeout time.Duration // HTTP client timeout (default: 30s)
}

// NewSDK creates a new Instagram SDK instance
func NewSDK(config Config) (*SDK, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}
	if config.IGUserID == "" {
		return nil, fmt.Errorf("Instagram user ID is required")
	}

	if config.APIVersion == "" {
		config.APIVersion = "v19.0"
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second
	}

	return &SDK{
		accessToken: config.AccessToken,
		igUserID:    config.IGUserID,
		apiVersion:  config.APIVersion,
		httpClient:  &http.Client{Timeout: config.HTTPTimeout},
	}, nil
}

// MediaType represents the type of media being uploaded
type MediaType string

const (
	MediaTypeImage   MediaType = "IMAGE"
	MediaTypeVideo   MediaType = "VIDEO"
	MediaTypeReels   MediaType = "REELS"
	MediaTypeStories MediaType = "STORIES"
)

// Post represents a single Instagram post
type Post struct {
	ImageURL string    // URL to the image (must be publicly accessible)
	VideoURL string    // URL to the video (must be publicly accessible)
	Caption  string    // Post caption
	Type     MediaType // Type of media
}

// CarouselPost represents an Instagram carousel post (2-10 items)
type CarouselPost struct {
	Items   []Post // 2-10 items
	Caption string // Caption for the carousel
}

// ContainerStatus represents the status of a media container
type ContainerStatus string

const (
	StatusExpired    ContainerStatus = "EXPIRED"
	StatusError      ContainerStatus = "ERROR"
	StatusFinished   ContainerStatus = "FINISHED"
	StatusInProgress ContainerStatus = "IN_PROGRESS"
	StatusPublished  ContainerStatus = "PUBLISHED"
)

// PublishSinglePost publishes a single image or video post
// It handles all the complexity of creating containers, waiting for processing, and publishing
func (s *SDK) PublishSinglePost(ctx context.Context, post Post) (string, error) {
	if err := s.validateSinglePost(post); err != nil {
		return "", fmt.Errorf("invalid post: %w", err)
	}

	// Step 1: Create media container
	containerID, err := s.createSingleMediaContainer(post)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Step 2: Wait for Instagram to process the media
	if err := s.waitForContainer(ctx, containerID); err != nil {
		return "", fmt.Errorf("container processing failed: %w", err)
	}

	// Step 3: Publish the container
	mediaID, err := s.publishContainer(containerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish: %w", err)
	}

	return mediaID, nil
}

// PublishCarousel publishes a carousel post with 2-10 items
func (s *SDK) PublishCarousel(ctx context.Context, carousel CarouselPost) (string, error) {
	if err := s.validateCarousel(carousel); err != nil {
		return "", fmt.Errorf("invalid carousel: %w", err)
	}

	// Step 1: Create containers for each item
	var itemIDs []string
	for i, item := range carousel.Items {
		containerID, err := s.createCarouselItemContainer(item)
		if err != nil {
			return "", fmt.Errorf("failed to create container for item %d: %w", i, err)
		}

		// Wait for each item to be processed
		if err := s.waitForContainer(ctx, containerID); err != nil {
			return "", fmt.Errorf("item %d processing failed: %w", i, err)
		}

		itemIDs = append(itemIDs, containerID)
	}

	// Step 2: Create carousel container
	carouselID, err := s.createCarouselContainer(carousel.Caption, itemIDs)
	if err != nil {
		return "", fmt.Errorf("failed to create carousel container: %w", err)
	}

	// Step 3: Wait for carousel processing
	if err := s.waitForContainer(ctx, carouselID); err != nil {
		return "", fmt.Errorf("carousel processing failed: %w", err)
	}

	// Step 4: Publish the carousel
	mediaID, err := s.publishContainer(carouselID)
	if err != nil {
		return "", fmt.Errorf("failed to publish carousel: %w", err)
	}

	return mediaID, nil
}

// createSingleMediaContainer creates a container for a single post
func (s *SDK) createSingleMediaContainer(post Post) (string, error) {
	params := url.Values{}
	params.Add("access_token", s.accessToken)

	if post.Caption != "" {
		params.Add("caption", post.Caption)
	}

	// IMPORTANT: media_type should ONLY be sent for VIDEO, REELS, or STORIES
	// For images, do NOT send media_type parameter
	if post.Type == MediaTypeVideo || post.Type == MediaTypeReels || post.Type == MediaTypeStories {
		params.Add("media_type", string(post.Type))
		params.Add("video_url", post.VideoURL)
	} else {
		// For images, just send image_url without media_type
		params.Add("image_url", post.ImageURL)
	}

	return s.makePostRequest("media", params)
}

// createCarouselItemContainer creates a container for a carousel item
func (s *SDK) createCarouselItemContainer(item Post) (string, error) {
	params := url.Values{}
	params.Add("access_token", s.accessToken)
	params.Add("is_carousel_item", "true")

	if item.Type == MediaTypeVideo {
		params.Add("media_type", string(MediaTypeVideo))
		params.Add("video_url", item.VideoURL)
	} else {
		params.Add("image_url", item.ImageURL)
	}

	return s.makePostRequest("media", params)
}

// createCarouselContainer creates the main carousel container
func (s *SDK) createCarouselContainer(caption string, childrenIDs []string) (string, error) {
	params := url.Values{}
	params.Add("access_token", s.accessToken)
	params.Add("media_type", "CAROUSEL")
	params.Add("children", strings.Join(childrenIDs, ","))

	if caption != "" {
		params.Add("caption", caption)
	}

	return s.makePostRequest("media", params)
}

// publishContainer publishes a media container
func (s *SDK) publishContainer(containerID string) (string, error) {
	params := url.Values{}
	params.Add("access_token", s.accessToken)
	params.Add("creation_id", containerID)

	return s.makePostRequest("media_publish", params)
}

// waitForContainer waits for a container to be processed by Instagram
func (s *SDK) waitForContainer(ctx context.Context, containerID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for container %s", containerID)
		case <-ticker.C:
			status, err := s.getContainerStatus(containerID)
			if err != nil {
				return err
			}

			switch status {
			case StatusFinished:
				return nil
			case StatusError, StatusExpired:
				return fmt.Errorf("container %s failed with status: %s", containerID, status)
			case StatusInProgress:
				// Continue waiting
			}
		}
	}
}

// getContainerStatus gets the status of a media container
func (s *SDK) getContainerStatus(containerID string) (ContainerStatus, error) {
	endpoint := fmt.Sprintf("https://graph.instagram.com/%s/%s", s.apiVersion, containerID)

	params := url.Values{}
	params.Add("fields", "status_code")
	params.Add("access_token", s.accessToken)

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		StatusCode ContainerStatus `json:"status_code"`
		Error      *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}

	return result.StatusCode, nil
}

// makePostRequest makes a POST request to the Instagram API
func (s *SDK) makePostRequest(endpoint string, params url.Values) (string, error) {
	apiURL := fmt.Sprintf("https://graph.instagram.com/%s/%s/%s", s.apiVersion, s.igUserID, endpoint)

	req, err := http.NewRequest("POST", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}

	if result.ID == "" {
		return "", fmt.Errorf("no ID returned from API")
	}

	return result.ID, nil
}

// validateSinglePost validates a single post
func (s *SDK) validateSinglePost(post Post) error {
	if post.Type == MediaTypeVideo || post.Type == MediaTypeReels || post.Type == MediaTypeStories {
		if post.VideoURL == "" {
			return fmt.Errorf("video_url is required for %s", post.Type)
		}
	} else {
		if post.ImageURL == "" {
			return fmt.Errorf("image_url is required for images")
		}
	}

	// Validate URL format
	mediaURL := post.ImageURL
	if mediaURL == "" {
		mediaURL = post.VideoURL
	}

	if !strings.HasPrefix(mediaURL, "http://") && !strings.HasPrefix(mediaURL, "https://") {
		return fmt.Errorf("media URL must be a valid HTTP(S) URL")
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(mediaURL))
	if post.Type == MediaTypeVideo || post.Type == MediaTypeReels {
		if ext != ".mp4" && ext != ".mov" {
			return fmt.Errorf("video must be .mp4 or .mov format")
		}
	} else {
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return fmt.Errorf("image must be .jpg, .jpeg, or .png format")
		}
	}

	return nil
}

// validateCarousel validates a carousel post
func (s *SDK) validateCarousel(carousel CarouselPost) error {
	if len(carousel.Items) < 2 {
		return fmt.Errorf("carousel must have at least 2 items")
	}
	if len(carousel.Items) > 10 {
		return fmt.Errorf("carousel cannot have more than 10 items")
	}

	for i, item := range carousel.Items {
		if err := s.validateSinglePost(item); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}
	}

	return nil
}
