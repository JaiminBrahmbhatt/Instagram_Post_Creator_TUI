package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSettingsDirView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.browserTable, cmd = m.browserTable.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, Keys.Enter) {
			m.enterBrowserDirectory()
		} else if key.Matches(keyMsg, Keys.Select) {
			absPath, _ := filepath.Abs(m.browserDir)
			return m.updatePhotoDir(absPath)
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
		case SettingsTitleNgrok:
			m.currentView = SettingsNgrokView
			m.ngrokFocusIndex = 0
			m.ngrokInput.SetValue(api.GetNgrokToken())
			m.domainInput.SetValue(api.GetNgrokDomain())
			m.ngrokInput.Focus()
			m.domainInput.Blur()
		case SettingsTitleGrouping:
			m.currentView = SettingsGroupingView
			m.groupingFocusIndex = 0
			// Load current settings from DB
			backend, _ := m.db.GetSetting("grouping_backend")
			if backend == "" {
				backend = api.BackendClaude
			}
			m.groupingBackend = backend
			model, _ := m.db.GetSetting("grouping_model")
			m.groupingModelInput.SetValue(model)
			m.updateGroupingModelPlaceholder()
			ollamaURL, _ := m.db.GetSetting("ollama_base_url")
			if ollamaURL == "" {
				ollamaURL = api.DefaultOllamaURL
			}
			m.groupingURLInput.SetValue(ollamaURL)
			m.groupingModelInput.Blur()
			m.groupingURLInput.Blur()
		case SettingsTitleEnv:
			m.currentView = SettingsAuthView
			m.authFocusIndex = 0
			m.authEditing = false
			// Pre-fill with current values
			m.authInputs[0].SetValue(api.GetCredential("INSTA_ACCESS_TOKEN"))
			m.authInputs[1].SetValue(api.GetCredential("INSTA_IG_ID"))
			m.authInputs[2].SetValue(fmt.Sprintf("%v", m.db.GetConfigBool("dry_run", "DRY_RUN")))

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
			case key.Matches(keyMsg, Keys.ShiftTab):
				m.toggleAuthFieldVisibility(m.authFocusIndex)
			}
			return tea.Batch(cmds...)
		}

		// Editing Mode Logic
		switch {
		case key.Matches(keyMsg, Keys.Back): // Escape/q to stop editing
			m.authEditing = false
			m.resetAuthEchoModes()
			for i := range m.authInputs {
				m.authInputs[i].Blur()
			}
			return nil
		case key.Matches(keyMsg, Keys.ShiftTab):
			m.toggleAuthFieldVisibility(m.authFocusIndex)
		case key.Matches(keyMsg, Keys.Tab):
			m.authFocusIndex = (m.authFocusIndex + 1) % len(m.authInputs)
			for i := range m.authInputs {
				if i == m.authFocusIndex {
					cmds = append(cmds, m.authInputs[i].Focus())
				} else {
					m.authInputs[i].Blur()
					if i < 2 && !m.authFieldVisible[i] {
						m.authInputs[i].EchoMode = textinput.EchoPassword
					} else if i < 2 {
						m.authInputs[i].EchoMode = textinput.EchoNormal
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
					} else {
						m.authInputs[i].Blur()
						if i < 2 && !m.authFieldVisible[i] {
							m.authInputs[i].EchoMode = textinput.EchoPassword
						} else if i < 2 {
							m.authInputs[i].EchoMode = textinput.EchoNormal
						} else {
							m.authInputs[i].EchoMode = textinput.EchoNormal
						}
					}
				}
			} else {
				// Save all values
				token := m.authInputs[0].Value()
				igID := m.authInputs[1].Value()
				dryRunStr := m.authInputs[2].Value()

				if token != "" && igID != "" {
					// Keychain
					api.SetCredential("INSTA_ACCESS_TOKEN", token)
					api.SetCredential("INSTA_IG_ID", igID)

					// DB
					m.db.SetSetting("dry_run", dryRunStr)

					// Update client
					m.client.AccessToken = token
					m.client.IGID = igID

					m.resetAuthEchoModes()
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

func (m *Model) updateSettingsNgrokView(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Back):
			m.ngrokInput.Blur()
			m.domainInput.Blur()
			m.resetNgrokEchoModes()
			m.currentView = SettingsView
			return nil
		case key.Matches(keyMsg, Keys.ShiftTab):
			m.toggleNgrokFieldVisibility(m.ngrokFocusIndex)
			return nil
		case key.Matches(keyMsg, Keys.Tab):
			m.ngrokFocusIndex = (m.ngrokFocusIndex + 1) % 2
			if m.ngrokFocusIndex == 0 {
				cmds = append(cmds, m.ngrokInput.Focus())
				m.domainInput.Blur()
			} else {
				m.ngrokInput.Blur()
				cmds = append(cmds, m.domainInput.Focus())
			}
			return tea.Batch(cmds...)
		case key.Matches(keyMsg, Keys.Enter):
			if m.ngrokFocusIndex == 0 {
				m.ngrokFocusIndex = 1
				m.ngrokInput.Blur()
				cmds = append(cmds, m.domainInput.Focus())
				return tea.Batch(cmds...)
			}
			// Save both
			token := m.ngrokInput.Value()
			domain := m.domainInput.Value()

			err1 := api.SaveNgrokToken(token)
			err2 := api.SaveNgrokDomain(domain)

			if err1 != nil || err2 != nil {
				m.statusMsg = fmt.Sprintf("Error saving: %v %v", err1, err2)
			} else {
				m.statusMsg = "Settings saved! Restart to activate changes."
			}
			m.ngrokInput.Blur()
			m.domainInput.Blur()
			m.resetNgrokEchoModes()
			m.currentView = SettingsView
			return nil
		}
	}

	// Only update inputs if we didn't handle a special key
	var cmd tea.Cmd
	m.ngrokInput, cmd = m.ngrokInput.Update(msg)
	cmds = append(cmds, cmd)
	m.domainInput, cmd = m.domainInput.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// toggleAuthFieldVisibility toggles EchoMode for the given auth field.
// Index 2 (dry_run) is never a secret.
func (m *Model) toggleAuthFieldVisibility(idx int) {
	if idx >= len(m.authFieldVisible) || idx >= 2 {
		return
	}
	m.authFieldVisible[idx] = !m.authFieldVisible[idx]
	if m.authFieldVisible[idx] {
		m.authInputs[idx].EchoMode = textinput.EchoNormal
	} else {
		m.authInputs[idx].EchoMode = textinput.EchoPassword
	}
}

// resetAuthEchoModes hides all auth secret fields.
func (m *Model) resetAuthEchoModes() {
	for i := 0; i < 2 && i < len(m.authInputs); i++ {
		m.authInputs[i].EchoMode = textinput.EchoPassword
		if i < len(m.authFieldVisible) {
			m.authFieldVisible[i] = false
		}
	}
}

// toggleNgrokFieldVisibility toggles EchoMode for the ngrok token field.
// Only index 0 (token) is a secret; domain (index 1) is never hidden.
func (m *Model) toggleNgrokFieldVisibility(idx int) {
	if idx != 0 {
		return
	}
	m.ngrokFieldVisible[0] = !m.ngrokFieldVisible[0]
	if m.ngrokFieldVisible[0] {
		m.ngrokInput.EchoMode = textinput.EchoNormal
	} else {
		m.ngrokInput.EchoMode = textinput.EchoPassword
	}
}

// resetNgrokEchoModes hides the ngrok token field.
func (m *Model) resetNgrokEchoModes() {
	m.ngrokInput.EchoMode = textinput.EchoPassword
	if len(m.ngrokFieldVisible) > 0 {
		m.ngrokFieldVisible[0] = false
	}
}

// updateGroupingModelPlaceholder sets a context-appropriate placeholder on
// the model input based on the current grouping backend.
func (m *Model) updateGroupingModelPlaceholder() {
	if m.groupingBackend == api.BackendOllama {
		m.groupingModelInput.Placeholder = "e.g. gemma3:4b, llava, llava-phi3, moondream"
	} else {
		m.groupingModelInput.Placeholder = "e.g. claude-opus-4-7, claude-sonnet-4-6, claude-haiku-4-5-20251001"
	}
}

// updateSettingsGroupingView handles input for the AI Grouping Backend settings screen.
// Focus 0 = backend toggle (←/→), 1 = model input, 2 = Ollama URL input.
func (m *Model) updateSettingsGroupingView(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, Keys.Back) {
			m.groupingModelInput.Blur()
			m.groupingURLInput.Blur()
			m.currentView = SettingsView
			return nil
		}

		switch m.groupingFocusIndex {
		case 0: // Backend toggle
			switch {
			case key.Matches(keyMsg, Keys.Left), key.Matches(keyMsg, Keys.Right):
				if m.groupingBackend == api.BackendClaude {
					m.groupingBackend = api.BackendOllama
				} else {
					m.groupingBackend = api.BackendClaude
				}
				m.updateGroupingModelPlaceholder()
				return nil
			case key.Matches(keyMsg, Keys.Down), key.Matches(keyMsg, Keys.Tab), key.Matches(keyMsg, Keys.Enter):
				m.groupingFocusIndex = 1
				cmds = append(cmds, m.groupingModelInput.Focus())
				return tea.Batch(cmds...)
			}

		case 1: // Model input
			switch {
			case key.Matches(keyMsg, Keys.Up):
				m.groupingFocusIndex = 0
				m.groupingModelInput.Blur()
				return nil
			case key.Matches(keyMsg, Keys.Tab), key.Matches(keyMsg, Keys.Enter):
				if m.groupingBackend == api.BackendOllama {
					m.groupingFocusIndex = 2
					m.groupingModelInput.Blur()
					cmds = append(cmds, m.groupingURLInput.Focus())
					return tea.Batch(cmds...)
				}
				return m.saveGroupingSettings()
			}

		case 2: // Ollama URL input
			switch {
			case key.Matches(keyMsg, Keys.Up):
				m.groupingFocusIndex = 1
				m.groupingURLInput.Blur()
				cmds = append(cmds, m.groupingModelInput.Focus())
				return tea.Batch(cmds...)
			case key.Matches(keyMsg, Keys.Enter):
				return m.saveGroupingSettings()
			}
		}
	}

	// Forward key events to the focused input.
	switch m.groupingFocusIndex {
	case 1:
		var cmd tea.Cmd
		m.groupingModelInput, cmd = m.groupingModelInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		var cmd tea.Cmd
		m.groupingURLInput, cmd = m.groupingURLInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) saveGroupingSettings() tea.Cmd {
	m.db.SetSetting("grouping_backend", m.groupingBackend)
	model := m.groupingModelInput.Value()
	m.db.SetSetting("grouping_model", model)
	if m.groupingBackend == api.BackendOllama {
		url := m.groupingURLInput.Value()
		if url == "" {
			url = api.DefaultOllamaURL
		}
		m.db.SetSetting("ollama_base_url", url)
	}
	m.groupingModelInput.Blur()
	m.groupingURLInput.Blur()
	m.statusMsg = "Grouping settings saved!"
	m.currentView = SettingsView
	return nil
}
