package api

import (
	instagram "github.com/JaiminBrahmbhatt/insta-go-sdk"
)

const APIVersion = "v24.0"

type MediaType = instagram.MediaType

const (
	MediaTypeImage              = instagram.MediaTypeImage
	MediaTypeReels              = instagram.MediaTypeReels
	MediaTypeStories            = instagram.MediaTypeStories
	MediaTypeVideo              = instagram.MediaTypeVideo
	MediaTypeCarousel MediaType = "CAROUSEL"
)

type ContainerStatus = instagram.ContainerStatus

const (
	ContainerStatusExpired    = instagram.StatusExpired
	ContainerStatusError      = instagram.StatusError
	ContainerStatusFinished   = instagram.StatusFinished
	ContainerStatusInProgress = instagram.StatusInProgress
	ContainerStatusPublished  = instagram.StatusPublished
)

var SupportedExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov"}

type ContainerResponse struct {
	ID string `json:"id"`
}

type ContainerStatusResponse struct {
	ID         string          `json:"id"`
	StatusCode ContainerStatus `json:"status_code"`
}

type PublishingLimit = instagram.PublishingLimit
