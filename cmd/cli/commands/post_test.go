package commands_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/commands"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
)

// createTempMediaFile creates a real temp file so CalculateHash works in SavePost.
func createTempMediaFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("createTempMediaFile: %v", err)
	}
	return path
}

func TestPostCreate_Draft(t *testing.T) {
	a := newTestApp(t)
	dir := t.TempDir()
	path := createTempMediaFile(t, dir, "a.jpg")

	pretty := false
	postCmd := commands.NewPostCmd(&a, &pretty)
	resp := execCmd(t, postCmd, []string{"create", "--media", path, "--caption", "hello"})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["status"] != "draft" {
		t.Errorf("expected draft, got %v", data["status"])
	}
	if data["post_id"] == nil {
		t.Error("post_id missing")
	}
}

func TestPostCreate_Scheduled(t *testing.T) {
	a := newTestApp(t)
	dir := t.TempDir()
	path := createTempMediaFile(t, dir, "b.jpg")

	pretty := false
	postCmd := commands.NewPostCmd(&a, &pretty)
	resp := execCmd(t, postCmd, []string{"create", "--media", path, "--caption", "hi", "--schedule", "2026-12-01T10:00:00Z"})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["status"] != "scheduled" {
		t.Errorf("expected scheduled, got %v", data["status"])
	}
}

func TestPostList(t *testing.T) {
	a := newTestApp(t)
	dir := t.TempDir()
	path := createTempMediaFile(t, dir, "c.jpg")

	a.DB.SavePost("test caption", []string{path}, "", db.StatusDraft)

	pretty := false
	postCmd := commands.NewPostCmd(&a, &pretty)
	resp := execCmd(t, postCmd, []string{"list"})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	posts := data["posts"].([]interface{})
	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}
}

func TestPostDelete(t *testing.T) {
	a := newTestApp(t)
	dir := t.TempDir()
	path := createTempMediaFile(t, dir, "d.jpg")

	id, _ := a.DB.SavePost("del me", []string{path}, "", db.StatusDraft)

	pretty := false
	postCmd := commands.NewPostCmd(&a, &pretty)
	resp := execCmd(t, postCmd, []string{"delete", "--id", fmt.Sprintf("%d", id)})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
}

func TestPostList_StatusFilter(t *testing.T) {
	d, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	dir := t.TempDir()
	p1 := filepath.Join(dir, "e.jpg")
	os.WriteFile(p1, []byte("img"), 0644)
	d.SavePost("draft post", []string{p1}, "", db.StatusDraft)
	d.SavePost("scheduled post", []string{p1}, "2026-12-01T10:00:00Z", db.StatusScheduled)

	testApp := &appPkg.App{DB: d}
	pretty := false
	postCmd := commands.NewPostCmd(&testApp, &pretty)

	resp := execCmd(t, postCmd, []string{"list", "--status", "draft"})
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	posts := data["posts"].([]interface{})
	if len(posts) != 1 {
		t.Errorf("expected 1 draft post, got %d", len(posts))
	}
}

func TestPostPublish_NilScheduler(t *testing.T) {
	d, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	pretty := false
	testApp := &appPkg.App{DB: d}
	postCmd := commands.NewPostCmd(&testApp, &pretty)
	resp := execCmd(t, postCmd, []string{"publish", "--id", "1"})
	if resp["status"] != "error" {
		t.Errorf("expected error when scheduler is nil, got %v", resp["status"])
	}
}
