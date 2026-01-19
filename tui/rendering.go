package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderAppShell(content string) string {
	header := AppTitleStyle.Render("Insta Auto-Post")
	
	// Ensure content takes up available space minus header/footer
	// We might need more sophisticated height calculation here later
	
	// For now, just join them
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		lipgloss.NewStyle().Margin(1, 2).Render(content),
	)
}

func (m *Model) renderProcessingView() string {
	spinner := m.spinner.View()

	// Post Status Card
	statusCard := LoadingBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			LoadingStyle.Render(spinner+"  "+m.currentStatus),
			"\n",
			lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Hold tight, we're uploading to Instagram"),
		),
	)

	// Logs monitor
	var logLines []string
	for _, l := range m.lastLogs {
		logLines = append(logLines, LogEntryStyle.Render(l))
	}
	// Ensure it fills the height or at least looks consistent
	for len(logLines) < 8 {
		logLines = append(logLines, "")
	}

	logMonitor := LogBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(Theme.Primary).Render("Activity Log"),
			"\n",
			strings.Join(logLines, "\n"),
		),
	)

	// Layout the whole thing
	mainDisplay := lipgloss.JoinVertical(lipgloss.Center,
		statusCard,
		"\n",
		logMonitor,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainDisplay)
}

func (m *Model) renderSuccessView() string {
	icon := "✅"
	if strings.Contains(m.lastResult, "❌") || strings.Contains(m.lastResult, "Error") {
		icon = "❌"
	}

	successCard := SuccessBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			LoadingStyle.Render(icon+"  "+m.lastResult),
			"\n",
			lipgloss.NewStyle().Foreground(Theme.Subtle).Render("Press any key to continue"),
		),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, successCard)
}

func (m *Model) renderCurrentView() string {
	switch m.currentView {
	case BrowserView:
		return m.viewBrowser()
	case ComposerView:
		return m.viewComposer()
	case DashboardView:
		return m.viewDashboard()
	case MenuView:
		return m.list.View()
	case SchedulerView:
		return "Scheduled Posts & History (q: back)\n\n" + m.table.View()
	case SettingsAuthView:
		return m.viewSettingsAuth()
	case SettingsDirView:
		return m.viewSettingsDir()
	case SettingsView:
		return m.settingsList.View()
	case SetupView:
		return m.viewSetup()
	default:
		return ""
	}
}
