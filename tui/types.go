package tui

import (
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewState definitions
type ViewState int

const (
	MenuView ViewState = iota
	SetupView
	DashboardView
	BrowserView
	SchedulerView
	SettingsView
	SettingsDirView
	SettingsAuthView
	SettingsNgrokView
	ComposerView
	AIGroupView
)

// List Item
type item struct {
	title, desc string
	path        string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Messages
type quotaMsg struct {
	usage int
	total int
	err   error
}

func fetchQuotaCmd(client *api.Client) func() tea.Msg {
	return func() tea.Msg {
		limit, err := client.GetPublishingLimit()
		if err != nil {
			return quotaMsg{err: err}
		}
		return quotaMsg{
			usage: limit.QuotaUsage,
			total: limit.Config.QuotaTotal,
		}
	}
}

type logMsg string

func watchLogsCmd(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-sub)
	}
}

type aiGroupMsg struct {
	groups []api.PhotoGroup
	err    error
}
