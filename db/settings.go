package db

import (
	"database/sql"
	"os"
)

func (db *Database) GetSetting(key string) (string, error) {
	var value string
	err := db.Conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *Database) SetSetting(key, value string) error {
	_, err := db.Conn.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// GetConfig retrieves a setting with priority: DB > Env > Default
func (db *Database) GetConfig(key string, fallbackEnv string) string {
	val, err := db.GetSetting(key)
	if err == nil && val != "" {
		return val
	}
	if fallbackEnv != "" {
		if envVal := os.Getenv(fallbackEnv); envVal != "" {
			return envVal
		}
	}
	return ""
}

// GetConfigBool retrieves a boolean setting (true/false) with priority: DB > Env > Default
func (db *Database) GetConfigBool(key string, fallbackEnv string) bool {
	val := db.GetConfig(key, fallbackEnv)
	return val == "true" || val == "1"
}
