package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	instagram "github.com/JaiminBrahmbhatt/insta-go-sdk"
)

func (s *Scheduler) PublishPost(postID int64, caption string) {
	dryRun := s.DB.GetConfigBool("dry_run", "DRY_RUN")

	if dryRun {
		s.report("[DRY RUN] Publishing post %d...", postID)
	} else {
		s.report("Publishing post %d: %s", postID, caption)
	}

	mediaPaths, err := s.DB.GetPostMedia(postID)
	if err != nil {
		s.report("❌ Publish error (post %d): %v", postID, err)
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if len(mediaPaths) == 0 {
		s.report("❌ Error: No media found for post %d", postID)
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if len(mediaPaths) > 10 {
		s.report("❌ Error: Post %d has too many media files (%d > 10).", postID, len(mediaPaths))
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if dryRun {
		urlPrefix := getPublicURLPrefix()
		s.report("[DRY RUN] Would upload %d files to %s", len(mediaPaths), urlPrefix)
	} else {
		ctx := context.Background()
		var igPostID string

		options, _ := s.DB.GetPostOptions(postID)

		if len(mediaPaths) == 1 {
			path := mediaPaths[0]
			publicURL, mType, err := s.prepareMedia(path)
			if err != nil {
				s.report("❌ Preparation failed: %v", err)
				s.DB.MarkPostStatus(postID, db.StatusFailed)
				return
			}

			// Use Graph client for single image when alt_text, location_id, or user_tags are set (SDK doesn't support these).
			if mType == MediaTypeImage && (options.AltText != "" || options.LocationID != "" || len(options.UserTags) > 0) {
				igPostID, err = s.publishSingleImageWithOptions(ctx, publicURL, caption, options)
				if err != nil {
					s.report("❌ API Error: %v", err)
					s.DB.MarkPostStatus(postID, db.StatusFailed)
					return
				}
				s.report("✅ Successfully published! Instagram ID: %s", igPostID)
				s.DB.MarkPostPublished(postID)
				s.report("✅ Successfully finished post %d callback!", postID)
				return
			}

			// Publish as Story (single image or video, 24h expiry).
			if options.PostAsStory {
				igPostID, err = s.publishStory(ctx, publicURL, mType == MediaTypeVideo, caption, options.UserTags)
				if err != nil {
					s.report("❌ API Error: %v", err)
					s.DB.MarkPostStatus(postID, db.StatusFailed)
					return
				}
				s.report("✅ Successfully published story! Instagram ID: %s", igPostID)
				s.DB.MarkPostPublished(postID)
				s.report("✅ Successfully finished post %d callback!", postID)
				return
			}

			// Use Graph client for single video (Reel) when any reel option is set (share_to_feed, cover, thumb, collaborators, audio).
			reelOpts := options.ShareToFeed || options.CoverURL != "" || options.ThumbOffset > 0 || len(options.Collaborators) > 0 || options.AudioName != ""
			if mType == MediaTypeVideo && reelOpts {
				igPostID, err = s.publishSingleReelWithOptions(ctx, publicURL, caption, options)
				if err != nil {
					s.report("❌ API Error: %v", err)
					s.DB.MarkPostStatus(postID, db.StatusFailed)
					return
				}
				s.report("✅ Successfully published! Instagram ID: %s", igPostID)
				s.DB.MarkPostPublished(postID)
				s.report("✅ Successfully finished post %d callback!", postID)
				return
			}

			post := instagram.Post{
				Caption: caption,
				Type:    mType,
			}
			if mType == MediaTypeVideo {
				post.VideoURL = publicURL
			} else {
				post.ImageURL = publicURL
			}

			s.report("Creating and publishing single post...")
			igPostID, err = s.Client.SDK.PublishSinglePost(ctx, post)
		} else {
			s.report("Preparing %d items for carousel...", len(mediaPaths))
			var items []instagram.Post
			for _, path := range mediaPaths {
				publicURL, mType, err := s.prepareMedia(path)
				if err != nil {
					s.report("❌ Preparation failed for %s: %v", path, err)
					s.DB.MarkPostStatus(postID, db.StatusFailed)
					return
				}

				item := instagram.Post{
					Type: mType,
				}
				if mType == MediaTypeVideo {
					item.VideoURL = publicURL
				} else {
					item.ImageURL = publicURL
				}
				items = append(items, item)
			}

			s.report("Creating and publishing carousel container...")
			igPostID, err = s.Client.SDK.PublishCarousel(ctx, instagram.CarouselPost{
				Items:   items,
				Caption: caption,
			})
		}

		if err != nil {
			s.report("❌ API Error: %v", err)
			s.DB.MarkPostStatus(postID, db.StatusFailed)
			return
		}
		s.report("✅ Successfully published! Instagram ID: %s", igPostID)
	}

	if !dryRun {
		s.DB.MarkPostPublished(postID)
	}

	if dryRun {
		s.report("✅ [DRY RUN] Post %d complete", postID)
	} else {
		s.report("✅ Successfully finished post %d callback!", postID)
	}
}

func (s *Scheduler) prepareMedia(path string) (string, MediaType, error) {
	urlPrefix := getPublicURLPrefix()

	ext := strings.ToLower(filepath.Ext(path))
	mType := MediaTypeImage
	if ext == ".mp4" || ext == ".mov" {
		mType = MediaTypeVideo
	}

	photosDir, err := s.DB.GetSetting("photos_dir")
	if err != nil || photosDir == "" {
		photosDir = "photos"
	}
	absPhotosDir, err := filepath.Abs(photosDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute photos dir: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}
	relPath, err := filepath.Rel(absPhotosDir, absPath)
	if err != nil {
		relPath = filepath.Base(path)
	}

	publicURL := urlPrefix + relPath
	return publicURL, mType, nil
}

func getPublicURLPrefix() string {
	if url := GetTunnelURL(); url != "" {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		return url
	}
	if url := os.Getenv("PUBLIC_URL_PREFIX"); url != "" {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		return url
	}
	return "https://example.com/"
}

// publishSingleImageWithOptions creates a container with alt_text/location_id/user_tags, polls until FINISHED, then publishes.
func (s *Scheduler) publishSingleImageWithOptions(ctx context.Context, imageURL, caption string, options db.PostOptions) (string, error) {
	graph := NewGraphClient(s.Client.AccessToken, s.Client.IGID)
	containerID, err := graph.CreateImageContainer(ctx, imageURL, CreateImageContainerOptions{
		Caption:    caption,
		AltText:    options.AltText,
		LocationID: options.LocationID,
		UserTags:   options.UserTags,
	})
	if err != nil {
		return "", err
	}
	s.report("Container created, waiting for processing...")

	// Poll status (recommended: once per minute, max ~5 minutes)
	const maxWait = 5 * time.Minute
	const pollInterval = 20 * time.Second
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)
		status, err := graph.GetContainerStatus(ctx, containerID)
		if err != nil {
			return "", err
		}
		switch status {
		case ContainerStatusFinished:
			s.report("Publishing container...")
			return graph.PublishContainer(ctx, containerID)
		case ContainerStatusError, ContainerStatusExpired:
			return "", fmt.Errorf("container status: %s", status)
		case ContainerStatusInProgress, ContainerStatusPublished:
			continue
		default:
			continue
		}
	}
	return "", fmt.Errorf("container did not finish within %v", maxWait)
}

// publishSingleReelWithOptions creates a Reels container with optional cover, thumb_offset, collaborators, audio_name, share_to_feed; polls then publishes.
func (s *Scheduler) publishSingleReelWithOptions(ctx context.Context, videoURL, caption string, options db.PostOptions) (string, error) {
	graph := NewGraphClient(s.Client.AccessToken, s.Client.IGID)
	opts := ReelContainerOptions{
		Caption:       caption,
		ShareToFeed:   options.ShareToFeed,
		CoverURL:      options.CoverURL,
		ThumbOffset:   options.ThumbOffset,
		Collaborators: options.Collaborators,
		AudioName:     options.AudioName,
	}
	containerID, err := graph.CreateReelContainer(ctx, videoURL, opts)
	if err != nil {
		return "", err
	}
	s.report("Reel container created, waiting for processing...")
	return s.pollAndPublish(ctx, graph, containerID, "reel")
}

// publishStory creates a Story container (image or video), polls until FINISHED, then publishes. Stories expire after 24h.
func (s *Scheduler) publishStory(ctx context.Context, mediaURL string, isVideo bool, caption string, userTags []string) (string, error) {
	graph := NewGraphClient(s.Client.AccessToken, s.Client.IGID)
	containerID, err := graph.CreateStoryContainer(ctx, mediaURL, isVideo, caption, userTags)
	if err != nil {
		return "", err
	}
	s.report("Story container created, waiting for processing...")
	return s.pollAndPublish(ctx, graph, containerID, "story")
}

func (s *Scheduler) pollAndPublish(ctx context.Context, graph *GraphClient, containerID, label string) (string, error) {
	const maxWait = 5 * time.Minute
	const pollInterval = 20 * time.Second
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)
		status, err := graph.GetContainerStatus(ctx, containerID)
		if err != nil {
			return "", err
		}
		switch status {
		case ContainerStatusFinished:
			s.report("Publishing %s...", label)
			return graph.PublishContainer(ctx, containerID)
		case ContainerStatusError, ContainerStatusExpired:
			return "", fmt.Errorf("container status: %s", status)
		default:
			continue
		}
	}
	return "", fmt.Errorf("container did not finish within %v", maxWait)
}
