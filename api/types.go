package api

const APIVersion = "v19.0"

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
