package db

import (
	"database/sql"
	"os"

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
