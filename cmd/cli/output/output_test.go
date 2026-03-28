package output_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
)

func TestOKCompact(t *testing.T) {
	var buf bytes.Buffer
	output.OK(&buf, map[string]string{"key": "val"}, false)
	var resp map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
	data := resp["data"].(map[string]interface{})
	if data["key"] != "val" {
		t.Errorf("expected data.key=val, got %v", data["key"])
	}
}

func TestErrOutput(t *testing.T) {
	var buf bytes.Buffer
	output.Err(&buf, "something broke", false)
	var resp map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "error" {
		t.Errorf("expected status error, got %v", resp["status"])
	}
	if resp["message"] != "something broke" {
		t.Errorf("expected message, got %v", resp["message"])
	}
}

func TestPrettyOutput(t *testing.T) {
	var buf bytes.Buffer
	output.OK(&buf, map[string]int{"count": 1}, true)
	out := buf.String()
	if len(out) < 10 || out[0] != '{' {
		t.Error("expected pretty JSON starting with {")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("pretty JSON invalid: %v", err)
	}
}
