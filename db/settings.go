package db

import "database/sql"

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
