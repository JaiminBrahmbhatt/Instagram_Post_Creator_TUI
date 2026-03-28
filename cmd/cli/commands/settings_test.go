package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/commands"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func newTestApp(t *testing.T) *appPkg.App {
	t.Helper()
	d, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return &appPkg.App{DB: d}
}

func execCmd(t *testing.T, cmd *cobra.Command, args []string) map[string]interface{} {
	t.Helper()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Logf("command output: %s", buf.String())
		t.Fatalf("command error: %v", err)
	}
	out := buf.String()
	// Find the last JSON line (cobra may print usage on errors)
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	return resp
}

func TestSettingsSetAndGet(t *testing.T) {
	a := newTestApp(t)
	pretty := false

	settingsCmd := commands.NewSettingsCmd(&a, &pretty)
	resp := execCmd(t, settingsCmd, []string{"set", "--key", "photos_dir", "--value", "/tmp/photos"})
	if resp["status"] != "ok" {
		t.Fatalf("set failed: %v", resp)
	}

	settingsCmd2 := commands.NewSettingsCmd(&a, &pretty)
	resp = execCmd(t, settingsCmd2, []string{"get", "--key", "photos_dir"})
	if resp["status"] != "ok" {
		t.Fatalf("get failed: %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["value"] != "/tmp/photos" {
		t.Errorf("expected /tmp/photos, got %v", data["value"])
	}
}

func TestSettingsList(t *testing.T) {
	a := newTestApp(t)
	a.DB.SetSetting("foo", "bar")
	pretty := false
	settingsCmd := commands.NewSettingsCmd(&a, &pretty)
	resp := execCmd(t, settingsCmd, []string{"list"})
	if resp["status"] != "ok" {
		t.Fatalf("list failed: %v", resp)
	}
}

// Prevent unused import
var _ = fmt.Sprintf
