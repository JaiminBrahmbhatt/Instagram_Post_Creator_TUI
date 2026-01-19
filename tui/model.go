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
		return selectedStyle.Render(fmt.Sprintf("[x] %s", i.title))
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
	photosDir     string
	showLimitWarn bool
	setupStep     int // 0: dir, 1: cleanup
	mediaCount    int
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

	m := Model{
		db:          database,
		list:        l,
		browserList: bl,
		input:       ti,
		currentView: "menu",
	}

	// Check for first-time setup
	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = "setup"
		m.setupStep = 0
		m.input.Placeholder = "Enter photos directory path (or leave empty for ./photos)..."
		m.input.Focus()
	} else {
		m.photosDir = dir
		database.RunCleanup()
		m.checkMediaCount()
	}

	return m
}

func (m *Model) checkMediaCount() {
	count, _ := m.db.GetDirMediaCount(m.photosDir)
	m.mediaCount = count
	if count >= 1000 {
		m.showLimitWarn = true
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
			if m.currentView == "menu" || m.currentView == "setup" {
				m.quitting = true
				return m, tea.Quit
			}
			m.currentView = "menu"
			return m, nil
		case "enter":
			if m.showLimitWarn {
				m.showLimitWarn = false
				return m, nil
			}
			switch m.currentView {
			case "setup":
				if m.setupStep == 0 {
					dir := m.input.Value()
					if dir == "" {
						dir = "photos"
					}
					// Create directory if it doesn't exist
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						os.MkdirAll(dir, 0755)
					}
					m.photosDir = dir
					m.db.SetSetting("photos_dir", dir)
					m.setupStep = 1
					m.input.Blur()
					return m, nil
				}
			case "menu":
				it := m.list.SelectedItem()
				if it == nil {
					return m, nil
				}
				selectedItem := it.(item)
				switch selectedItem.title {
				case "Dashboard":
					m.currentView = "dashboard"
				case "Media Browser":
					m.checkMediaCount()
					if m.showLimitWarn {
						return m, nil
					}
					m.currentView = "browser"
					return m, m.loadMediaCmd()
				case "Scheduled Posts":
					m.currentView = "scheduler"
				}
			case "browser":
				items := m.browserList.Items()
				if len(items) == 0 {
					return m, nil
				}
				idx := m.browserList.Index()
				it := items[idx].(item)
				it.selected = !it.selected
				m.browserList.SetItem(idx, it)
			case "composer":
				m.caption = m.input.Value()
				_, err := m.db.SavePost(m.caption, m.selectedMedia, "")
				if err != nil {
					m.statusMsg = "Error saving post: " + err.Error()
				} else {
					m.statusMsg = "Post saved as draft!"
				}
				m.currentView = "menu"
			}
		case "y", "n":
			if m.currentView == "setup" && m.setupStep == 1 {
				val := "0"
				if msg.String() == "y" {
					val = "1"
				}
				m.db.SetSetting("auto_cleanup", val)
				m.currentView = "menu"
				if val == "1" {
					m.db.RunCleanup()
				}
				m.checkMediaCount() // Check now that setup is done
				return m, nil
			}
		case " ":
			if m.currentView == "browser" {
				items := m.browserList.Items()
				if len(items) == 0 {
					return m, nil
				}
				idx := m.browserList.Index()
				it := items[idx].(item)
				it.selected = !it.selected
				m.browserList.SetItem(idx, it)
			}
		case "c":
			if m.currentView == "browser" {
				var selected []string
				for _, i := range m.browserList.Items() {
					if it, ok := i.(item); ok && it.selected {
						selected = append(selected, it.path)
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
	case "composer", "setup":
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showLimitWarn {
		warn := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("202")).
			Padding(1, 2).
			Bold(true)

		return docStyle.Render(warn.Render(fmt.Sprintf(
			"⚠️  PHOTO LIMIT REACHED\n\nYou have %d photos in your directory.\nPlease remove older photos if they are irrelevant or already posted.\n\nPress Enter to continue...",
			m.mediaCount,
		)))
	}

	var s string
	switch m.currentView {
	case "setup":
		title := titleStyle.Render("First Time Setup")
		if m.setupStep == 0 {
			s = title + "\n\n" + m.input.View() + "\n\n(Enter to confirm, 'q' to quit)"
		} else {
			s = title + "\n\nAuto Cleanup\n\nWould you like to automatically remove photos after 30 days if they have been posted?\n\n(y/n)"
		}
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
		files, _ := os.ReadDir(m.photosDir)
		var items []list.Item
		for _, f := range files {
			if !f.IsDir() {
				ext := strings.ToLower(filepath.Ext(f.Name()))
				if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
					fullPath := filepath.Join(m.photosDir, f.Name())
					items = append(items, item{
						title: f.Name(),
						desc:  "Unposted",
						path:  fullPath,
					})
				}
			}
		}
		return mediaLoadedMsg(items)
	}
}
