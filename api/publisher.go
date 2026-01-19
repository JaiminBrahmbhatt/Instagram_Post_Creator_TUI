package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
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
		log.Printf("Publish error (post %d): %v", postID, err)
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if len(mediaPaths) == 0 {
		log.Printf("Error: No media found for post %d", postID)
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if len(mediaPaths) > 10 {
		log.Printf("Error: Post %d has too many media files (%d > 10). Carousel limit is 10.", postID, len(mediaPaths))
		s.DB.MarkPostStatus(postID, db.StatusFailed)
		return
	}

	if dryRun {
		urlPrefix := s.DB.GetConfig("public_url_prefix", "PUBLIC_URL_PREFIX")
		if urlPrefix == "" {
			urlPrefix = "https://example.com/"
		}
		s.report("[DRY RUN] Would upload %d files to %s", len(mediaPaths), urlPrefix)
	} else {
		carouselID, err := s.createContainers(mediaPaths, caption)
		if err != nil {
			log.Printf("Failed to create containers: %v", err)
			s.DB.MarkPostStatus(postID, db.StatusFailed)
			return
		}

		// 3. WAIT for processing before publishing
		s.report("Waiting for Instagram processing...")
		if err := s.Client.WaitForContainer(carouselID); err != nil {
			s.report("❌ Processing failed: %v", err)
			s.DB.MarkPostStatus(postID, db.StatusFailed)
			return
		}

		// 4. Publish
		s.report("Publishing content...")
		if _, err := s.Client.PublishContainer(carouselID); err != nil {
			s.report("❌ Failed to publish: %v", err)
			s.DB.MarkPostStatus(postID, db.StatusFailed)
			return
		}
	}

	// 5. Update status
	if !dryRun {
		s.DB.MarkPostPublished(postID)
	}

	if dryRun {
		s.report("✅ [DRY RUN] Post %d complete", postID)
	} else {
		s.report("✅ Successfully published post %d!", postID)
	}
}

func (s *Scheduler) createContainers(mediaPaths []string, caption string) (string, error) {
	urlPrefix := s.DB.GetConfig("public_url_prefix", "PUBLIC_URL_PREFIX")
	if urlPrefix == "" {
		urlPrefix = "https://example.com/"
	}

	var itemIDs []string
	isCarousel := len(mediaPaths) > 1

	s.report("Creating containers for %d files...", len(mediaPaths))

	// Get photos directory to calculate relative paths
	photosDir, err := s.DB.GetSetting("photos_dir")
	if err != nil || photosDir == "" {
		photosDir = "photos"
	}
	absPhotosDir, err := filepath.Abs(photosDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute photos dir: %w", err)
	}

	for _, path := range mediaPaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
		relPath, err := filepath.Rel(absPhotosDir, absPath)
		if err != nil {
			relPath = filepath.Base(path) // Fallback
		}

		publicURL := urlPrefix + relPath

		// Detect media type
		ext := strings.ToLower(filepath.Ext(path))
		mType := MediaTypeImage
		if ext == ".mp4" || ext == ".mov" {
			mType = MediaTypeVideo
		}

		s.report("  - Uploading %s (%s)", filepath.Base(path), mType)

		// For single posts, pass the caption here.
		// For carousel items, caption must be empty and isCarouselItem=true.
		itemCaption := ""
		if !isCarousel {
			itemCaption = caption
		}

		id, err := s.Client.CreateMediaContainer(publicURL, itemCaption, mType, isCarousel)
		if err != nil {
			return "", err
		}
		itemIDs = append(itemIDs, id)
	}

	if isCarousel {
		s.report("Creating carousel container...")
		return s.Client.CreateCarouselContainer(caption, itemIDs)
	}
	return itemIDs[0], nil
}
