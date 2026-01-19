package db

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

func (db *Database) GetDirMediaCount(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		// Match extensions from api constants (shared)
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".mp4" || ext == ".mov" {
			count++
		}
	}
	return count, nil
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
		AND p.published_at < datetime('now', '-' || ? || ' days')
	`, CleanupDaysThreshold)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		if err := os.Remove(path); err != nil {
			// Log error but continue
			continue
		}
		db.Conn.Exec("UPDATE media SET ignore = 1 WHERE path = ?", path)
	}
	return nil
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
