package commands_test

import (
	"fmt"
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/commands"
)

func TestMediaList_Empty(t *testing.T) {
	a := newTestApp(t)
	pretty := false
	mediaCmd := commands.NewMediaCmd(&a, &pretty)
	resp := execCmd(t, mediaCmd, []string{"list"})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	items := data["media"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}

func TestMediaIgnore(t *testing.T) {
	a := newTestApp(t)
	_, err := a.DB.Conn.Exec(`INSERT INTO media (path, hash) VALUES ('test.jpg', 'abc123')`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	var id int64
	a.DB.Conn.QueryRow("SELECT id FROM media WHERE path='test.jpg'").Scan(&id)

	pretty := false
	mediaCmd := commands.NewMediaCmd(&a, &pretty)
	resp := execCmd(t, mediaCmd, []string{"ignore", "--id", fmt.Sprintf("%d", id)})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	var ignored bool
	a.DB.Conn.QueryRow("SELECT ignore FROM media WHERE id=?", id).Scan(&ignored)
	if !ignored {
		t.Error("expected media to be marked ignored")
	}
}
