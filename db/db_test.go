package db

import (
	"testing"
)

func TestInitDBInMemory(t *testing.T) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	_, err = db.Conn.Exec("INSERT INTO settings (key, value) VALUES ('test', 'val')")
	if err != nil {
		t.Fatalf("schema not applied: %v", err)
	}
}
