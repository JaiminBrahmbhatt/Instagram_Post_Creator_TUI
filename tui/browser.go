package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/charmbracelet/bubbles/table"
)

// SortMode for the browser table.
type SortMode int

const (
	SortNameAsc  SortMode = iota
	SortNameDesc
	SortDateAsc
	SortDateDesc
	SortSizeAsc
	SortSizeDesc
	sortModeMax
)

func (s SortMode) String() string {
	switch s {
	case SortNameAsc:
		return "name ↑"
	case SortNameDesc:
		return "name ↓"
	case SortDateAsc:
		return "date ↑"
	case SortDateDesc:
		return "date ↓"
	case SortSizeAsc:
		return "size ↑"
	case SortSizeDesc:
		return "size ↓"
	default:
		return ""
	}
}

// nextSort cycles to the next sort mode.
func nextSort(current SortMode) SortMode {
	return (current + 1) % sortModeMax
}

// filterBrowserRows returns rows whose name column contains query (case-insensitive).
func filterBrowserRows(rows []table.Row, query string) []table.Row {
	if query == "" {
		return rows
	}
	q := strings.ToLower(query)
	var out []table.Row
	for _, r := range rows {
		name := strings.ToLower(r[1])
		if strings.Contains(name, q) {
			out = append(out, r)
		}
	}
	return out
}

// sortBrowserRows sorts rows by the given mode. Directories always sort first.
func sortBrowserRows(rows []table.Row, mode SortMode) []table.Row {
	dirs := make([]table.Row, 0)
	files := make([]table.Row, 0)
	for _, r := range rows {
		if strings.HasPrefix(r[1], "📁") {
			dirs = append(dirs, r)
		} else {
			files = append(files, r)
		}
	}
	sort.SliceStable(files, func(i, j int) bool {
		switch mode {
		case SortNameDesc:
			return files[i][1] > files[j][1]
		case SortDateAsc:
			return files[i][4] < files[j][4]
		case SortDateDesc:
			return files[i][4] > files[j][4]
		case SortSizeAsc:
			return files[i][3] < files[j][3]
		case SortSizeDesc:
			return files[i][3] > files[j][3]
		default: // SortNameAsc
			return files[i][1] < files[j][1]
		}
	})
	return append(dirs, files...)
}

// buildBrowserRows reads a directory and builds table rows.
// postedPaths is a set of absolute paths that have been posted.
func buildBrowserRows(dir string, selectedMedia []string, onlyDirs bool, postedPaths map[string]bool) []table.Row {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var rows []table.Row
	for _, f := range files {
		info, _ := f.Info()
		name := f.Name()
		size, mod, icon := "", "", "📄"
		if f.IsDir() {
			icon = "📁"
			name = name + "/"
		} else {
			if onlyDirs {
				continue
			}
			ext := strings.ToLower(filepath.Ext(name))
			if !slices.Contains(api.SupportedExtensions, ext) {
				continue
			}
			if info != nil {
				size = fmt.Sprintf("%.1f KB", float64(info.Size())/1024)
				mod = info.ModTime().Format("2006-01-02")
			}
		}
		selected := " "
		if !onlyDirs && slices.Contains(selectedMedia, filepath.Join(dir, f.Name())) {
			selected = "✓"
		}
		status := ""
		if !f.IsDir() {
			absPath, _ := filepath.Abs(filepath.Join(dir, f.Name()))
			if postedPaths[absPath] {
				status = "posted"
			}
		}
		rows = append(rows, table.Row{selected, icon + " " + name, status, size, mod})
	}
	return rows
}

// browserColumns returns the table columns for the browser.
func browserColumns(tableWidth int) []table.Column {
	nameWidth := tableWidth - 3 - 10 - 10 - 12
	if nameWidth < 20 {
		nameWidth = 20
	}
	return []table.Column{
		{Title: " ", Width: 3},
		{Title: "Name", Width: nameWidth},
		{Title: "Status", Width: 10},
		{Title: "Size", Width: 10},
		{Title: "Modified", Width: 12},
	}
}
