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
	if string(Theme.Text) == "" {
		t.Error("Theme.Text is empty")
	}
}

func TestNewStyles(t *testing.T) {
	// Red phase: specific modern styles we want to implement
	if AppTitleStyle.GetForeground() == lipgloss.Color("") {
		t.Error("AppTitleStyle should have a foreground color")
	}

	if CardStyle.GetBorderStyle().Top == "" {
		t.Error("CardStyle should have a border")
	}

	if BadgeStyle.GetPaddingLeft() == 0 {
		t.Error("BadgeStyle should have padding")
	}
}
