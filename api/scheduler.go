package api

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jaiminb/insta-auto-post/db"
)

type Scheduler struct {
	DB     *db.Database
	Client *Client
}

func NewScheduler(database *db.Database, client *Client) *Scheduler {
	return &Scheduler{
		DB:     database,
		Client: client,
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.CheckAndPublish()
		}
	}()
}

func (s *Scheduler) CheckAndPublish() {
	// Query for posts that are scheduled and due
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
		urlPrefix = "https://example.com/" // Fallback
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

	var carouselID string

	if dryRun {
		log.Printf("[DRY RUN] Would upload %d files to %s", len(mediaPaths), urlPrefix)
		carouselID = "DRY_RUN_ID"
	} else {
		// 2. Create item containers
		var itemIDs []string
		for _, path := range mediaPaths {
			// For personal usage, we'll assume the user has a way to serve these
			// or we'll need to upload them to a temporary host.
			publicURL := urlPrefix + filepath.Base(path)
			id, err := s.Client.CreateMediaContainer(publicURL, true)
			if err != nil {
				log.Printf("Failed to create container for %s: %v", path, err)
				return
			}
			itemIDs = append(itemIDs, id)
		}

		// 3. Create Carousel container
		id, err := s.Client.CreateCarouselContainer(caption, itemIDs)
		if err != nil {
			log.Printf("Failed to create carousel: %v", err)
			return
		}
		carouselID = id

		// 4. Publish
		_, err = s.Client.PublishContainer(carouselID)
		if err != nil {
			log.Printf("Failed to publish carousel: %v", err)
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
