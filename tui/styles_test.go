package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestThemePalette(t *testing.T) {
	if string(Theme.Primary) == "" {
		t.Error("Theme.Primary is empty")
	}
	if string(Theme.Background) == "" {
		t.Error("Theme.Background is empty")
	}
	if string(Theme.Surface) == "" {
		t.Error("Theme.Surface is empty")
	}
	if string(Theme.TextPrimary) == "" {
		t.Error("Theme.TextPrimary is empty")
	}
}

func TestNewPalette(t *testing.T) {
	if string(Theme.Primary) != "#DA7756" {
		t.Errorf("Primary should be #DA7756 (Claude orange), got %s", Theme.Primary)
	}
	if string(Theme.Background) != "#1A1A1A" {
		t.Errorf("Background should be #1A1A1A, got %s", Theme.Background)
	}
	if string(Theme.Surface) != "#1C1C1C" {
		t.Errorf("Surface should be #1C1C1C, got %s", Theme.Surface)
	}
	if string(Theme.Border) != "#333333" {
		t.Errorf("Border should be #333333, got %s", Theme.Border)
	}
}

func TestNewStyles(t *testing.T) {
	// Red phase: specific modern styles we want to implement
	if AppHeaderStyle.GetForeground() == lipgloss.Color("") {
		t.Error("AppHeaderStyle should have a foreground color")
	}

	if CardStyle.GetBorderStyle().Top == "" {
		t.Error("CardStyle should have a border")
	}

	if BadgeStyle.GetPaddingLeft() == 0 {
		t.Error("BadgeStyle should have padding")
	}
}
