package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#353533")).
			Padding(0, 1).
			Bold(true)
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
	browserTable  table.Model
	browserDir    string
	settingsList  list.Model
	showFullHelp  bool
}

func InitialModel(database *db.Database) Model {
	// Menu items
	menuItems := []list.Item{
		item{title: "Dashboard", desc: "View limits and engagement"},
		item{title: "Media Browser", desc: "Select photos for carousel"},
		item{title: "Scheduled Posts", desc: "Manage your queue"},
		item{title: "Settings", desc: "Configure app settings"},
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
		{Title: "Timestamp (Local)", Width: 25},
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

	bt := table.New(
		table.WithColumns([]table.Column{
			{Title: " ", Width: 3},
			{Title: "Name", Width: 40},
			{Title: "Size", Width: 10},
			{Title: "Modified", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	bt.SetStyles(s)

	// Settings menu
	settingsMenuItems := []list.Item{
		item{title: "Change Photos Directory", desc: "Set the root folder for media browsing"},
		item{title: "Auto Cleanup", desc: "Toggle 30-day post cleanup"},
	}
	sl := list.New(settingsMenuItems, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Settings"
	sl.SetShowHelp(false)

	m := Model{
		db:           database,
		list:         l,
		fp:           fp,
		input:        ti,
		table:        t,
		browserTable: bt,
		settingsList: sl,
		currentView:  "menu",
	}

	// Check for first-time setup
	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = "setup"
		m.setupStep = 0
		m.fp.DirAllowed = true
		m.fp.FileAllowed = false // Only show directories
		m.fp.ShowPermissions = false
		m.fp.AllowedTypes = nil
	} else {
		// Ensure absolute path
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		m.photosDir = absDir
		m.browserDir = absDir
		m.fp.CurrentDirectory = absDir
		database.RunCleanup()
		m.checkMediaCount()
	}

	return m
}

func (m *Model) refreshBrowserTable() {
	files, err := os.ReadDir(m.browserDir)
	if err != nil {
		m.statusMsg = "Error reading dir: " + err.Error()
		return
	}

	var rows []table.Row
	// Add ".." entry if not at root
	if m.browserDir != "/" {
		rows = append(rows, table.Row{" ", "..", "", ""})
	}

	onlyDirs := m.currentView == "settings_dir" || (m.currentView == "setup" && m.setupStep == 0)

	for _, f := range files {
		info, _ := f.Info()
		name := f.Name()
		size := ""
		mod := ""
		icon := "📄"

		if f.IsDir() {
			icon = "📁"
			name = name + "/"
		} else {
			if onlyDirs {
				continue
			}
			// Filter for images like the filepicker did
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				continue
			}
			size = fmt.Sprintf("%.1f KB", float64(info.Size())/1024)
			mod = info.ModTime().Format("2006-01-02 15:04")
		}

		selected := " "
		if !onlyDirs {
			fullPath := filepath.Join(m.browserDir, f.Name())
			for _, s := range m.selectedMedia {
				if s == fullPath {
					selected = "x"
					break
				}
			}
		}

		rows = append(rows, table.Row{
			selected,
			icon + " " + name,
			size,
			mod,
		})
	}
	m.browserTable.SetRows(rows)
}

func (m *Model) checkMediaCount() {
	count, _ := m.db.GetDirMediaCount(m.photosDir)
	m.mediaCount = count
	if count >= 1000 {
		m.showLimitWarn = true
	}
}

func (m *Model) formatLocalTime(utcStr string) string {
	if utcStr == "" || utcStr == "NULL" {
		return "-"
	}
	// Try ISO format first (e.g. 2006-01-02T15:04:05Z)
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		// Fallback to SQLite space format (e.g. 2006-01-02 15:04:05)
		t, err = time.Parse("2006-01-02 15:04:05", utcStr)
		if err != nil {
			return utcStr // Return raw if both fail
		}
	}
	// Convert to local time and include timezone
	return t.Local().Format("2006-01-02 15:04 MST")
}

func (m *Model) refreshTable() {
	posts, _ := m.db.GetPosts()
	var rows []table.Row
	for _, p := range posts {
		timeToShow := ""
		switch p.Status {
		case "draft":
			timeToShow = m.formatLocalTime(p.CreatedAt)
		case "published":
			timeToShow = m.formatLocalTime(p.PublishedAt)
		default:
			timeToShow = m.formatLocalTime(p.ScheduledAt)
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			strings.ToUpper(p.Status),
			timeToShow,
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
			if m.currentView == "settings_dir" {
				m.currentView = "settings"
				return m, nil
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
					m.browserDir = m.photosDir
					m.fp.AllowedTypes = []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"} // Restore image filters
					m.refreshBrowserTable()
					return m, nil
				case "Scheduled Posts":
					m.refreshTable()
					m.currentView = "scheduler"
				case "Settings":
					m.currentView = "settings"
				}
			case "settings":
				it := m.settingsList.SelectedItem()
				if it == nil {
					return m, nil
				}
				selectedItem := it.(item)
				switch selectedItem.title {
				case "Change Photos Directory":
					m.currentView = "settings_dir"
					m.browserDir = m.photosDir
					if m.browserDir == "" {
						m.browserDir, _ = os.Getwd()
					}
					m.refreshBrowserTable()
					m.browserTable.GotoTop()
					return m, nil
				case "Auto Cleanup":
					m.setupStep = 1 // Reuse setup cleanup view
					m.currentView = "setup"
				}
			case "settings_dir":
				// File picker handled in switch below
			case "browser":
				selectedRow := m.browserTable.SelectedRow()
				if len(selectedRow) < 2 {
					return m, nil
				}
				name := selectedRow[1]
				isDir := strings.HasPrefix(name, "📁")
				cleanName := name
				if isDir {
					cleanName = strings.TrimPrefix(name, "📁 ")
					cleanName = strings.TrimSuffix(cleanName, "/")
				} else {
					cleanName = strings.TrimPrefix(name, "📄 ")
				}

				if cleanName == ".." {
					m.browserDir = filepath.Dir(m.browserDir)
					m.refreshBrowserTable()
					return m, nil
				}

				fullPath := filepath.Join(m.browserDir, cleanName)
				if isDir {
					m.browserDir = fullPath
					m.refreshBrowserTable()
					m.browserTable.GotoTop()
				} else {
					// Toggle selection
					idx := -1
					for i, s := range m.selectedMedia {
						if s == fullPath {
							idx = i
							break
						}
					}
					if idx >= 0 {
						m.selectedMedia = append(m.selectedMedia[:idx], m.selectedMedia[idx+1:]...)
					} else {
						m.selectedMedia = append(m.selectedMedia, fullPath)
					}
					m.refreshBrowserTable()
				}
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
		case "s":
			if m.currentView == "settings_dir" || (m.currentView == "setup" && m.setupStep == 0) {
				path := m.fp.CurrentDirectory
				if m.currentView == "settings_dir" {
					path = m.browserDir
				}
				absPath, _ := filepath.Abs(path)
				m.photosDir = absPath
				m.browserDir = absPath
				m.db.SetSetting("photos_dir", absPath)
				if m.currentView == "setup" {
					m.setupStep = 1
					m.fp.DirAllowed = false
					m.fp.FileAllowed = true
					m.fp.CurrentDirectory = absPath
				} else {
					m.currentView = "settings"
					m.statusMsg = "Photo directory updated!"
				}
				return m, nil
			}
		case "?":
			m.showFullHelp = !m.showFullHelp
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		// Account for path (1) + help (1) + status (1) + margins/padding (~4)
		footerHeight := 6
		m.list.SetSize(msg.Width-h, msg.Height-v-footerHeight)
		m.settingsList.SetSize(msg.Width-h, msg.Height-v-footerHeight)
		m.fp.SetHeight(msg.Height - v - footerHeight - 2)
		m.table.SetHeight(msg.Height - v - footerHeight - 4)
	}

	var cmd tea.Cmd
	switch m.currentView {
	case "menu":
		m.list, cmd = m.list.Update(msg)
	case "settings":
		m.settingsList, cmd = m.settingsList.Update(msg)
	case "browser", "settings_dir":
		m.browserTable, cmd = m.browserTable.Update(msg)
		if m.currentView == "settings_dir" && msg != nil {
			// Handle Enter for directory navigation in settings_dir
			if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
				selectedRow := m.browserTable.SelectedRow()
				if len(selectedRow) >= 2 {
					name := selectedRow[1]
					if strings.HasPrefix(name, "📁") || strings.Contains(name, "..") {
						cleanName := strings.TrimPrefix(name, "📁 ")
						cleanName = strings.TrimSuffix(cleanName, "/")

						if cleanName == ".." {
							m.browserDir = filepath.Dir(m.browserDir)
						} else {
							m.browserDir = filepath.Join(m.browserDir, cleanName)
						}
						m.refreshBrowserTable()
						m.browserTable.GotoTop()
						return m, nil
					}
				}
			}
		}
	case "setup":
		if m.setupStep == 0 {
			m.fp, cmd = m.fp.Update(msg)
			if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
				absPath, _ := filepath.Abs(path)
				m.photosDir = absPath
				m.browserDir = absPath
				m.db.SetSetting("photos_dir", absPath)
				if m.currentView == "setup" {
					m.setupStep = 1
					m.fp.DirAllowed = false
					m.fp.FileAllowed = true
					m.fp.CurrentDirectory = absPath
				} else {
					m.currentView = "settings"
					m.statusMsg = "Photo directory updated!"
				}
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
			s = title + "\n\nPick a directory for your photos:\n\n" + m.fp.View()
		} else {
			s = title + "\n\nAuto Cleanup\n\nWould you like to automatically remove photos after 30 days if they have been posted?\n\n(y/n)"
		}
	case "settings":
		s = m.settingsList.View()
	case "settings_dir":
		s = titleStyle.Render("Change Photos Directory") + "\n\n" +
			"Navigation: Enter to open folder • Esc/q: Back to settings\n" +
			"Selection:  Press 's' to select THE CURRENT folder\n\n" +
			"Current: " + m.browserDir + "\n\n" +
			m.browserTable.View()
	case "dashboard":
		s = "Dashboard View (Work in Progress)"
	case "browser":
		s = "Select Media (Enter: toggle selection • c: continue • q: back)\n\n"
		if len(m.selectedMedia) > 0 {
			s += fmt.Sprintf("Selected (%d): ", len(m.selectedMedia))
			var names []string
			for _, p := range m.selectedMedia {
				names = append(names, filepath.Base(p))
			}
			s += strings.Join(names, ", ") + "\n\n"
		}
		s += m.browserTable.View()
	case "composer":
		s = fmt.Sprintf(
			"Composer (Enter: schedule NOW • d: save draft • q: cancel)\n\nSelected: %d files\n\n%s",
			len(m.selectedMedia),
			m.input.View(),
		)
	case "scheduler":
		s = "Scheduled Posts & History (q: back)\n\n" + m.table.View()
	default: // menu
		// We use a custom viewer for the list to control the help
		m.list.SetShowHelp(false)
		s = m.list.View()
	}

	// Bottom sections: Path and Footer
	var footer string

	// Highlighted Path (only if we have a context or in browser)
	currentPath := m.browserDir
	if m.currentView == "setup" && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" {
		footer += pathStyle.Render("📍 "+currentPath) + "\n"
	}

	// Default help menu
	if m.showFullHelp {
		footer += helpStyle.Render("enter: select • c: continue • d: draft • esc/q: back • ?: back")
	} else {
		keys := "↑/k up • ↓/j down • / filter • q quit"
		if m.currentView == "browser" && len(m.selectedMedia) > 0 {
			keys += " • c continue"
		}
		if m.currentView == "settings_dir" || (m.currentView == "setup" && m.setupStep == 0) {
			keys += " • s select"
		}
		keys += " • ? more"
		footer += helpStyle.Render(keys)
	}

	if m.statusMsg != "" {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMsg)
	}

	return docStyle.Render(s + "\n\n" + footer)
}
