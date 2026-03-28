package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	return &Model{
		currentView:       MenuView,
		authInputs:        NewEnvInputs(),
		ngrokInput:        NewNgrokInput(),
		domainInput:       NewDomainInput(),
		ngrokFieldVisible: []bool{false, false},
		authFieldVisible:  []bool{false, false, false},
	}
}

func TestShiftTabTogglesNgrokVisibility(t *testing.T) {
	m := newTestModel(t)
	m.currentView = SettingsNgrokView
	m.ngrokFocusIndex = 0
	m.ngrokFieldVisible = []bool{false, false}
	m.ngrokInput.Focus()

	// Initial state: token field should be EchoPassword
	if m.ngrokInput.EchoMode != textinput.EchoPassword {
		t.Skip("ngrokInput not initialized with EchoPassword — skipping")
	}

	// Send shift+tab key
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newM, _ := m.Update(shiftTabMsg)
	updated := newM.(*Model)

	if !updated.ngrokFieldVisible[0] {
		t.Error("after shift+tab, field 0 should be visible")
	}
	if updated.ngrokInput.EchoMode != textinput.EchoNormal {
		t.Error("after shift+tab, echo mode should be Normal")
	}

	// Second shift+tab: should hide again
	newM2, _ := updated.Update(shiftTabMsg)
	updated2 := newM2.(*Model)
	if updated2.ngrokFieldVisible[0] {
		t.Error("after second shift+tab, field 0 should be hidden again")
	}
}
