package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaiminb/insta-auto-post/db"
)

type Scheduler struct {
	Client      *Client
	DB          *db.Database
	ReportChan  chan string
	TriggerChan chan struct{}
}

func NewScheduler(database *db.Database, client *Client) *Scheduler {
	return &Scheduler{
		Client:      client,
		DB:          database,
		ReportChan:  make(chan string, 10),
		TriggerChan: make(chan struct{}, 1),
	}
}

func (s *Scheduler) CheckAndPublish() {
	rows, err := s.DB.Conn.Query(`
		SELECT id, caption FROM posts
		WHERE status = 'scheduled'
		AND scheduled_at <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		log.Printf("Scheduler error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var postID int64
		var caption string
		if err := rows.Scan(&postID, &caption); err != nil {
			continue
		}

		go s.PublishPost(postID, caption)
	}
}

func (s *Scheduler) PublishPost(postID int64, caption string) {
	dryRun := os.Getenv("DRY_RUN") == "true"
	urlPrefix := os.Getenv("PUBLIC_URL_PREFIX")
	if urlPrefix == "" {
		urlPrefix = "https://example.com/"
	}

	if dryRun {
		s.report("[DRY RUN] Publishing post %d...", postID)
	} else {
		s.report("Publishing post %d: %s", postID, caption)
	}

	// 1. Get media for post
	rows, err := s.DB.Conn.Query(`
		SELECT m.path FROM media m
		JOIN post_media pm ON m.id = pm.media_id
		WHERE pm.post_id = ?
		ORDER BY pm.display_order ASC
	`, postID)
	if err != nil {
		log.Printf("Publish error (post %d): %v", postID, err)
		return
	}
	defer rows.Close()

	var mediaPaths []string
	for rows.Next() {
		var path string
		rows.Scan(&path)
		mediaPaths = append(mediaPaths, path)
	}

	if len(mediaPaths) == 0 {
		log.Printf("Error: No media found for post %d", postID)
		s.markFailed(postID)
		return
	}

	if len(mediaPaths) > 10 {
		log.Printf("Error: Post %d has too many media files (%d > 10). Carousel limit is 10.", postID, len(mediaPaths))
		s.markFailed(postID)
		return
	}

	var carouselID string

	if dryRun {
		s.report("[DRY RUN] Would upload %d files to %s", len(mediaPaths), urlPrefix)
		carouselID = "DRY_RUN_ID"
	} else {
		// 2. Create item containers
		var itemIDs []string
		isCarousel := len(mediaPaths) > 1

		s.report("Creating containers for %d files...", len(mediaPaths))

		// Get photos directory to calculate relative paths
		var photosDir string
		s.DB.Conn.QueryRow("SELECT value FROM settings WHERE key = 'photos_dir'").Scan(&photosDir)
		if photosDir == "" {
			photosDir = "photos"
		}
		absPhotosDir, _ := filepath.Abs(photosDir)

		for _, path := range mediaPaths {
			absPath, _ := filepath.Abs(path)
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

			// TEST OVERRIDE: If the filename is internet_test.jpg, use a real public URL
			// so Instagram can actually download it for the test.
			if filepath.Base(path) == "internet_test.jpg" {
				publicURL = "https://picsum.photos/seed/insta_auto_post/1080/1080.jpg"
			}

			log.Printf("Creating media container for URL: %s", publicURL)

			// For single posts, pass the caption here.
			// For carousel items, caption must be empty and isCarouselItem=true.
			itemCaption := ""
			if !isCarousel {
				itemCaption = caption
			}

			id, err := s.Client.CreateMediaContainer(publicURL, itemCaption, mType, isCarousel)
			if err != nil {
				log.Printf("Failed to create container for %s: %v", path, err)
				s.markFailed(postID)
				return
			}
			itemIDs = append(itemIDs, id)
		}

		if isCarousel {
			s.report("Creating carousel container...")
			id, err := s.Client.CreateCarouselContainer(caption, itemIDs)
			if err != nil {
				s.report("❌ Failed to create carousel: %v", err)
				s.markFailed(postID)
				return
			}
			carouselID = id
		} else {
			carouselID = itemIDs[0]
		}

		// 3. WAIT for processing before publishing
		s.report("Waiting for Instagram processing...")
		if err := s.Client.WaitForContainer(carouselID); err != nil {
			s.report("❌ Processing failed: %v", err)
			s.markFailed(postID)
			return
		}

		// 4. Publish
		s.report("Publishing content...")
		if _, err := s.Client.PublishContainer(carouselID); err != nil {
			s.report("❌ Failed to publish: %v", err)
			s.markFailed(postID)
			return
		}
	}

	// 5. Update status
	s.DB.Conn.Exec("UPDATE posts SET status = 'published', published_at = CURRENT_TIMESTAMP WHERE id = ?", postID)
	s.DB.Conn.Exec("UPDATE media SET is_posted = 1 WHERE id IN (SELECT media_id FROM post_media WHERE post_id = ?)", postID)

	if dryRun {
		s.report("✅ [DRY RUN] Post %d complete", postID)
	} else {
		s.report("✅ Successfully published post %d!", postID)
	}
}

func (s *Scheduler) report(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println(msg) // Now safely writing to debug.log
	select {
	case s.ReportChan <- msg:
	default:
		// Drop if full to avoid blocking
	}
}

func (s *Scheduler) markFailed(postID int64) {
	s.DB.Conn.Exec("UPDATE posts SET status = 'failed' WHERE id = ?", postID)
}

func (s *Scheduler) Trigger() {
	select {
	case s.TriggerChan <- struct{}{}:
	default:
		// Already triggered
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CheckAndPublish()
			case <-s.TriggerChan:
				s.CheckAndPublish()
			}
		}
	}()
}
