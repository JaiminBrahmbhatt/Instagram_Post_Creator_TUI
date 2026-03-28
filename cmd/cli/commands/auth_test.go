package commands_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/commands"
)

func TestAuthStatus(t *testing.T) {
	var buf bytes.Buffer
	commands.RunAuthStatus(&buf, false)
	var resp map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected ok, got %v", resp["status"])
	}
	data := resp["data"].(map[string]interface{})
	if _, ok := data["access_token"]; !ok {
		t.Error("access_token field missing")
	}
	if _, ok := data["ig_id"]; !ok {
		t.Error("ig_id field missing")
	}
}
