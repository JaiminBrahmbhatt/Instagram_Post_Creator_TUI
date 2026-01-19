package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) viewBrowser() string {
	s := "Select Media (Enter: toggle selection • c: continue • q: back)\n\n"
	if len(m.selectedMedia) > 0 {
		s += fmt.Sprintf("Selected (%d): ", len(m.selectedMedia))
		var names []string
		for _, p := range m.selectedMedia {
			names = append(names, filepath.Base(p))
		}
		s += strings.Join(names, ", ") + "\n\n"
	}
	return s + m.browserTable.View()
}

func (m *Model) viewComposer() string {
	if m.isProcessing {
		return "" // Content is handled by the global isProcessing overlay in View()
	}

	// Check if we just finished
	if !m.isProcessing && len(m.lastLogs) > 0 && strings.Contains(m.lastLogs[len(m.lastLogs)-1], "Successfully") {
		return "Done! Your post is live.\n\nPress 'q' or 'Esc' to return to the main menu."
	}

	return fmt.Sprintf(
		"Composer (Enter: schedule NOW • d: save draft • q: cancel)\n\nSelected: %d files\n\n%s",
		len(m.selectedMedia),
		m.input.View(),
	)
}

func (m *Model) viewDashboard() string {
	title := TitleStyle.Render("Instagram API Limits")
	
	quotaText := fmt.Sprintf("%d / %d posts used", m.quotaUsage, m.quotaTotal)
	if m.quotaTotal == 0 {
		quotaText = "Loading or unavailable..."
	}
	
	usageCard := CardStyle.Render(quotaText)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		usageCard,
		"\n",
		lipgloss.NewStyle().Foreground(Theme.Subtle).Render("(24-hour moving window)"),
		"\n",
		lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Press 'q' to return to menu"),
	)
}

func (m *Model) viewFooter() string {
	var footer string

	// Path Footer
	currentPath := m.browserDir
	if m.currentView == SetupView && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" && (m.currentView == BrowserView || m.currentView == SettingsDirView) {
		footer += PathStyle.Render("📍 "+currentPath) + "\n"
	}

	// Help Footer
	if m.currentView != MenuView && m.currentView != SettingsView {
		var km help.KeyMap
		switch m.currentView {
		case BrowserView:
			km = BrowserKeyMap{KeyMap: Keys, HasSelection: len(m.selectedMedia) > 0}
		case ComposerView:
			km = ComposerKeyMap{KeyMap: Keys}
		case SetupView:
			km = SetupKeyMap{KeyMap: Keys, Step: m.setupStep}
		case SettingsAuthView:
			km = AuthKeyMap{KeyMap: Keys}
		case SettingsDirView:
			km = SettingsDirKeyMap{KeyMap: Keys}
		case DashboardView:
			km = SettingsDirKeyMap{KeyMap: Keys} // Basic help
		}

		if km != nil {
			footer += "\n" + m.help.View(km)
		}
	}

	if m.statusMsg != "" {
		footer += "\n" + StatusMsgStyle.Render(m.statusMsg)
	}

	return footer
}

func (m *Model) cleanLogLine(line string) string {
	// standard Go log format: 2026/01/19 12:28:57 Message (20 chars + space)
	if len(line) > 20 && strings.Contains(line[:20], "/") && strings.Contains(line[:20], ":") {
		return strings.TrimSpace(line[20:])
	}
	return line
}

func (m *Model) viewSettingsDir() string {
	return TitleStyle.Render("Change Photos Directory") + "\n\n" +
		"Navigation: Enter to open folder • Esc/q: Back to settings\n" +
		"Selection:  Press 's' to select THE CURRENT folder\n\n" +
		"Current: " + m.browserDir + "\n\n" +
		m.browserTable.View()
}

func (m *Model) viewSettingsAuth() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Manage API Credentials") + "\n\n")
	b.WriteString("These values are stored SECURELY in your OS Keychain.\n\n")

	for i := range m.authInputs {
		prefix := "  "
		if i == m.authFocusIndex {
			prefix = lipgloss.NewStyle().Foreground(ColorPrimary).Render("> ")
		}
		b.WriteString(prefix + m.authInputs[i].View() + "\n\n")
	}

	if m.authEditing {
		b.WriteString("\n(Tab: switch fields • Enter: next/save • q: cancel)")
	} else {
		b.WriteString("\n(↑/↓: select field • Enter: VIEW & EDIT • q: back)")
	}
	return b.String()
}

func (m *Model) viewSetup() string {
	title := TitleStyle.Render("First Time Setup")
	if m.setupStep == 0 {
		return title + "\n\nPick a directory for your photos:\n\n" + m.fp.View()
	}
	return title + "\n\nAuto Cleanup\n\nWould you like to automatically remove photos after 30 days if they have been posted?\n\n(y/n)"
}
