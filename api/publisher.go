package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// 1. Get media for post
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

		if len(mediaPaths) == 1 {
			// Single Post
			path := mediaPaths[0]
			publicURL, mType, err := s.prepareMedia(path)
			if err != nil {
				s.report("❌ Preparation failed: %v", err)
				s.DB.MarkPostStatus(postID, db.StatusFailed)
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
			// Carousel Post
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

	// 5. Update status
	if !dryRun {
		s.DB.MarkPostPublished(postID)
	}

	if dryRun {
		s.report("✅ [DRY RUN] Post %d complete", postID)
	} else {
		s.report("✅ Successfully finished post %d callback!", postID)
	}
}

// prepareMedia calculates the public URL and detects media type for a file
func (s *Scheduler) prepareMedia(path string) (string, MediaType, error) {
	urlPrefix := getPublicURLPrefix()

	// Detect media type
	ext := strings.ToLower(filepath.Ext(path))
	mType := MediaTypeImage
	if ext == ".mp4" || ext == ".mov" {
		mType = MediaTypeVideo
	}

	// Get photos directory to calculate relative paths
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
		relPath = filepath.Base(path) // Fallback
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
