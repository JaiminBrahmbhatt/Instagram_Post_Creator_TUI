package db

import (
	"encoding/json"
	"fmt"
)

type Post struct {
	Caption     string
	CreatedAt   string
	ID          int64
	MediaCount  int
	PublishedAt string
	ScheduledAt string
	Status      PostStatus
}

func (db *Database) GetPosts(limit, offset int) ([]Post, error) {
	query := `
		SELECT p.id, p.caption,
		       COALESCE(p.scheduled_at, ''),
		       COALESCE(p.published_at, ''),
		       p.created_at,
		       p.status,
		       COUNT(pm.media_id)
		FROM posts p
		LEFT JOIN post_media pm ON p.id = pm.post_id
		GROUP BY p.id
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.Conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Caption, &p.ScheduledAt, &p.PublishedAt, &p.CreatedAt, &p.Status, &p.MediaCount); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (db *Database) GetScheduledPosts() ([]Post, error) {
	rows, err := db.Conn.Query(`
		SELECT id, caption FROM posts
		WHERE status = 'scheduled'
		AND scheduled_at <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Caption); err != nil {
			continue
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (db *Database) GetPostMedia(postID int64) ([]string, error) {
	rows, err := db.Conn.Query(`
		SELECT m.path FROM media m
		JOIN post_media pm ON m.id = pm.media_id
		WHERE pm.post_id = ?
		ORDER BY pm.display_order ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaPaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		mediaPaths = append(mediaPaths, path)
	}
	return mediaPaths, nil
}

func (db *Database) MarkPostStatus(postID int64, status PostStatus) error {
	_, err := db.Conn.Exec("UPDATE posts SET status = ? WHERE id = ?", status, postID)
	return err
}

func (db *Database) MarkPostPublished(postID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE posts SET status = 'published', published_at = CURRENT_TIMESTAMP WHERE id = ?", postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE media SET is_posted = 1 WHERE id IN (SELECT media_id FROM post_media WHERE post_id = ?)", postID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// PostOptions holds optional per-post fields (accessibility, location, user tags, reel options, story).
type PostOptions struct {
	AltText      string
	LocationID   string
	UserTags     []string // Instagram usernames to tag
	ShareToFeed  bool     // for Reels: also show in main feed
	CoverURL     string   // for Reels: custom cover image URL (JPEG, 8MB max)
	ThumbOffset  int      // for Reels: thumbnail frame in ms (0 = not set)
	Collaborators []string // for Reels: co-author usernames
	AudioName    string   // for Reels: original audio attribution
	PostAsStory  bool     // publish as Story (24h, single image or video)
}

func (db *Database) GetPostOptions(postID int64) (PostOptions, error) {
	var o PostOptions
	var userTagsRaw, collabRaw string
	var shareToFeed, postAsStory int
	err := db.Conn.QueryRow(
		`SELECT COALESCE(alt_text,''), COALESCE(location_id,''), COALESCE(user_tags,'[]'), COALESCE(share_to_feed,0),
		 COALESCE(cover_url,''), COALESCE(thumb_offset,0), COALESCE(collaborators,'[]'), COALESCE(audio_name,''), COALESCE(post_as_story,0) FROM posts WHERE id = ?`,
		postID,
	).Scan(&o.AltText, &o.LocationID, &userTagsRaw, &shareToFeed, &o.CoverURL, &o.ThumbOffset, &collabRaw, &o.AudioName, &postAsStory)
	if err != nil {
		return o, err
	}
	_ = json.Unmarshal([]byte(userTagsRaw), &o.UserTags)
	_ = json.Unmarshal([]byte(collabRaw), &o.Collaborators)
	o.ShareToFeed = shareToFeed == 1
	o.PostAsStory = postAsStory == 1
	return o, nil
}

func (db *Database) SavePost(caption string, mediaPaths []string, scheduledAt string, status PostStatus, options PostOptions) (int64, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// scheduled_at: full datetime string (UTC) for scheduled posts, or nil for draft
	var scheduledAtVal interface{}
	if scheduledAt != "" {
		scheduledAtVal = scheduledAt
	}

	userTagsJSON, _ := json.Marshal(options.UserTags)
	if userTagsJSON == nil {
		userTagsJSON = []byte("[]")
	}
	collabJSON, _ := json.Marshal(options.Collaborators)
	if collabJSON == nil {
		collabJSON = []byte("[]")
	}
	shareVal := 0
	if options.ShareToFeed {
		shareVal = 1
	}
	storyVal := 0
	if options.PostAsStory {
		storyVal = 1
	}
	res, err := tx.Exec(
		`INSERT INTO posts (caption, scheduled_at, status, alt_text, location_id, user_tags, share_to_feed, cover_url, thumb_offset, collaborators, audio_name, post_as_story) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		caption, scheduledAtVal, status, options.AltText, options.LocationID, string(userTagsJSON), shareVal,
		options.CoverURL, options.ThumbOffset, string(collabJSON), options.AudioName, storyVal,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert post: %w", err)
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	for i, path := range mediaPaths {
		hash, err := CalculateHash(path)
		if err != nil {
			return 0, fmt.Errorf("failed to calculate hash for %s: %w", path, err)
		}

		_, err = tx.Exec(`
			INSERT INTO media (path, hash) VALUES (?, ?)
			ON CONFLICT(path) DO UPDATE SET hash = excluded.hash
		`, path, hash)
		if err != nil {
			return 0, fmt.Errorf("failed to register media %s: %w", path, err)
		}

		var mediaID int64
		err = tx.QueryRow("SELECT id FROM media WHERE path = ?", path).Scan(&mediaID)
		if err != nil {
			return 0, fmt.Errorf("failed to get media id for %s: %w", path, err)
		}

		_, err = tx.Exec("INSERT INTO post_media (post_id, media_id, display_order) VALUES (?, ?, ?)", postID, mediaID, i)
		if err != nil {
			return 0, fmt.Errorf("failed to link media %d to post %d: %w", mediaID, postID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return postID, nil
}
