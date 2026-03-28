package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderAppShell(content string) string {
	header := AppHeaderStyle.Width(m.width).Render("Instagram Auto-Post")
	return AppContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, header, content),
	)
}

func (m *Model) renderBreadcrumb() string {
	var parts []string
	parts = append(parts, BreadcrumbStyle.Render("Home"))

	switch m.currentView {
	case DashboardView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Dashboard"))
	case BrowserView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Media Browser"))
	case ComposerView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbStyle.Render("Media Browser"))
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Compose"))
	case SchedulerView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Scheduled Posts"))
	case SettingsView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Settings"))
	case SettingsDirView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbStyle.Render("Settings"))
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Photos Directory"))
	case SettingsAuthView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbStyle.Render("Settings"))
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Environment"))
	case SettingsNgrokView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbStyle.Render("Settings"))
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Ngrok"))
	case SetupView:
		parts = append(parts, BreadcrumbSeparatorStyle.Render(" › "))
		parts = append(parts, BreadcrumbActiveStyle.Render("Setup"))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func (m *Model) renderProcessingView() string {
	spinner := m.spinner.View()

	statusCard := LoadingBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			LoadingStyle.Render(spinner+"  "+m.currentStatus),
			"\n",
			BodySecondaryStyle.Render("Hold tight, we're uploading to Instagram"),
		),
	)

	var logLines []string
	for _, l := range m.lastLogs {
		logLines = append(logLines, LogEntryStyle.Render(l))
	}
	for len(logLines) < 8 {
		logLines = append(logLines, "")
	}

	logMonitor := LogBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LogHeaderStyle.Render("Activity Log"),
			"\n",
			strings.Join(logLines, "\n"),
		),
	)

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
			BodySecondaryStyle.Render("Press any key to continue"),
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
		header := H2Style.Render("📅 Scheduled Posts & History")
		helpText := BodyTertiaryStyle.Render("q: back")
		return lipgloss.JoinVertical(lipgloss.Left, header, "", m.table.View(), "", helpText)
	case SettingsAuthView:
		return m.viewSettingsAuth()
	case SettingsNgrokView:
		return m.viewSettingsNgrok()
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
