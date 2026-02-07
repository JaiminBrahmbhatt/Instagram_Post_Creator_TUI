package db

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type PostStatus string

const (
	StatusDraft      PostStatus = "draft"
	StatusScheduled  PostStatus = "scheduled"
	StatusPublishing PostStatus = "publishing"
	StatusPublished  PostStatus = "published"
	StatusFailed     PostStatus = "failed"
)

const CleanupDaysThreshold = 30

type Database struct {
	Conn *sql.DB
}

func InitDB(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	d := &Database{Conn: db}
	if err := d.executeSchema(); err != nil {
		return nil, err
	}
	if err := d.migratePostsTable(); err != nil {
		return nil, err
	}

	return d, nil
}

// migratePostsTable adds new columns to posts for existing databases.
func (db *Database) migratePostsTable() error {
	for _, col := range []string{
		"alt_text TEXT", "location_id TEXT", "user_tags TEXT", "share_to_feed INTEGER DEFAULT 0",
		"cover_url TEXT", "thumb_offset INTEGER DEFAULT 0", "collaborators TEXT", "audio_name TEXT", "post_as_story INTEGER DEFAULT 0",
	} {
		_, err := db.Conn.Exec("ALTER TABLE posts ADD COLUMN " + col)
		if err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

func (db *Database) executeSchema() error {
	paths := []string{
		"db/schema.sql",
		"schema.sql",
		"../db/schema.sql",
	}

	var schema []byte
	var err error

	for _, path := range paths {
		schema, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		schema = []byte(`
CREATE TABLE IF NOT EXISTS media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    is_posted BOOLEAN DEFAULT FALSE,
    ignore BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    caption TEXT,
    scheduled_at TIMESTAMP,
    published_at TIMESTAMP,
    ig_container_id TEXT,
    status TEXT CHECK(status IN ('draft', 'scheduled', 'publishing', 'published', 'failed')) DEFAULT 'draft',
    engagement_likes INTEGER DEFAULT 0,
    engagement_comments INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    alt_text TEXT,
    location_id TEXT,
    user_tags TEXT,
    share_to_feed INTEGER DEFAULT 0,
    cover_url TEXT,
    thumb_offset INTEGER DEFAULT 0,
    collaborators TEXT,
    audio_name TEXT,
    post_as_story INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS post_media (
    post_id INTEGER,
    media_id INTEGER,
    display_order INTEGER,
    ig_item_container_id TEXT,
    PRIMARY KEY (post_id, media_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES media(id)
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
`)
	}

	_, err = db.Conn.Exec(string(schema))
	return err
}
