package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func NewBrowserTable() table.Model {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: " ", Width: 4},
			{Title: "Name", Width: 40},
			{Title: "Size", Width: 10},
			{Title: "Modified", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(getTableStyles())
	return t
}

func NewCaptionInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Write your caption here..."
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewMenu() list.Model {
	items := []list.Item{
		item{title: MenuTitleDashboard, desc: "View limits and engagement"},
		item{title: MenuTitleMediaBrowser, desc: "Select photos for carousel"},
		item{title: MenuTitleScheduledPosts, desc: "Manage your queue"},
		item{title: MenuTitleSettings, desc: "Configure app settings"},
	}

	l := list.New(items, NewCustomDelegate(), 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.Styles.Title = H2Style
	l.Styles.PaginationStyle = PaginationStyle
	l.Styles.HelpStyle = HelpStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keys.Enter,
			Keys.Quit,
		}
	}
	return l
}

func NewPostsTable() table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 10},
		{Title: "Timestamp (Local)", Width: 25},
		{Title: "Media", Width: 5},
		{Title: "Caption", Width: 40},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(getTableStyles())
	return t
}

func NewSettingsList() list.Model {
	items := []list.Item{
		item{title: SettingsTitlePhotosDir, desc: "Set the root folder for media browsing"},
		item{title: SettingsTitleCleanup, desc: "Toggle 30-day post cleanup"},
		item{title: SettingsTitleNgrok, desc: "Manage Ngrok Auth Token"},
		item{title: SettingsTitleEnv, desc: "Update API Keys and Dry Run Mode"},
	}
	l := list.New(items, NewCustomDelegate(), 0, 0)
	l.Title = "Settings"
	l.SetShowHelp(true)
	l.Styles.Title = H2Style
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keys.Enter,
			Keys.Back,
		}
	}
	return l
}

func getTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Theme.Border).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(Theme.Background).
		Background(Theme.Primary).
		Bold(false)
	return s
}

func NewEnvInputs() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Instagram Access Token"
	inputs[0].EchoMode = textinput.EchoPassword
	inputs[0].EchoCharacter = '•'
	inputs[0].CharLimit = 512
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Instagram IG ID"
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	inputs[1].CharLimit = 64
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Dry Run (true/false)"
	inputs[2].CharLimit = 5
	inputs[2].Width = 10

	return inputs
}

func NewNgrokInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter Ngrok Auth Token"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 128
	ti.Width = 50
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewDomainInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "e.g. your-domain.ngrok-free.app (Optional)"
	ti.CharLimit = 128
	ti.Width = 50
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}
