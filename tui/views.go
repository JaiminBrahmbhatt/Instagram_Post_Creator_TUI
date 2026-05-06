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

func (m *Model) viewAIGroup() string {
	header := H2Style.Render("AI Photo Groups")

	if len(m.aiGroups) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			BodySecondaryStyle.Render("No groups found. Try with more varied photos or check your ANTHROPIC_API_KEY."))
	}

	// Group list with selection indicator
	var rows []string
	for i, g := range m.aiGroups {
		indicator := "  "
		nameStyle := BodyStyle
		if i == m.aiGroupIndex {
			indicator = lipgloss.NewStyle().Foreground(Theme.Primary).Render("› ")
			nameStyle = lipgloss.NewStyle().Foreground(Theme.Primary).Bold(true)
		}
		badge := BadgeInfoStyle.Render(fmt.Sprintf("%d photos", len(g.Photos)))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			indicator, nameStyle.Render(g.Name), "  ", badge))
	}

	// Detail card for selected group
	sel := m.aiGroups[m.aiGroupIndex]
	var names []string
	for _, p := range sel.Photos {
		names = append(names, "  "+filepath.Base(p))
	}
	detailCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			CardHeaderStyle.Render(sel.Name),
			BodyTertiaryStyle.Render(sel.Reason),
			"",
			BodySecondaryStyle.Render(strings.Join(names, "\n")),
		),
	)

	limitNote := ""
	if len(m.aiGroups) > 0 && len(m.aiGroups[0].Photos) > 0 {
		limitNote = BodyTertiaryStyle.Render(fmt.Sprintf("Analyzed up to %d photos — press Enter to compose with selected group", api.MaxPhotosPerBatch))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		strings.Join(rows, "\n"), "",
		detailCard, "",
		limitNote,
	)
}

func (m *Model) viewBrowser() string {
	header := H2Style.Render("Media Browser")

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

	parts := []string{header}
	if selectionInfo != "" {
		parts = append(parts, selectionInfo)
	}

	if m.filterMode {
		cursor := lipgloss.NewStyle().Foreground(Theme.Primary).Render("▌")
		filterBar := lipgloss.JoinHorizontal(lipgloss.Left,
			BadgeInfoStyle.Render("FILTER"),
			"  ",
			InputFocusedStyle.Render(m.filterQuery+cursor),
		)
		parts = append(parts, filterBar, "")
	}

	parts = append(parts, m.browserTable.View())

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) viewComposer() string {
	header := H2Style.Render("Compose New Post")

	fileCountBadge := BadgeInfoStyle.Render(fmt.Sprintf("%d Files Selected", len(m.selectedMedia)))

	inputLabel := InputLabelStyle.Render("Caption")
	inputBox := m.input.View()
	inputCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			inputLabel,
			inputBox,
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		fileCountBadge,
		"",
		inputCard,
	)
}

func (m *Model) viewDashboard() string {
	header := H2Style.Render("Dashboard")

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

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		quotaCard,
		tunnelCard,
	)
}

func (m *Model) viewFooter() string {
	var elements []string

	currentPath := m.browserDir
	if currentPath != "" && (m.currentView == BrowserView || m.currentView == SettingsDirView) {
		pathLine := CodeStyle.Render("📍 " + currentPath)
		if m.filterQuery != "" {
			pathLine += "  " + BadgeInfoStyle.Render("filter: "+m.filterQuery)
		}
		if m.sortMode != SortNameAsc {
			pathLine += "  " + BodyTertiaryStyle.Render("sort: "+m.sortMode.String())
		}
		elements = append(elements, pathLine)
	}

	if m.statusMsg != "" {
		msgStyle := SuccessStyle
		if strings.HasPrefix(m.statusMsg, "Error") {
			msgStyle = ErrorStyle
		}
		elements = append(elements, msgStyle.Render(m.statusMsg))
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
		case SettingsNgrokView:
			km = NgrokKeyMap{KeyMap: Keys}
		case DashboardView:
			km = DashboardKeyMap{KeyMap: Keys}
		case SchedulerView:
			km = DashboardKeyMap{KeyMap: Keys}
		case AIGroupView:
			km = AIGroupKeyMap{KeyMap: Keys}
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
	header := H2Style.Render("Change Photos Directory")

	helpCard := InfoBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			BodyStyle.Render("Enter: open folder"),
			BodyStyle.Render("s:     select current folder as root"),
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
	header := H2Style.Render("Environment Configuration")

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

		visibleBadge := ""
		if i < 2 && i < len(m.authFieldVisible) && m.authFieldVisible[i] {
			visibleBadge = " " + BadgeWarningStyle.Render("[visible]")
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			indicator,
			labelStyle.Render(labelText),
			" ",
			inputView,
			visibleBadge,
		)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		infoCard,
		"",
		strings.Join(rows, "\n"),
	)
}

func (m *Model) viewSetup() string {
	header := H2Style.Render("First Time Setup")
	if m.setupStep == 0 {
		helpCard := InfoBoxStyle.Render(
			BodySecondaryStyle.Render("Navigate to your photos directory. Press 's' to select the current folder."),
		)
		return lipgloss.JoinVertical(lipgloss.Left,
			header, "",
			helpCard, "",
			BodySecondaryStyle.Render("Current: ")+CodeStyle.Render(m.browserDir),
			"", m.browserTable.View(),
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
	header := H2Style.Render("Ngrok Configuration")

	infoCard := InfoBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			BodySecondaryStyle.Render("Configure your Ngrok Authtoken and optional Static Domain."),
			BodySecondaryStyle.Render("Press shift+tab to toggle token visibility."),
		),
	)

	labels := []string{
		"Ngrok Authtoken",
		"Static Domain (Optional)",
	}

	inputs := []textinput.Model{m.ngrokInput, m.domainInput}

	tokenBadge := ""
	if len(m.ngrokFieldVisible) > 0 && m.ngrokFieldVisible[0] {
		tokenBadge = " " + BadgeWarningStyle.Render("[visible]")
	}

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

		badge := ""
		if i == 0 {
			badge = tokenBadge
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			indicator,
			labelStyle.Render(labelText),
			" ",
			inputView,
			badge,
		)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		infoCard,
		"",
		strings.Join(rows, "\n"),
	)
}
