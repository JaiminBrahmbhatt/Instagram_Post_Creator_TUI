package tui

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/charmbracelet/bubbles/table"
)

func (m *Model) enterBrowserDirectory() {
	selectedRow := m.browserTable.SelectedRow()
	if len(selectedRow) < 2 {
		return
	}
	name := selectedRow[1]
	if !strings.HasPrefix(name, "📁") {
		return
	}
	cleanName := strings.TrimPrefix(name, "📁 ")
	cleanName = strings.TrimSuffix(cleanName, "/")
	newDir := filepath.Join(m.browserDir, cleanName)

	if m.currentView == BrowserView && m.photosDir != "" {
		absNew, _ := filepath.Abs(newDir)
		absRoot, _ := filepath.Abs(m.photosDir)
		rel, err := filepath.Rel(absRoot, absNew)
		if err != nil || strings.HasPrefix(rel, "..") {
			return
		}
	}
	m.browserDir = newDir
	m.filterQuery = ""
	m.refreshBrowserTable()
	m.browserTable.GotoTop()
}

func (m *Model) parentDirectory() {
	newDir := filepath.Dir(m.browserDir)
	if m.currentView == BrowserView && m.photosDir != "" {
		absNew, _ := filepath.Abs(newDir)
		absRoot, _ := filepath.Abs(m.photosDir)
		rel, err := filepath.Rel(absRoot, absNew)
		if err != nil || strings.HasPrefix(rel, "..") {
			return
		}
	}
	m.browserDir = newDir
	m.filterQuery = ""
	m.refreshBrowserTable()
	m.browserTable.GotoTop()
}

func (m *Model) handleBrowserSelection() {
	selectedRow := m.browserTable.SelectedRow()
	if len(selectedRow) < 2 {
		return
	}
	name := selectedRow[1]
	isDir := strings.HasPrefix(name, "📁")
	cleanName := strings.TrimPrefix(name, "📁 ")
	cleanName = strings.TrimSuffix(cleanName, "/")
	cleanName = strings.TrimPrefix(cleanName, "📄 ")

	fullPath := filepath.Join(m.browserDir, cleanName)
	if isDir {
		// Path traversal check
		if (m.currentView == BrowserView) && m.photosDir != "" {
			absNew, _ := filepath.Abs(fullPath)
			absRoot, _ := filepath.Abs(m.photosDir)
			rel, err := filepath.Rel(absRoot, absNew)
			if err != nil || strings.HasPrefix(rel, "..") {
				return
			}
		}
		m.browserDir = fullPath
		m.filterQuery = ""
		m.refreshBrowserTable()
		m.browserTable.GotoTop()
	} else {
		// Toggle selection
		idx := slices.Index(m.selectedMedia, fullPath)
		if idx >= 0 {
			m.selectedMedia = append(m.selectedMedia[:idx], m.selectedMedia[idx+1:]...)
		} else {
			m.selectedMedia = append(m.selectedMedia, fullPath)
		}
		m.refreshBrowserTable()
	}
}

func (m *Model) refreshBrowserTable() {
	onlyDirs := m.currentView == SettingsDirView || (m.currentView == SetupView && m.setupStep == 0)

	postedPaths := make(map[string]bool)
	if rows, err := m.db.Conn.Query("SELECT path FROM media WHERE is_posted=1"); err == nil {
		defer rows.Close()
		for rows.Next() {
			var p string
			rows.Scan(&p)
			abs, _ := filepath.Abs(p)
			postedPaths[abs] = true
		}
	}

	rows := buildBrowserRows(m.browserDir, m.selectedMedia, onlyDirs, postedPaths)
	if m.filterQuery != "" {
		rows = filterBrowserRows(rows, m.filterQuery)
	}
	rows = sortBrowserRows(rows, m.sortMode)

	m.browserTable.SetColumns(browserColumns(m.width - 4))
	m.browserTable.SetRows(rows)
}

func (m *Model) refreshTable() {
	posts, err := m.db.GetPosts(m.tableLimit, m.tableOffset)
	if err != nil {
		m.statusMsg = "Error fetching posts: " + err.Error()
		return
	}
	var rows []table.Row
	for _, p := range posts {
		timeToShow := ""
		switch p.Status {
		case db.StatusDraft:
			timeToShow = m.formatLocalTime(p.CreatedAt)
		case db.StatusPublished:
			timeToShow = m.formatLocalTime(p.PublishedAt)
		default:
			timeToShow = m.formatLocalTime(p.ScheduledAt)
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			strings.ToUpper(string(p.Status)),
			timeToShow,
			fmt.Sprintf("%d", p.MediaCount),
			p.Caption,
		})
	}
	m.table.SetRows(rows)
}
