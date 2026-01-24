package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestNgrokSettingsModel(t *testing.T) {
	m := Model{}
	// This ensures SettingsNgrokView is defined
	m.currentView = SettingsNgrokView

	// This ensures ngrokInput is defined
	m.ngrokInput = textinput.New()
	m.ngrokInput.SetValue("initial-token")

	if m.ngrokInput.Value() != "initial-token" {
		t.Error("Input value mismatch")
	}
}
