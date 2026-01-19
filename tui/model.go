package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
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
	fp            filepicker.Model
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
	table         table.Model
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

	fp := filepicker.New()
	fp.AllowedTypes = []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"}
	fp.CurrentDirectory, _ = os.Getwd()
	fp.Height = 20

	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 10},
		{Title: "Scheduled At", Width: 20},
		{Title: "Media", Width: 5},
		{Title: "Caption", Width: 40},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

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
	t.SetStyles(s)

	m := Model{
		db:          database,
		list:        l,
		fp:          fp,
		input:       ti,
		table:       t,
		currentView: "menu",
	}

	// Check for first-time setup
	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = "setup"
		m.setupStep = 0
		m.fp.DirAllowed = true
		m.fp.FileAllowed = false
	} else {
		// Ensure absolute path
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		m.photosDir = absDir
		m.fp.CurrentDirectory = absDir
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

func (m *Model) refreshTable() {
	posts, _ := m.db.GetPosts()
	var rows []table.Row
	for _, p := range posts {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			strings.ToUpper(p.Status),
			p.ScheduledAt,
			fmt.Sprintf("%d", p.MediaCount),
			p.Caption,
		})
	}
	m.table.SetRows(rows)
}

func (m Model) Init() tea.Cmd {
	return m.fp.Init()
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
					// We'll check selection after the model update
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
					m.fp.DirAllowed = true
					m.fp.FileAllowed = true
					m.fp.CurrentDirectory = m.photosDir
					m.statusMsg = fmt.Sprintf("Browsing: %s", m.photosDir)
					return m, m.fp.Init()
				case "Scheduled Posts":
					m.refreshTable()
					m.currentView = "scheduler"
				}
			case "browser":
				// We'll check selection after the model update
			case "composer":
				m.caption = m.input.Value()
				// Default to scheduled for now (+0 minutes)
				_, err := m.db.SavePost(m.caption, m.selectedMedia, "+0 minutes", "scheduled")
				if err != nil {
					m.statusMsg = "Error saving post: " + err.Error()
				} else {
					m.statusMsg = "Post scheduled for now!"
					m.selectedMedia = nil // Clear selection after saving
				}
				m.currentView = "menu"
			}
		case "d":
			if m.currentView == "composer" {
				m.caption = m.input.Value()
				_, err := m.db.SavePost(m.caption, m.selectedMedia, "", "draft")
				if err != nil {
					m.statusMsg = "Error saving post: " + err.Error()
				} else {
					m.statusMsg = "Post saved as draft!"
					m.selectedMedia = nil
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
				// We don't use space for selection here as filepicker uses Enter,
				// but we could use it to toggle if we wanted to build custom logic.
			}
		case "c":
			if m.currentView == "browser" {
				if len(m.selectedMedia) > 0 {
					m.currentView = "composer"
					m.input.Focus()
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.fp.SetHeight(msg.Height - v - 10) // Leave space for status
	}

	var cmd tea.Cmd
	switch m.currentView {
	case "menu":
		m.list, cmd = m.list.Update(msg)
	case "browser":
		m.fp, cmd = m.fp.Update(msg)
		if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
			idx := -1
			for i, s := range m.selectedMedia {
				if s == path {
					idx = i
					break
				}
			}
			if idx >= 0 {
				// Remove from selection
				m.selectedMedia = append(m.selectedMedia[:idx], m.selectedMedia[idx+1:]...)
				m.statusMsg = fmt.Sprintf("Removed: %s", filepath.Base(path))
			} else {
				// Add to selection
				m.selectedMedia = append(m.selectedMedia, path)
				m.statusMsg = fmt.Sprintf("Added: %s", filepath.Base(path))
			}
		}
	case "setup":
		if m.setupStep == 0 {
			m.fp, cmd = m.fp.Update(msg)
			if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
				absPath, _ := filepath.Abs(path)
				m.photosDir = absPath
				m.db.SetSetting("photos_dir", absPath)
				m.setupStep = 1
				m.fp.DirAllowed = false
				m.fp.FileAllowed = true
				m.fp.CurrentDirectory = absPath
			}
		} else {
			m.input, cmd = m.input.Update(msg)
		}
	case "composer":
		m.input, cmd = m.input.Update(msg)
	case "scheduler":
		m.table, cmd = m.table.Update(msg)
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
			s = title + "\n\nPick a directory for your photos:\n\n" + m.fp.View() + "\n\n(Enter to select, 'q' to quit)"
		} else {
			s = title + "\n\nAuto Cleanup\n\nWould you like to automatically remove photos after 30 days if they have been posted?\n\n(y/n)"
		}
	case "dashboard":
		s = "Dashboard View (Work in Progress)\n\nPress 'q' to go back."
	case "browser":
		s = "Select Media (Enter to add to selection, 'c' to continue, 'q' to menu)\n\n"
		if len(m.selectedMedia) > 0 {
			s += fmt.Sprintf("Selected (%d): ", len(m.selectedMedia))
			var names []string
			for _, p := range m.selectedMedia {
				names = append(names, filepath.Base(p))
			}
			s += strings.Join(names, ", ") + "\n\n"
		}
		s += m.fp.View()
		if m.statusMsg != "" {
			s += "\n\n" + m.statusMsg
		}
	case "composer":
		s = fmt.Sprintf(
			"Composer\n\nSelected: %d files\n\n%s\n\n(Enter to schedule NOW, 'd' to save draft, 'q' to cancel)",
			len(m.selectedMedia),
			m.input.View(),
		)
	case "scheduler":
		s = "Scheduled Posts & History\n\n" + m.table.View() + "\n\nPress 'q' for menu"
	default:
		s = m.list.View()
		if m.statusMsg != "" {
			s += "\n\n" + m.statusMsg
		}
	}
	return docStyle.Render(s)
}
