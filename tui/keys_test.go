package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeysTabAndShiftTabAreSeparate(t *testing.T) {
	tabKeys := Keys.Tab.Keys()
	shiftTabKeys := Keys.ShiftTab.Keys()

	for _, k := range tabKeys {
		if k == "shift+tab" {
			t.Error("Tab binding must not include shift+tab")
		}
	}
	for _, k := range shiftTabKeys {
		if k == "tab" {
			t.Error("ShiftTab binding must not include tab")
		}
	}
	_ = key.NewBinding // ensure import used
}

func TestBackBindingIsEscOnly(t *testing.T) {
	for _, k := range Keys.Back.Keys() {
		if k == "q" {
			t.Error("Back binding must not include 'q' — q is Quit only")
		}
	}
}

func TestFilterAndSortBindingsExist(t *testing.T) {
	if len(Keys.Filter.Keys()) == 0 {
		t.Error("Filter key binding missing")
	}
	if len(Keys.SortCycle.Keys()) == 0 {
		t.Error("SortCycle key binding missing")
	}
}
