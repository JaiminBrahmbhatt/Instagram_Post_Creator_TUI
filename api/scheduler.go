package api

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jaiminb/insta-auto-post/db"
)

type Scheduler struct {
	Client *Client
	DB     *db.Database
}

func NewScheduler(database *db.Database, client *Client) *Scheduler {
	return &Scheduler{
		Client: client,
		DB:     database,
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
		log.Printf("[DRY RUN] Publishing post %d with caption: %s", postID, caption)
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
		log.Printf("[DRY RUN] Would upload %d files to %s", len(mediaPaths), urlPrefix)
		carouselID = "DRY_RUN_ID"
	} else {
		// 2. Create item containers
		var itemIDs []string
		isCarousel := len(mediaPaths) > 1

		for _, path := range mediaPaths {
			publicURL := urlPrefix + filepath.Base(path)
			// If it's a single image, isCarouselItem=false. If >1, it's true.
			id, err := s.Client.CreateMediaContainer(publicURL, isCarousel)
			if err != nil {
				log.Printf("Failed to create container for %s: %v", path, err)
				s.markFailed(postID)
				return
			}
			itemIDs = append(itemIDs, id)
		}

		if isCarousel {
			id, err := s.Client.CreateCarouselContainer(caption, itemIDs)
			if err != nil {
				log.Printf("Failed to create carousel: %v", err)
				s.markFailed(postID)
				return
			}
			carouselID = id
		} else {
			// Logic for single media post (if needed in future, currently we assume carousel flow OR simple flow)
			// Actually, if it's a single item, we just use that ID as the "containerID" to publish.
			carouselID = itemIDs[0]
		}

		// 3. WAIT for processing before publishing
		log.Printf("Waiting for container %s to be ready...", carouselID)
		if err := s.Client.WaitForContainer(carouselID); err != nil {
			log.Printf("Container processing failed: %v", err)
			s.markFailed(postID)
			return
		}

		// 4. Publish
		if _, err := s.Client.PublishContainer(carouselID); err != nil {
			log.Printf("Failed to publish container: %v", err)
			s.markFailed(postID)
			return
		}
	}

	// 5. Update status
	s.DB.Conn.Exec("UPDATE posts SET status = 'published', published_at = CURRENT_TIMESTAMP WHERE id = ?", postID)
	s.DB.Conn.Exec("UPDATE media SET is_posted = 1 WHERE id IN (SELECT media_id FROM post_media WHERE post_id = ?)", postID)

	if dryRun {
		log.Printf("[DRY RUN] Post %d marked as published", postID)
	} else {
		log.Printf("Successfully published post %d", postID)
	}
}

func (s *Scheduler) markFailed(postID int64) {
	s.DB.Conn.Exec("UPDATE posts SET status = 'failed' WHERE id = ?", postID)
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.CheckAndPublish()
		}
	}()
}
