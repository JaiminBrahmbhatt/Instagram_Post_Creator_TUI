package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSetupView(msg tea.Msg) tea.Cmd {
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

	// Step 0: directory selection using browserTable
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Enter):
			m.enterBrowserDirectory()
		case key.Matches(keyMsg, Keys.Select):
			absPath, _ := filepath.Abs(m.browserDir)
			m.updatePhotoDir(absPath)
		case key.Matches(keyMsg, Keys.Left):
			m.parentDirectory()
		}
		return nil
	}
	m.browserTable, cmd = m.browserTable.Update(msg)
	return cmd
}
