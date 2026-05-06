package tui

import (
	"testing"
)

func TestNewMenu(t *testing.T) {
	l := NewMenu()
	
	// We want to remove the title from the component itself since the App Shell handles it
	if l.Title != "" {
		t.Error("Menu title should be empty (handled by App Shell)")
	}
	
	// Verify delegate is set (implicit by looking at behavior, but we can check items)
	if len(l.Items()) != 5 {
		t.Errorf("Expected 5 menu items, got %d", len(l.Items()))
	}
}

