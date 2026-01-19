package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusScheduled PostStatus = "scheduled"
	StatusPublished PostStatus = "published"
	StatusFailed    PostStatus = "failed"
)

type Database struct {
	Conn *sql.DB
}

type Post struct {
	Caption     string
	CreatedAt   string
	ID          int64
	MediaCount  int
	PublishedAt string
	ScheduledAt string
	Status      PostStatus
}

func CalculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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

	return d, nil
}

func (db *Database) executeSchema() error {
	schema, err := os.ReadFile("db/schema.sql")
	if err != nil {
		return err
	}
	_, err = db.Conn.Exec(string(schema))
	return err
}

func (db *Database) GetDirMediaCount(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, f := range files {
		if !f.IsDir() {
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				count++
			}
		}
	}
	return count, nil
}

func (db *Database) GetPosts() ([]Post, error) {
	rows, err := db.Conn.Query(`
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
	`)
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

func (db *Database) GetSetting(key string) (string, error) {
	var value string
	err := db.Conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *Database) GetUnpostedMedia() ([]string, error) {
	rows, err := db.Conn.Query("SELECT path FROM media WHERE is_posted = 0 AND ignore = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func (db *Database) RegisterMedia(path string) error {
	hash, err := CalculateHash(path)
	if err != nil {
		return err
	}

	_, err = db.Conn.Exec(`
		INSERT INTO media (path, hash) VALUES (?, ?)
		ON CONFLICT(path) DO UPDATE SET hash = excluded.hash
	`, path, hash)
	return err
}

func (db *Database) RunCleanup() error {
	autoCleanup, err := db.GetSetting("auto_cleanup")
	if err != nil || autoCleanup != "1" {
		return nil
	}

	rows, err := db.Conn.Query(`
		SELECT DISTINCT m.path FROM media m
		JOIN post_media pm ON m.id = pm.media_id
		JOIN posts p ON pm.post_id = p.id
		WHERE m.is_posted = 1
		AND p.published_at < datetime('now', '-30 days')
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		os.Remove(path)
		db.Conn.Exec("UPDATE media SET ignore = 1 WHERE path = ?", path)
	}
	return nil
}

func (db *Database) SavePost(caption string, mediaPaths []string, scheduledAt string, status PostStatus) (int64, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if scheduledAt == "" {
		scheduledAt = "NULL"
	}

	res, err := tx.Exec("INSERT INTO posts (caption, scheduled_at, status) VALUES (?, datetime('now', ?), ?)", caption, scheduledAt, status)
	if err != nil {
		return 0, err
	}

	postID, _ := res.LastInsertId()

	for i, path := range mediaPaths {
		hash, _ := CalculateHash(path)
		_, err = tx.Exec(`
			INSERT INTO media (path, hash) VALUES (?, ?)
			ON CONFLICT(path) DO NOTHING
		`, path, hash)
		if err != nil {
			return 0, err
		}

		var mediaID int64
		err = tx.QueryRow("SELECT id FROM media WHERE path = ?", path).Scan(&mediaID)
		if err != nil {
			return 0, err
		}

		_, err = tx.Exec("INSERT INTO post_media (post_id, media_id, display_order) VALUES (?, ?, ?)", postID, mediaID, i)
		if err != nil {
			return 0, err
		}
	}

	return postID, tx.Commit()
}

func (db *Database) SetSetting(key, value string) error {
	_, err := db.Conn.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}
