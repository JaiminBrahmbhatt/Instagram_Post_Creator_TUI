package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) checkMediaCount() {
	count, err := m.db.GetDirMediaCount(m.photosDir)
	if err != nil {
		m.statusMsg = "Error counting media: " + err.Error()
		return
	}
	m.mediaCount = count
	if count >= PhotoLimitThreshold {
		m.showLimitWarn = true
	}
}

func (m *Model) formatLocalTime(utcStr string) string {
	if utcStr == "" || utcStr == "NULL" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", utcStr)
		if err != nil {
			return utcStr
		}
	}
	return t.Local().Format("2006-01-02 15:04 MST")
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.help.Width = msg.Width
	h, v := AppContainerStyle.GetFrameSize()
	baseHeight := msg.Height - v - 1

	m.list.SetSize(msg.Width-h, baseHeight)
	m.settingsList.SetSize(msg.Width-h, baseHeight)

	manualFooterHeight := 4
	m.table.SetHeight(msg.Height - v - manualFooterHeight - 4)

	browserHeight := msg.Height - v - 18
	if browserHeight < 5 {
		browserHeight = 5
	}
	m.browserTable.SetHeight(browserHeight)
}

func (m *Model) savePost(status db.PostStatus, scheduleTime, successMsg string) {
	caption := m.input.Value()

	if len(caption) > 2200 {
		m.statusMsg = "Error: Caption exceeds 2200 character limit"
		return
	}

	if len(m.selectedMedia) == 0 {
		m.statusMsg = "Error: No media selected"
		return
	}

	_, err := m.db.SavePost(caption, m.selectedMedia, scheduleTime, status)
	if err != nil {
		m.statusMsg = "Error saving post: " + err.Error()
	} else {
		m.statusMsg = successMsg
		m.selectedMedia = nil
		if status == db.StatusScheduled {
			m.isProcessing = true
			select {
			case m.schedulerTrigger <- struct{}{}:
			default:
			}
			return
		}
	}
	m.currentView = MenuView
}

// groupPhotosCmd scans photosDir for images and asks Claude to cluster them.
func groupPhotosCmd(photosDir string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(photosDir)
		if err != nil {
			return aiGroupMsg{err: fmt.Errorf("reading directory: %w", err)}
		}
		var paths []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			switch strings.ToLower(filepath.Ext(e.Name())) {
			case ".jpg", ".jpeg", ".png", ".gif":
				paths = append(paths, filepath.Join(photosDir, e.Name()))
			}
		}
		if len(paths) == 0 {
			return aiGroupMsg{err: fmt.Errorf("no image files found in %s", photosDir)}
		}
		client := api.NewClaudeClient()
		groups, err := client.GroupPhotos(paths)
		return aiGroupMsg{groups: groups, err: err}
	}
}

func (m *Model) updatePhotoDir(path string) {
	m.photosDir = path
	m.browserDir = path
	m.db.SetSetting("photos_dir", path)
	if m.currentView == SetupView {
		m.setupStep = 1
	} else {
		m.currentView = SettingsView
		m.statusMsg = "Photo directory updated!"
	}
}
