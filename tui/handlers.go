package tui

import (
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleBackKey() (tea.Model, tea.Cmd) {
	if m.currentView == MenuView && m.list.FilterState() == list.Filtering {
		return m, nil
	}
	if m.currentView == MenuView || m.currentView == SetupView {
		m.quitting = true
		return m, tea.Quit
	}
	if m.currentView == SettingsDirView || m.currentView == SettingsAuthView {
		m.currentView = SettingsView
		return m, nil
	}
	m.currentView = MenuView
	return m, nil
}

func (m *Model) updateBrowserView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.filterMode {
			switch {
			case key.Matches(keyMsg, Keys.Back):
				m.filterMode = false
				m.filterQuery = ""
				m.refreshBrowserTable()
				return nil
			case keyMsg.Type == tea.KeyBackspace:
				if len(m.filterQuery) > 0 {
					m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
					m.refreshBrowserTable()
				}
				return nil
			case keyMsg.Type == tea.KeyRunes:
				m.filterQuery += string(keyMsg.Runes)
				m.refreshBrowserTable()
				return nil
			}
			return nil
		}

		switch {
		case key.Matches(keyMsg, Keys.Filter):
			m.filterMode = true
			return nil
		case key.Matches(keyMsg, Keys.SortCycle):
			m.sortMode = nextSort(m.sortMode)
			m.refreshBrowserTable()
			return nil
		case key.Matches(keyMsg, Keys.Left):
			m.parentDirectory()
			return nil
		case key.Matches(keyMsg, Keys.Enter):
			m.handleBrowserSelection()
			return nil
		case key.Matches(keyMsg, Keys.Continue) && len(m.selectedMedia) > 0:
			m.currentView = ComposerView
			m.input.Focus()
			return nil
		}
	}

	m.browserTable, cmd = m.browserTable.Update(msg)
	return cmd
}

func (m *Model) updateComposerView(msg tea.Msg) tea.Cmd {
	if m.isProcessing {
		return nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Back):
			m.currentView = MenuView
			return nil
		case key.Matches(keyMsg, Keys.Enter):
			m.savePost(db.StatusScheduled, "+0 minutes", "Post scheduled for now!")
		case key.Matches(keyMsg, Keys.Draft):
			m.savePost(db.StatusDraft, "", "Post saved as draft!")
		}
	}
	return cmd
}

func (m *Model) updateMenuView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Enter) {
		if m.showLimitWarn {
			m.showLimitWarn = false
			return nil
		}

		it := m.list.SelectedItem()
		if it == nil {
			return nil
		}
		selectedItem, ok := it.(item)
		if !ok {
			return nil
		}

		switch selectedItem.title {
		case MenuTitleDashboard:
			m.currentView = DashboardView
			if m.client != nil {
				return fetchQuotaCmd(m.client)
			}
		case MenuTitleMediaBrowser:
			m.checkMediaCount()
			if m.showLimitWarn {
				return nil
			}
			m.currentView = BrowserView
			m.browserDir = m.photosDir
			m.fp.AllowedTypes = api.SupportedExtensions
			m.refreshBrowserTable()
		case MenuTitleScheduledPosts:
			m.refreshTable()
			m.currentView = SchedulerView
		case MenuTitleSettings:
			m.currentView = SettingsView
		}
	}
	return cmd
}
