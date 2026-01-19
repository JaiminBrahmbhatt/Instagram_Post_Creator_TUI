package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#FAFAFA")
	ColorDarkGray  = lipgloss.Color("#353533")
	ColorError     = lipgloss.Color("#FF7644")
	ColorSuccess   = lipgloss.Color("#22C55E")
	ColorAccent    = lipgloss.Color("#EE6FF8")

	// Base Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(ColorPrimary).
			Padding(0, 1)

	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	PathStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Background(ColorDarkGray).
			Padding(0, 1).
			Bold(true)

	// List Styles
	PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	// Warning/Error Styles
	WarnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2).
			Bold(true)

	StatusMsgStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	SpinnerStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	LoadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	LoadingBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(2, 4).
			Align(lipgloss.Center)

	LogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkGray).
			Padding(1, 2).
			Width(60).
			Height(8)

	LogEntryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#949494")).
			Italic(true)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(2, 4).
			Align(lipgloss.Center)
)

// List Delegate Styles
func NewCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorPrimary).
		Foreground(ColorPrimary).
		PaddingLeft(2)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorPrimary).
		Foreground(ColorPrimary).
		PaddingLeft(2)
	return d
}
