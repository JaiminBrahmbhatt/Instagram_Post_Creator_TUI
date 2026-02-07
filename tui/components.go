package tui

import (
	"os"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/charmbracelet/bubbles/filepicker"
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

func NewAltTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Alt text for accessibility (optional)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewLocationIDInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Facebook Page location ID (optional)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewCustomScheduleInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "e.g. 2026-02-10 14:30 or tomorrow 9am"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewUserTagsInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Comma-separated usernames, e.g. user1, user2"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewShareToFeedInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "y or n (for single video/reel: also show in feed)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewCoverURLInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Reel cover image URL (optional, JPEG 8MB max)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewThumbOffsetInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Thumbnail offset in ms (e.g. 3500 for 3.5s, 0 = not set)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewCollaboratorsInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Reel collaborators (comma-separated usernames)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewAudioNameInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Original audio name for Reels (optional)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewPostAsStoryInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "y or n (post as Story, 24h expiry, single image/video)"
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Theme.TextSecondary)
	return ti
}

func NewFilePicker() filepicker.Model {
	fp := filepicker.New()
	fp.AllowedTypes = api.SupportedExtensions
	fp.CurrentDirectory, _ = os.Getwd()
	fp.SetHeight(20)

	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(Theme.PrimaryBright)
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(Theme.Primary).Bold(true)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(Theme.Info)
	fp.Styles.File = lipgloss.NewStyle().Foreground(Theme.TextPrimary)
	fp.Styles.DisabledFile = lipgloss.NewStyle().Foreground(Theme.TextTertiary)
	fp.Styles.EmptyDirectory = lipgloss.NewStyle().Foreground(Theme.TextSecondary)

	return fp
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
	l.Styles.Title = TitleStyle
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
	l.Styles.Title = TitleStyle
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
