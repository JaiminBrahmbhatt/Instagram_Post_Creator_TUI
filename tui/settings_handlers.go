package tui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
)

func (m *Model) updateSettingsDirView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.browserTable, cmd = m.browserTable.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, Keys.Enter) {
			m.enterBrowserDirectory()
		} else if key.Matches(keyMsg, Keys.Select) {
			// Select current folder
			absPath, _ := filepath.Abs(m.browserDir)
			m.updatePhotoDir(absPath)
		}
	}
	return cmd
}

func (m *Model) updateSettingsView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.settingsList, cmd = m.settingsList.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Enter) {
		it := m.settingsList.SelectedItem()
		if it == nil {
			return nil
		}
		selectedItem, ok := it.(item)
		if !ok {
			return nil
		}

		switch selectedItem.title {
		case SettingsTitlePhotosDir:
			m.currentView = SettingsDirView
			m.browserDir = m.photosDir
			if m.browserDir == "" {
				m.browserDir, _ = os.Getwd()
			}
			m.refreshBrowserTable()
			m.browserTable.GotoTop()
		case SettingsTitleCleanup:
			m.setupStep = 1 // Reuse setup cleanup view
			m.currentView = SetupView
		case SettingsTitleEnv:
			m.currentView = SettingsAuthView
			m.authFocusIndex = 0
			m.authEditing = false
			// Pre-fill with current values
			m.authInputs[0].SetValue(api.GetCredential("INSTA_ACCESS_TOKEN"))
			m.authInputs[1].SetValue(api.GetCredential("INSTA_IG_ID"))
			m.authInputs[2].SetValue(m.db.GetConfig("public_url_prefix", "PUBLIC_URL_PREFIX"))
			for i := range m.authInputs {
				m.authInputs[i].Blur()
				if i < 2 {
					m.authInputs[i].EchoMode = textinput.EchoPassword
				} else {
					m.authInputs[i].EchoMode = textinput.EchoNormal
				}
			}
		}
	}
	return cmd
}

func (m *Model) updateSettingsAuthView(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if !m.authEditing {
			switch {
			case key.Matches(keyMsg, Keys.Up):
				m.authFocusIndex--
				if m.authFocusIndex < 0 {
					m.authFocusIndex = len(m.authInputs) - 1
				}
			case key.Matches(keyMsg, Keys.Down):
				m.authFocusIndex = (m.authFocusIndex + 1) % len(m.authInputs)
			case key.Matches(keyMsg, Keys.Enter):
				m.authEditing = true
				cmds = append(cmds, m.authInputs[m.authFocusIndex].Focus())
				m.authInputs[m.authFocusIndex].EchoMode = textinput.EchoNormal
			}
			return tea.Batch(cmds...)
		}

		// Editing Mode Logic
		switch {
		case key.Matches(keyMsg, Keys.Back): // Escape/q to stop editing
			m.authEditing = false
			for i := range m.authInputs {
				m.authInputs[i].Blur()
				m.authInputs[i].EchoMode = textinput.EchoPassword
			}
			return nil
		case key.Matches(keyMsg, Keys.Tab):
			m.authFocusIndex = (m.authFocusIndex + 1) % len(m.authInputs)
			for i := range m.authInputs {
				if i == m.authFocusIndex {
					cmds = append(cmds, m.authInputs[i].Focus())
					m.authInputs[i].EchoMode = textinput.EchoNormal
				} else {
					m.authInputs[i].Blur()
					if i < 2 {
						m.authInputs[i].EchoMode = textinput.EchoPassword
					} else {
						m.authInputs[i].EchoMode = textinput.EchoNormal
					}
				}
			}
		case key.Matches(keyMsg, Keys.Enter):
			if m.authFocusIndex < len(m.authInputs)-1 {
				m.authFocusIndex++
				for i := range m.authInputs {
					if i == m.authFocusIndex {
						cmds = append(cmds, m.authInputs[i].Focus())
						m.authInputs[i].EchoMode = textinput.EchoNormal
					} else {
						m.authInputs[i].Blur()
						if i < 2 {
							m.authInputs[i].EchoMode = textinput.EchoPassword
						} else {
							m.authInputs[i].EchoMode = textinput.EchoNormal
						}
					}
				}
			} else {
				// Save all values
				token := m.authInputs[0].Value()
				igID := m.authInputs[1].Value()
				urlPrefix := m.authInputs[2].Value()
				dryRunStr := m.authInputs[3].Value()

				if token != "" && igID != "" {
					// Keychain
					api.SetCredential("INSTA_ACCESS_TOKEN", token)
					api.SetCredential("INSTA_IG_ID", igID)
					
					// DB
					m.db.SetSetting("public_url_prefix", urlPrefix)
					m.db.SetSetting("dry_run", dryRunStr)

					// Update client
					m.client.AccessToken = token
					m.client.IGID = igID
					
					m.statusMsg = "Configuration saved!"
					m.currentView = SettingsView
					m.authEditing = false
				} else {
					m.statusMsg = "Error: API Token and IG ID are required"
				}
			}
		}
	}

	for i := range m.authInputs {
		var cmd tea.Cmd
		m.authInputs[i], cmd = m.authInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}
