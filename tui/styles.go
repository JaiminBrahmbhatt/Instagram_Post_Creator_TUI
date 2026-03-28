package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// DESIGN SYSTEM - Inspired by Claude Code
// ============================================================================

// Spacing Tokens (consistent spacing scale)
const (
	SpaceXS = 1
	SpaceSM = 2
	SpaceMD = 3
	SpaceLG = 4
	SpaceXL = 6

	PhotoLimitThreshold = 1000
)

// Color Palette - Refined for better visual hierarchy
type ColorTheme struct {
	// Primary Colors
	Primary       lipgloss.Color
	PrimaryDim    lipgloss.Color
	PrimaryBright lipgloss.Color

	// Neutral Colors
	Background lipgloss.Color
	Surface    lipgloss.Color
	Border     lipgloss.Color
	BorderDim  lipgloss.Color

	// Text Colors
	TextPrimary   lipgloss.Color
	TextSecondary lipgloss.Color
	TextTertiary  lipgloss.Color
	TextInverse   lipgloss.Color

	// Semantic Colors
	Success    lipgloss.Color
	SuccessDim lipgloss.Color
	Error      lipgloss.Color
	ErrorDim   lipgloss.Color
	Warning    lipgloss.Color
	WarningDim lipgloss.Color
	Info       lipgloss.Color
	InfoDim    lipgloss.Color
}

var Theme = ColorTheme{
	// Primary — Claude Code clay/salmon
	Primary:       lipgloss.Color("#DA7756"),
	PrimaryDim:    lipgloss.Color("#B85A38"),
	PrimaryBright: lipgloss.Color("#E8956D"),

	// Neutrals — near-black dark theme
	Background: lipgloss.Color("#1A1A1A"),
	Surface:    lipgloss.Color("#1C1C1C"),
	Border:     lipgloss.Color("#333333"),
	BorderDim:  lipgloss.Color("#2A2A2A"),

	// Text hierarchy
	TextPrimary:   lipgloss.Color("#E0DEF4"),
	TextSecondary: lipgloss.Color("#A6AEBF"),
	TextTertiary:  lipgloss.Color("#6E7681"),
	TextInverse:   lipgloss.Color("#1A1A1A"),

	// Semantic — unchanged
	Success:    lipgloss.Color("#7DC4A0"),
	SuccessDim: lipgloss.Color("#5DA480"),
	Error:      lipgloss.Color("#E88388"),
	ErrorDim:   lipgloss.Color("#C86368"),
	Warning:    lipgloss.Color("#E5C890"),
	WarningDim: lipgloss.Color("#C5A870"),
	Info:       lipgloss.Color("#7AA2F7"),
	InfoDim:    lipgloss.Color("#5A82D7"),
}

// ============================================================================
// LAYOUT STYLES
// ============================================================================

var (
	// App Shell - Top-level container
	AppContainerStyle = lipgloss.NewStyle().
				Padding(0, SpaceSM)

	// Header - Consistent across all views
	AppHeaderStyle = lipgloss.NewStyle().
			Foreground(Theme.TextInverse).
			Background(Theme.Primary).
			Padding(0, SpaceSM).
			Bold(true)

	// Breadcrumb navigation
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(Theme.TextSecondary).
			MarginTop(SpaceXS).
			MarginBottom(SpaceSM)

	BreadcrumbSeparatorStyle = lipgloss.NewStyle().
					Foreground(Theme.TextTertiary)

	BreadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(Theme.Primary).
				Bold(true)

	// Content container - max width for readability
	ContentContainerStyle = lipgloss.NewStyle().
				MaxWidth(120).
				Padding(SpaceSM, 0)

	// Status bar at bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Theme.TextSecondary).
			Background(Theme.Surface).
			Padding(0, SpaceSM).
			MarginTop(SpaceXS)
)

// ============================================================================
// TYPOGRAPHY STYLES
// ============================================================================

var (
	// Headings
	H1Style = lipgloss.NewStyle().
		Foreground(Theme.TextPrimary).
		Bold(true).
		MarginBottom(SpaceSM)

	H2Style = lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true).
		MarginBottom(SpaceXS)

	H3Style = lipgloss.NewStyle().
		Foreground(Theme.TextSecondary).
		Bold(true)

	// Body text
	BodyStyle = lipgloss.NewStyle().
			Foreground(Theme.TextPrimary)

	BodySecondaryStyle = lipgloss.NewStyle().
				Foreground(Theme.TextSecondary)

	BodyTertiaryStyle = lipgloss.NewStyle().
				Foreground(Theme.TextTertiary)

	// Code/Path
	CodeStyle = lipgloss.NewStyle().
			Foreground(Theme.Info).
			Background(Theme.Surface).
			Padding(0, SpaceXS)

	// Links/Interactive
	LinkStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Underline(true)

	LinkHoverStyle = lipgloss.NewStyle().
			Foreground(Theme.PrimaryBright).
			Underline(true).
			Bold(true)
)

// ============================================================================
// COMPONENT STYLES
// ============================================================================

var (
	// Cards - Elevated surfaces
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border).
			Padding(SpaceSM, SpaceMD).
			Background(Theme.Surface).
			MarginBottom(SpaceSM)

	CardHeaderStyle = lipgloss.NewStyle().
			Foreground(Theme.TextPrimary).
			Bold(true).
			MarginBottom(SpaceXS)

	CardBodyStyle = lipgloss.NewStyle().
			Foreground(Theme.TextSecondary)

	// Badges - Small status indicators
	BadgeStyle = lipgloss.NewStyle().
			Foreground(Theme.TextInverse).
			Background(Theme.Primary).
			Padding(0, SpaceXS).
			Bold(true)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Foreground(Theme.TextInverse).
				Background(Theme.Success).
				Padding(0, SpaceXS).
				Bold(true)

	BadgeErrorStyle = lipgloss.NewStyle().
			Foreground(Theme.TextInverse).
			Background(Theme.Error).
			Padding(0, SpaceXS).
			Bold(true)

	BadgeWarningStyle = lipgloss.NewStyle().
				Foreground(Theme.TextInverse).
				Background(Theme.Warning).
				Padding(0, SpaceXS).
				Bold(true)

	BadgeInfoStyle = lipgloss.NewStyle().
			Foreground(Theme.TextInverse).
			Background(Theme.Info).
			Padding(0, SpaceXS).
			Bold(true)

	// Buttons/Actions
	ButtonStyle = lipgloss.NewStyle().
			Foreground(Theme.TextInverse).
			Background(Theme.Primary).
			Padding(0, SpaceSM).
			Bold(true)

	ButtonSecondaryStyle = lipgloss.NewStyle().
				Foreground(Theme.TextPrimary).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Theme.Border).
				Padding(0, SpaceSM)

	// Inputs
	InputStyle = lipgloss.NewStyle().
			Foreground(Theme.TextPrimary).
			Background(Theme.Surface).
			Padding(SpaceXS, SpaceSM).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(Theme.TextPrimary).
				Background(Theme.Surface).
				Padding(SpaceXS, SpaceSM).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Theme.Primary)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Theme.TextSecondary).
			MarginBottom(SpaceXS)

	InputLabelFocusedStyle = lipgloss.NewStyle().
				Foreground(Theme.Primary).
				Bold(true).
				MarginBottom(SpaceXS)

	// Lists
	ListItemStyle = lipgloss.NewStyle().
			Foreground(Theme.TextPrimary).
			PaddingLeft(SpaceSM)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(Theme.Primary).
				Background(Theme.Surface).
				PaddingLeft(SpaceSM).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(Theme.Primary)

	ListItemDescStyle = lipgloss.NewStyle().
				Foreground(Theme.TextSecondary)

	ListItemDescSelectedStyle = lipgloss.NewStyle().
					Foreground(Theme.TextSecondary)
)

// ============================================================================
// SEMANTIC STYLES
// ============================================================================

var (
	// Success states
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Theme.Success)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Success).
			Padding(SpaceMD, SpaceLG).
			Background(Theme.Surface).
			Align(lipgloss.Center)

	// Error states
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Theme.Error)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Error).
			Padding(SpaceMD, SpaceLG).
			Background(Theme.Surface)

	// Warning states
	WarningStyle = lipgloss.NewStyle().
			Foreground(Theme.Warning)

	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Warning).
			Padding(SpaceMD, SpaceLG).
			Background(Theme.Surface)

	// Info states
	InfoStyle = lipgloss.NewStyle().
			Foreground(Theme.Info)

	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Info).
			Padding(SpaceSM, SpaceMD).
			Background(Theme.Surface)

	// Loading states
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Bold(true)

	LoadingBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Primary).
			Padding(SpaceMD, SpaceLG).
			Background(Theme.Surface).
			Align(lipgloss.Center)

	// Activity Log
	LogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border).
			Padding(SpaceSM, SpaceMD).
			Background(Theme.Background).
			Width(70).
			Height(10)

	LogEntryStyle = lipgloss.NewStyle().
			Foreground(Theme.TextTertiary).
			Italic(true)

	LogHeaderStyle = lipgloss.NewStyle().
			Foreground(Theme.TextPrimary).
			Bold(true).
			MarginBottom(SpaceXS)
)

// ============================================================================
// LIST STYLES
// ============================================================================

var (
	PaginationStyle = list.DefaultStyles().PaginationStyle.
			Foreground(Theme.TextSecondary).
			PaddingLeft(SpaceMD)

	HelpStyle = list.DefaultStyles().HelpStyle.
			Foreground(Theme.TextTertiary).
			PaddingLeft(SpaceMD).
			PaddingBottom(SpaceXS)
)

// ============================================================================
// COMPONENT FACTORIES
// ============================================================================

// NewCustomDelegate creates a styled list delegate
func NewCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Selected item
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(Theme.Primary).
		Background(Theme.Surface).
		PaddingLeft(SpaceSM).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Theme.Primary).
		Bold(true)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(Theme.TextSecondary).
		Background(Theme.Surface).
		PaddingLeft(SpaceSM).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(Theme.Primary)

	// Normal item
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(Theme.TextPrimary).
		PaddingLeft(SpaceSM)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(Theme.TextSecondary).
		PaddingLeft(SpaceSM)

	// Dimmed item
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		Foreground(Theme.TextTertiary).
		PaddingLeft(SpaceSM)

	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(Theme.TextTertiary).
		PaddingLeft(SpaceSM)

	return d
}
