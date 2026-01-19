package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

func InitDB(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return &Database{Conn: db}, nil
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

func (db *Database) SavePost(caption string, mediaPaths []string, scheduledAt string) (int64, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO posts (caption, scheduled_at, status) VALUES (?, ?, 'draft')", caption, scheduledAt)
	if err != nil {
		return 0, err
	}

	postID, _ := res.LastInsertId()

	for i, path := range mediaPaths {
		// First ensure media exists in DB
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
