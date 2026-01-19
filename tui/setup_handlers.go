package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSetupView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle Auto-Cleanup Confirmation
	if m.setupStep == 1 {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.AutoCleanup) {
			val := "0"
			if keyMsg.String() == "y" {
				val = "1"
			}
			m.db.SetSetting("auto_cleanup", val)
			m.currentView = MenuView
			if val == "1" {
				m.db.RunCleanup()
			}
			m.checkMediaCount()
		}
		return nil
	}

	// Handle File Picker
	m.fp, cmd = m.fp.Update(msg)
	if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
		absPath, _ := filepath.Abs(path)
		m.updatePhotoDir(absPath)
	}
	// Fallback to Select key for directories if file picker didn't catch it naturally
	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Select) {
		path := m.fp.CurrentDirectory
		absPath, _ := filepath.Abs(path)
		m.updatePhotoDir(absPath)
	}

	return cmd
}
