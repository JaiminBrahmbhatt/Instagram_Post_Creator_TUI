package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaiminb/insta-auto-post/db"
)

var (
	titleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type item struct {
	title, desc string
	path        string
	selected    bool
}

func (i item) Title() string {
	if i.selected {
		return fmt.Sprintf("[x] %s", i.title)
	}
	return fmt.Sprintf("[ ] %s", i.title)
}
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	db            *db.Database
	list          list.Model
	browserList   list.Model
	input         textinput.Model
	currentView   string
	selectedMedia []string
	caption       string
	quitting      bool
	statusMsg     string
}

func InitialModel(database *db.Database) Model {
	// Menu items
	menuItems := []list.Item{
		item{title: "Dashboard", desc: "View limits and engagement"},
		item{title: "Media Browser", desc: "Select photos for carousel"},
		item{title: "Scheduled Posts", desc: "Manage your queue"},
	}

	ti := textinput.New()
	ti.Placeholder = "Write your caption here..."

	l := list.New(menuItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Insta Auto-Post"

	// Browser list
	bl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	bl.Title = "Select Media (Space/Enter to toggle, 'c' to continue)"

	return Model{
		db:          database,
		list:        l,
		browserList: bl,
		input:       ti,
		currentView: "menu",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == "menu" {
				m.quitting = true
				return m, tea.Quit
			}
			m.currentView = "menu"
			return m, nil
		case "enter":
			if m.currentView == "menu" {
				selectedItem := m.list.SelectedItem().(item)
				switch selectedItem.title {
				case "Dashboard":
					m.currentView = "dashboard"
				case "Media Browser":
					m.currentView = "browser"
					return m, m.loadMediaCmd()
				case "Scheduled Posts":
					m.currentView = "scheduler"
				}
			} else if m.currentView == "browser" {
				idx := m.browserList.Index()
				it := m.browserList.Items()[idx].(item)
				it.selected = !it.selected
				m.browserList.SetItem(idx, it)
			} else if m.currentView == "composer" {
				m.caption = m.input.Value()
				_, err := m.db.SavePost(m.caption, m.selectedMedia, "")
				if err != nil {
					m.statusMsg = "Error saving post: " + err.Error()
				} else {
					m.statusMsg = "Post saved as draft!"
				}
				m.currentView = "menu"
			}
		case " ":
			if m.currentView == "browser" {
				idx := m.browserList.Index()
				it := m.browserList.Items()[idx].(item)
				it.selected = !it.selected
				m.browserList.SetItem(idx, it)
			}
		case "c":
			if m.currentView == "browser" {
				var selected []string
				for _, it := range m.browserList.Items() {
					if it.(item).selected {
						selected = append(selected, it.(item).path)
					}
				}
				if len(selected) > 0 {
					m.selectedMedia = selected
					m.currentView = "composer"
					m.input.Focus()
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.browserList.SetSize(msg.Width-h, msg.Height-v)
	case mediaLoadedMsg:
		m.browserList.SetItems(msg)
	}

	var cmd tea.Cmd
	switch m.currentView {
	case "menu":
		m.list, cmd = m.list.Update(msg)
	case "browser":
		m.browserList, cmd = m.browserList.Update(msg)
	case "composer":
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s string
	switch m.currentView {
	case "dashboard":
		s = "Dashboard View (Work in Progress)\n\nPress 'q' to go back."
	case "browser":
		s = m.browserList.View()
	case "composer":
		s = fmt.Sprintf(
			"Composer\n\nSelected: %d files\n\n%s\n\n(Enter to save draft, 'q' to cancel)",
			len(m.selectedMedia),
			m.input.View(),
		)
	case "scheduler":
		s = "Scheduled Posts (Work in Progress)\n\nPress 'q' to go back."
	default:
		s = m.list.View()
		if m.statusMsg != "" {
			s += "\n\n" + m.statusMsg
		}
	}
	return docStyle.Render(s)
}

type mediaLoadedMsg []list.Item

func (m Model) loadMediaCmd() tea.Cmd {
	return func() tea.Msg {
		files, _ := os.ReadDir(".")
		var items []list.Item
		for _, f := range files {
			if !f.IsDir() {
				ext := strings.ToLower(filepath.Ext(f.Name()))
				if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
					items = append(items, item{
						title: f.Name(),
						desc:  "Unposted",
						path:  f.Name(),
					})
				}
			}
		}
		return mediaLoadedMsg(items)
	}
}
