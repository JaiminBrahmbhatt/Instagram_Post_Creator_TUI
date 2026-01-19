package tui

import (
	"os"

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
			{Title: " ", Width: 3},
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
	return ti
}

func NewFilePicker() filepicker.Model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"}
	fp.CurrentDirectory, _ = os.Getwd()
	fp.SetHeight(20)
	return fp
}

func NewMenu() list.Model {
	items := []list.Item{
		item{title: "Dashboard", desc: "View limits and engagement"},
		item{title: "Media Browser", desc: "Select photos for carousel"},
		item{title: "Scheduled Posts", desc: "Manage your queue"},
		item{title: "Settings", desc: "Configure app settings"},
	}

	l := list.New(items, NewCustomDelegate(), 0, 0)
	l.Title = "Insta Auto-Post"
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
		item{title: "Change Photos Directory", desc: "Set the root folder for media browsing"},
		item{title: "Auto Cleanup", desc: "Toggle 30-day post cleanup"},
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
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	return s
}
