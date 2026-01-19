package api

const APIVersion = "v24.0"

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
