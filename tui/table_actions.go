package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/jaiminb/insta-auto-post/api"
	"github.com/jaiminb/insta-auto-post/db"
)

func (m *Model) enterBrowserDirectory() {
	selectedRow := m.browserTable.SelectedRow()
	if len(selectedRow) < 2 {
		return
	}
	name := selectedRow[1]

	if strings.HasPrefix(name, "📁") || strings.Contains(name, "..") {
		cleanName := strings.TrimPrefix(name, "📁 ")
		cleanName = strings.TrimSuffix(cleanName, "/")
		if cleanName == ".." {
			m.browserDir = filepath.Dir(m.browserDir)
		} else {
			m.browserDir = filepath.Join(m.browserDir, cleanName)
		}
		m.refreshBrowserTable()
		m.browserTable.GotoTop()
	}
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

	if cleanName == ".." {
		m.browserDir = filepath.Dir(m.browserDir)
		m.refreshBrowserTable()
		return
	}

	fullPath := filepath.Join(m.browserDir, cleanName)
	if isDir {
		m.browserDir = fullPath
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
	files, err := os.ReadDir(m.browserDir)
	if err != nil {
		m.statusMsg = "Error reading dir: " + err.Error()
		return
	}

	var rows []table.Row
	if m.browserDir != "/" {
		rows = append(rows, table.Row{" ", "..", "", ""})
	}

	onlyDirs := m.currentView == SettingsDirView || (m.currentView == SetupView && m.setupStep == 0)

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
			size = fmt.Sprintf("%.1f KB", float64(info.Size())/1024)
			mod = info.ModTime().Format("2006-01-02 15:04")
		}

		selected := " "
		if !onlyDirs {
			if slices.Contains(m.selectedMedia, filepath.Join(m.browserDir, f.Name())) {
				selected = "x"
			}
		}

		rows = append(rows, table.Row{selected, icon + " " + name, size, mod})
	}
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
