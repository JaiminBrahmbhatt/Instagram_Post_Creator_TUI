package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewMenu(t *testing.T) {
	l := NewMenu()
	
	// We want to remove the title from the component itself since the App Shell handles it
	if l.Title != "" {
		t.Error("Menu title should be empty (handled by App Shell)")
	}
	
	// Verify delegate is set (implicit by looking at behavior, but we can check items)
	if len(l.Items()) != 4 {
		t.Errorf("Expected 4 menu items, got %d", len(l.Items()))
	}
}

func TestNewFilePicker(t *testing.T) {
	fp := NewFilePicker()
	
	// Check if cursor style has color (not empty)
	if fp.Styles.Cursor.GetForeground() == lipgloss.Color("") {
		t.Error("FilePicker cursor style should have a foreground color")
	}
	
	if fp.Styles.Directory.GetForeground() == lipgloss.Color("") {
		t.Error("FilePicker directory style should have a foreground color")
	}
}
