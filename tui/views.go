package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) viewBrowser() string {
	header := H2Style.Render("📁 Media Browser")

	var selectionInfo string
	if len(m.selectedMedia) > 0 {
		selectionBadge := BadgeSuccessStyle.Render(fmt.Sprintf("%d selected", len(m.selectedMedia)))
		var names []string
		for _, p := range m.selectedMedia {
			names = append(names, filepath.Base(p))
		}
		fileList := BodySecondaryStyle.Render(strings.Join(names, ", "))
		selectionInfo = lipgloss.JoinVertical(lipgloss.Left,
			selectionBadge,
			fileList,
			"",
		)
	}

	helpText := BodyTertiaryStyle.Render("Enter: toggle • c: continue • q: back")

	parts := []string{header}
	if selectionInfo != "" {
		parts = append(parts, selectionInfo)
	}
	parts = append(parts, m.browserTable.View(), "", helpText)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) viewComposer() string {
	if m.isProcessing {
		return ""
	}

	if !m.isProcessing && len(m.lastLogs) > 0 && strings.Contains(m.lastLogs[len(m.lastLogs)-1], "Successfully") {
		return SuccessStyle.Render("✅ Done! Your post is live.") + "\n\n" +
			BodySecondaryStyle.Render("Press 'q' or 'Esc' to return to the main menu.")
	}

	header := H2Style.Render("✍️  Compose New Post")

	fileCountBadge := BadgeInfoStyle.Render(fmt.Sprintf("%d Files Selected", len(m.selectedMedia)))

	inputLabel := InputLabelStyle.Render("Caption")
	inputBox := m.input.View()
	inputCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			inputLabel,
			inputBox,
		),
	)

	helpText := BodyTertiaryStyle.Render(
		"Actions:\n" +
			"• Enter: Schedule/Post Now\n" +
			"• d:     Save as Draft\n" +
			"• q:     Cancel",
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		fileCountBadge,
		"",
		inputCard,
		"",
		helpText,
	)
}

func (m *Model) viewDashboard() string {
	header := H2Style.Render("📊 Dashboard")

	quotaText := fmt.Sprintf("%d / %d posts used", m.quotaUsage, m.quotaTotal)
	if m.quotaTotal == 0 {
		quotaText = "Loading or unavailable..."
	}

	quotaCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			CardHeaderStyle.Render("Instagram API Limits"),
			BodyStyle.Render(quotaText),
			BodyTertiaryStyle.Render("(24-hour moving window)"),
		),
	)

	tunnelURL := api.GetTunnelURL()
	status := "Inactive"
	statusStyle := BodySecondaryStyle
	if tunnelURL != "" {
		status = "Active: " + tunnelURL
		statusStyle = SuccessStyle
	}
	tunnelCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			CardHeaderStyle.Render("Tunnel Status"),
			statusStyle.Render(status),
		),
	)

	helpText := BodyTertiaryStyle.Render("Press 'q' to return to menu")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		quotaCard,
		tunnelCard,
		"",
		helpText,
	)
}

func (m *Model) viewFooter() string {
	var elements []string

	currentPath := m.browserDir
	if m.currentView == SetupView && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" && (m.currentView == BrowserView || m.currentView == SettingsDirView) {
		elements = append(elements, CodeStyle.Render("📍 "+currentPath))
	}

	if m.statusMsg != "" {
		elements = append(elements, SuccessStyle.Render(m.statusMsg))
	}

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
			km = SettingsDirKeyMap{KeyMap: Keys}
		}

		if km != nil {
			helpView := m.help.View(km)
			elements = append(elements, lipgloss.NewStyle().MarginTop(1).Render(helpView))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

func (m *Model) cleanLogLine(line string) string {
	if len(line) > 20 && strings.Contains(line[:20], "/") && strings.Contains(line[:20], ":") {
		return strings.TrimSpace(line[20:])
	}
	return line
}

func (m *Model) viewSettingsDir() string {
	header := H2Style.Render("📁 Change Photos Directory")

	helpCard := InfoBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			BodyStyle.Render("Navigation: Enter to open folder"),
			BodyStyle.Render("Selection:  Press 's' to select THE CURRENT folder"),
			BodyTertiaryStyle.Render("Esc/q: Back to settings"),
		),
	)

	currentPath := CodeStyle.Render(m.browserDir)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		helpCard,
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, BodySecondaryStyle.Render("Current: "), currentPath),
		"",
		m.browserTable.View(),
	)
}

func (m *Model) viewSettingsAuth() string {
	header := H2Style.Render("⚙️  Environment Configuration")

	infoCard := InfoBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			BodySecondaryStyle.Render("Keys are stored in Keychain, other settings in local database."),
			BodySecondaryStyle.Render("Changes apply immediately to the next operation."),
		),
	)

	labels := []string{
		"Instagram Access Token",
		"Instagram IG ID",
		"Dry Run Mode",
	}

	var rows []string
	for i := range m.authInputs {
		indicator := "  "
		if i == m.authFocusIndex {
			indicator = lipgloss.NewStyle().Foreground(Theme.Primary).Render("› ")
		}

		var labelStyle lipgloss.Style
		if i == m.authFocusIndex {
			labelStyle = InputLabelFocusedStyle
		} else {
			labelStyle = InputLabelStyle
		}
		labelText := lipgloss.NewStyle().Width(25).Render(labels[i])

		inputView := m.authInputs[i].View()

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			indicator,
			labelStyle.Render(labelText),
			" ",
			inputView,
		)
		rows = append(rows, row)
	}

	var helpText string
	if m.authEditing {
		helpText = BodyTertiaryStyle.Render("Tab: switch • Enter: next/save • q: cancel")
	} else {
		helpText = BodyTertiaryStyle.Render("↑/↓: select • Enter: EDIT • q: back")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		infoCard,
		"",
		strings.Join(rows, "\n"),
		"",
		helpText,
	)
}

func (m *Model) viewSetup() string {
	header := H2Style.Render("🚀 First Time Setup")
	if m.setupStep == 0 {
		helpCard := InfoBoxStyle.Render(
			BodySecondaryStyle.Render("Pick a directory for your photos"),
		)
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			helpCard,
			"",
			m.fp.View(),
		)
	}
	questionCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			CardHeaderStyle.Render("Auto Cleanup"),
			BodyStyle.Render("Would you like to automatically remove photos after 30 days if they have been posted?"),
			"",
			BodyTertiaryStyle.Render("(y/n)"),
		),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		questionCard,
	)
}

func (m *Model) viewSettingsNgrok() string {
	header := H2Style.Render("🔐 Ngrok Configuration")

	infoCard := InfoBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			BodySecondaryStyle.Render("Configure your Ngrok Authtoken and optional Static Domain."),
			BodySecondaryStyle.Render("Press 'v' to toggle token visibility."),
		),
	)

	labels := []string{
		"Ngrok Authtoken",
		"Static Domain (Optional)",
	}

	inputs := []textinput.Model{m.ngrokInput, m.domainInput}

	var rows []string
	for i := 0; i < 2; i++ {
		indicator := "  "
		if i == m.ngrokFocusIndex {
			indicator = lipgloss.NewStyle().Foreground(Theme.Primary).Render("› ")
		}

		var labelStyle lipgloss.Style
		if i == m.ngrokFocusIndex {
			labelStyle = InputLabelFocusedStyle
		} else {
			labelStyle = InputLabelStyle
		}
		labelText := lipgloss.NewStyle().Width(25).Render(labels[i])

		inputView := inputs[i].View()

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			indicator,
			labelStyle.Render(labelText),
			" ",
			inputView,
		)
		rows = append(rows, row)
	}

	helpText := BodyTertiaryStyle.Render("Tab: switch • Enter: next/save • v: toggle visibility • q: back")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		infoCard,
		"",
		strings.Join(rows, "\n"),
		"",
		helpText,
	)
}
