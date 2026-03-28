package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestFilterBrowserRows(t *testing.T) {
	rows := []table.Row{
		{" ", "📷 beach.jpg", "", "2.1 KB", "2026-03-20"},
		{" ", "📷 sunset.png", "", "1.4 KB", "2026-03-21"},
		{" ", "📁 archive/", "", "", ""},
	}
	got := filterBrowserRows(rows, "beach")
	if len(got) != 1 {
		t.Errorf("expected 1 filtered row, got %d", len(got))
	}
	if got[0][1] != "📷 beach.jpg" {
		t.Errorf("wrong row returned: %v", got[0])
	}
}

func TestFilterBrowserRows_EmptyQuery(t *testing.T) {
	rows := []table.Row{
		{" ", "📷 a.jpg", "", "1 KB", "2026-03-20"},
		{" ", "📁 dir/", "", "", ""},
	}
	got := filterBrowserRows(rows, "")
	if len(got) != len(rows) {
		t.Errorf("empty query should return all rows, got %d", len(got))
	}
}

func TestSortBrowserRows_ByNameAsc(t *testing.T) {
	rows := []table.Row{
		{" ", "📷 z.jpg", "", "1 KB", "2026-03-22"},
		{" ", "📷 a.jpg", "", "2 KB", "2026-03-20"},
		{" ", "📷 m.jpg", "", "1 KB", "2026-03-21"},
	}
	sorted := sortBrowserRows(rows, SortNameAsc)
	if sorted[0][1] != "📷 a.jpg" {
		t.Errorf("first row should be a.jpg, got %v", sorted[0][1])
	}
	if sorted[2][1] != "📷 z.jpg" {
		t.Errorf("last row should be z.jpg, got %v", sorted[2][1])
	}
}

func TestSortBrowserRows_ByNameDesc(t *testing.T) {
	rows := []table.Row{
		{" ", "📷 a.jpg", "", "1 KB", "2026-03-20"},
		{" ", "📷 z.jpg", "", "2 KB", "2026-03-22"},
	}
	sorted := sortBrowserRows(rows, SortNameDesc)
	if sorted[0][1] != "📷 z.jpg" {
		t.Errorf("first row should be z.jpg, got %v", sorted[0][1])
	}
}
