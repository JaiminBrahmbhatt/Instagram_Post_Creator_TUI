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

func TestBackBindingIncludesQ(t *testing.T) {
	found := false
	for _, k := range Keys.Back.Keys() {
		if k == "q" {
			found = true
		}
	}
	if !found {
		t.Error("Back binding must include 'q'")
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
