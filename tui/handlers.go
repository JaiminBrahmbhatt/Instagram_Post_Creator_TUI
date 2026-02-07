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
	m.browserTable, cmd = m.browserTable.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Enter):
			m.handleBrowserSelection()
		case key.Matches(keyMsg, Keys.Continue):
			if len(m.selectedMedia) > 0 {
				m.currentView = ComposerView
				m.composerFocusIdx = 0
				m.setComposerFocus()
			}
		}
	}
	return cmd
}

func (m *Model) updateComposerView(msg tea.Msg) tea.Cmd {
	if m.isProcessing {
		return nil
	}

	// Keep viewport content and size in sync; use available height for scrolling
	w := m.width - 4
	if w < 20 {
		w = 20
	}
	h := m.height - 12
	if h < 8 {
		h = 8
	}
	m.composerViewport.Width = w
	m.composerViewport.Height = h
	m.composerViewport.SetContent(m.buildComposerContent())

	// Scroll keys: let viewport handle so user can see all fields
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "down", "pgup", "pgdown":
			var cmd tea.Cmd
			m.composerViewport, cmd = m.composerViewport.Update(msg)
			return cmd
		}
	}

	// Route key events to focused composer field
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Back):
			m.currentView = MenuView
			return nil
		case key.Matches(keyMsg, Keys.Enter):
			msg := "Post scheduled for " + ScheduleOptions[m.scheduleChoiceIdx].Label + "!"
			m.savePost(db.StatusScheduled, msg)
			return nil
		case key.Matches(keyMsg, Keys.Draft):
			m.savePost(db.StatusDraft, "Post saved as draft!")
			return nil
		case key.Matches(keyMsg, Keys.Tab):
			maxIdx := 9 // caption..share to feed, cover, thumb, collab, audio, post as story
			if len(ScheduleOptions) > 0 && ScheduleOptions[m.scheduleChoiceIdx].Modifier == "" {
				maxIdx = 10 // include custom schedule input
			}
			if keyMsg.String() == "shift+tab" {
				m.composerFocusIdx--
				if m.composerFocusIdx < 0 {
					m.composerFocusIdx = maxIdx
				}
			} else {
				m.composerFocusIdx = (m.composerFocusIdx + 1) % (maxIdx + 1)
			}
			m.setComposerFocus()
			return nil
		case keyMsg.String() == "t":
			// Cycle schedule time (Post now → In 1 hour → ... → Custom → Post now)
			m.scheduleChoiceIdx = (m.scheduleChoiceIdx + 1) % len(ScheduleOptions)
			if ScheduleOptions[m.scheduleChoiceIdx].Modifier != "" && m.composerFocusIdx == 10 {
				m.composerFocusIdx = 0
				m.setComposerFocus()
			}
			return nil
		}
	}

	var cmd tea.Cmd
	switch m.composerFocusIdx {
	case 0:
		m.input, cmd = m.input.Update(msg)
	case 1:
		m.altTextInput, cmd = m.altTextInput.Update(msg)
	case 2:
		m.locationIDInput, cmd = m.locationIDInput.Update(msg)
	case 3:
		m.userTagsInput, cmd = m.userTagsInput.Update(msg)
	case 4:
		m.shareToFeedInput, cmd = m.shareToFeedInput.Update(msg)
	case 5:
		m.coverURLInput, cmd = m.coverURLInput.Update(msg)
	case 6:
		m.thumbOffsetInput, cmd = m.thumbOffsetInput.Update(msg)
	case 7:
		m.collaboratorsInput, cmd = m.collaboratorsInput.Update(msg)
	case 8:
		m.audioNameInput, cmd = m.audioNameInput.Update(msg)
	case 9:
		m.postAsStoryInput, cmd = m.postAsStoryInput.Update(msg)
	case 10:
		m.customScheduleInput, cmd = m.customScheduleInput.Update(msg)
	default:
		m.input, cmd = m.input.Update(msg)
	}
	return cmd
}

func (m *Model) setComposerFocus() {
	m.input.Blur()
	m.altTextInput.Blur()
	m.locationIDInput.Blur()
	m.userTagsInput.Blur()
	m.shareToFeedInput.Blur()
	m.coverURLInput.Blur()
	m.thumbOffsetInput.Blur()
	m.collaboratorsInput.Blur()
	m.audioNameInput.Blur()
	m.postAsStoryInput.Blur()
	m.customScheduleInput.Blur()
	switch m.composerFocusIdx {
	case 0:
		m.input.Focus()
	case 1:
		m.altTextInput.Focus()
	case 2:
		m.locationIDInput.Focus()
	case 3:
		m.userTagsInput.Focus()
	case 4:
		m.shareToFeedInput.Focus()
	case 5:
		m.coverURLInput.Focus()
	case 6:
		m.thumbOffsetInput.Focus()
	case 7:
		m.collaboratorsInput.Focus()
	case 8:
		m.audioNameInput.Focus()
	case 9:
		m.postAsStoryInput.Focus()
	case 10:
		m.customScheduleInput.Focus()
	default:
		m.input.Focus()
	}
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
