package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Design System: Color Palette
type ColorTheme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Surface    lipgloss.Color
	Text       lipgloss.Color
	Subtle     lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Highlight  lipgloss.Color
}

var Theme = ColorTheme{
	Primary:    lipgloss.Color("#7D56F4"), // Purple
	Secondary:  lipgloss.Color("#EE6FF8"), // Pink
	Background: lipgloss.Color("#1A1B26"), // Deep Blue/Black
	Surface:    lipgloss.Color("#24283B"), // Lighter Blue/Black for cards
	Text:       lipgloss.Color("#C0CAF5"), // White-ish
	Subtle:     lipgloss.Color("#565F89"), // Gray-ish
	Error:      lipgloss.Color("#F7768E"), // Red
	Success:    lipgloss.Color("#9ECE6A"), // Green
	Warning:    lipgloss.Color("#E0AF68"), // Orange
	Highlight:  lipgloss.Color("#BB9AF7"), // Light Purple
}

var (
	// Legacy Color Vars (Mapped to new Theme for backward compatibility)
	ColorPrimary   = Theme.Primary
	ColorSecondary = Theme.Text
	ColorDarkGray  = Theme.Surface
	ColorError     = Theme.Error
	ColorSuccess   = Theme.Success
	ColorAccent    = Theme.Secondary
	ColorLogGray   = Theme.Subtle
	ColorSubtle    = Theme.Subtle

	PhotoLimitThreshold = 1000

	// --- Modern Styles ---

	// App Shell
	AppTitleStyle = lipgloss.NewStyle().
			Foreground(Theme.Background).
			Background(Theme.Primary).
			Padding(0, 1).
			Bold(true).
			MarginLeft(2)

	// Containers
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Subtle).
			Padding(1, 2).
			Background(Theme.Surface)

	// Components
	BadgeStyle = lipgloss.NewStyle().
			Foreground(Theme.Background).
			Background(Theme.Highlight).
			Padding(0, 1).
			Bold(true)

	// --- Legacy Styles (Refined) ---

	TitleStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Bold(true).
			MarginBottom(1)

	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	PathStyle = lipgloss.NewStyle().
			Foreground(Theme.Highlight).
			Background(Theme.Surface).
			Padding(0, 1).
			Bold(true)

	// List Styles
	PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	// Warning/Error Styles
	WarnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Error).
			Padding(1, 2).
			Bold(true)

	StatusMsgStyle = lipgloss.NewStyle().
			Foreground(Theme.Success)

	SpinnerStyle = lipgloss.NewStyle().Foreground(Theme.Secondary)

	LoadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Theme.Primary)

	LoadingBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Primary).
			Padding(2, 4).
			Background(Theme.Surface).
			Align(lipgloss.Center)

	LogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Subtle).
			Padding(1, 2).
			Background(Theme.Background).
			Width(60).
			Height(8)

	LogEntryStyle = lipgloss.NewStyle().
			Foreground(Theme.Subtle).
			Italic(true)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Success).
			Padding(2, 4).
			Background(Theme.Surface).
			Align(lipgloss.Center)
)

// List Delegate Styles
func NewCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Theme.Primary).
		Foreground(Theme.Primary).
		PaddingLeft(2)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Theme.Primary).
		Foreground(Theme.Primary).
		PaddingLeft(2)
	return d
}
