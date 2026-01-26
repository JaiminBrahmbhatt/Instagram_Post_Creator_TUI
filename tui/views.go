package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
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

	header := lipgloss.NewStyle().Bold(true).Foreground(Theme.Primary).Render("Compose New Post")
	
	fileCountBadge := BadgeStyle.Render(fmt.Sprintf("%d Files Selected", len(m.selectedMedia)))

	inputBox := CardStyle.Render(m.input.View())

	helpText := lipgloss.NewStyle().Foreground(Theme.Subtle).Render(
		"Actions:\n" +
		"• Enter: Schedule/Post Now\n" +
		"• d:     Save as Draft\n" +
		"• q:     Cancel",
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"\n",
		fileCountBadge,
		"\n",
		inputBox,
		"\n",
		helpText,
	)
}

func (m *Model) viewDashboard() string {
	title := TitleStyle.Render("Instagram API Limits")

	quotaText := fmt.Sprintf("%d / %d posts used", m.quotaUsage, m.quotaTotal)
	if m.quotaTotal == 0 {
		quotaText = "Loading or unavailable..."
	}

	usageCard := CardStyle.Render(quotaText)

	// Tunnel Status
	tunnelTitle := TitleStyle.Render("Tunnel Status")
	tunnelURL := api.GetTunnelURL()
	status := "Inactive"
	if tunnelURL != "" {
		status = "Active: " + tunnelURL
	}
	tunnelCard := CardStyle.Render(status)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		usageCard,
		"\n",
		tunnelTitle,
		tunnelCard,
		"\n",
		lipgloss.NewStyle().Foreground(Theme.Subtle).Render("(24-hour moving window)"),
		"\n",
		lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Press 'q' to return to menu"),
	)
}

func (m *Model) viewFooter() string {
	var elements []string

	// Path
	currentPath := m.browserDir
	if m.currentView == SetupView && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" && (m.currentView == BrowserView || m.currentView == SettingsDirView) {
		elements = append(elements, PathStyle.Render("📍 "+currentPath))
	}

	// Status Message
	if m.statusMsg != "" {
		elements = append(elements, StatusMsgStyle.Render(m.statusMsg))
	}

	// Help
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
			helpView := m.help.View(km)
			elements = append(elements, lipgloss.NewStyle().MarginTop(1).Render(helpView))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
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
	b.WriteString(TitleStyle.Render("Environment Configuration") + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Keys are stored in Keychain, other settings in local database.") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Changes apply immediately to the next operation.") + "\n\n")

	labels := []string{
		"Instagram Access Token",
		"Instagram IG ID",
		"Dry Run Mode",
	}

	for i := range m.authInputs {
		label := labels[i]
		if i == m.authFocusIndex {
			label = lipgloss.NewStyle().Foreground(Theme.Primary).Bold(true).Render(label)
		} else {
			label = lipgloss.NewStyle().Foreground(Theme.Text).Render(label)
		}

		// Cursor indicator
		indicator := "  "
		if i == m.authFocusIndex {
			indicator = lipgloss.NewStyle().Foreground(Theme.Primary).Render("> ")
		}

		b.WriteString(indicator + label + "\n")
		b.WriteString("  " + m.authInputs[i].View() + "\n\n")
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

func (m *Model) viewSettingsNgrok() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Ngrok Configuration") + "\n\n")
	b.WriteString("Configure your Ngrok Authtoken and optional Static Domain.\n\n")

	// Token
	tokenLabel := "Ngrok Authtoken"
	if m.ngrokFocusIndex == 0 {
		tokenLabel = lipgloss.NewStyle().Foreground(Theme.Primary).Bold(true).Render("> " + tokenLabel)
	} else {
		tokenLabel = "  " + tokenLabel
	}
	b.WriteString(tokenLabel + "\n")
	b.WriteString("  " + m.ngrokInput.View() + "\n\n")

	// Domain
	domainLabel := "Static Domain (Optional)"
	if m.ngrokFocusIndex == 1 {
		domainLabel = lipgloss.NewStyle().Foreground(Theme.Primary).Bold(true).Render("> " + domainLabel)
	} else {
		domainLabel = "  " + domainLabel
	}
	b.WriteString(domainLabel + "\n")
	b.WriteString("  " + m.domainInput.View() + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(Theme.Subtle).Render("(Tab: switch • Enter: next/save • v: toggle visibility • Back: cancel)"))
	return b.String()
}
