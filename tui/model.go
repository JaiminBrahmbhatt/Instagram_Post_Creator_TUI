package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaiminb/insta-auto-post/api"
	"github.com/jaiminb/insta-auto-post/db"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	// List Styles
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#353533")).
			Padding(0, 1).
			Bold(true)
)

type item struct {
	title, desc string
	path        string
}

func (i item) Title() string       { return i.title }
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
	// ... other fields remain ...
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
	help          help.Model
	client        *api.Client
	quotaUsage    int
	quotaTotal    int
}

func InitialModel(database *db.Database, client *api.Client) Model {
	// Menu items
	menuItems := []list.Item{
		item{title: "Dashboard", desc: "View limits and engagement"},
		item{title: "Media Browser", desc: "Select photos for carousel"},
		item{title: "Scheduled Posts", desc: "Manage your queue"},
		item{title: "Settings", desc: "Configure app settings"},
	}

	ti := textinput.New()
	ti.Placeholder = "Write your caption here..."

	// Setup Fancy List
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingLeft(2)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingLeft(2)

	l := list.New(menuItems, delegate, 0, 0)
	l.Title = "Insta Auto-Post"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keys.Enter,
			Keys.Quit,
		}
	}

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
	sl := list.New(settingsMenuItems, delegate, 0, 0)
	sl.Title = "Settings"
	sl.SetShowHelp(true)
	sl.Styles.Title = titleStyle
	sl.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keys.Enter,
			Keys.Back,
		}
	}

	m := Model{
		db:           database,
		list:         l,
		fp:           fp,
		input:        ti,
		table:        t,
		browserTable: bt,
		settingsList: sl,
		currentView:  "menu",
		client:       client,
		help:         help.New(),
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
		switch {
		case key.Matches(msg, Keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, Keys.Back):
			if m.currentView == "menu" && m.list.FilterState() == list.Filtering {
				break
			}
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
		case key.Matches(msg, Keys.Enter):
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
					if m.client != nil {
						limit, err := m.client.GetPublishingLimit()
						if err == nil {
							m.quotaUsage = limit.QuotaUsage
							m.quotaTotal = limit.Config.QuotaTotal
						} else {
							m.statusMsg = "Error fetching limits: " + err.Error()
						}
					}
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
		case key.Matches(msg, Keys.Draft):
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
		case key.Matches(msg, Keys.AutoCleanup):
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

		case key.Matches(msg, Keys.Continue):
			if m.currentView == "browser" {
				if len(m.selectedMedia) > 0 {
					m.currentView = "composer"
					m.input.Focus()
				}
			}
		case key.Matches(msg, Keys.Select):
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
		case key.Matches(msg, Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		h, v := docStyle.GetFrameSize()
		// Determine heights
		// Wrapper margin/padding + status line (1)
		baseHeight := msg.Height - v - 1

		// List has built-in help, so give it more space
		m.list.SetSize(msg.Width-h, baseHeight)
		m.settingsList.SetSize(msg.Width-h, baseHeight)

		// Others need space for external help footer (~2 lines + status)
		manualFooterHeight := 4
		m.fp.SetHeight(msg.Height - v - manualFooterHeight - 2)
		m.table.SetHeight(msg.Height - v - manualFooterHeight - 4)
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
			if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, Keys.Enter) {
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
		title := titleStyle.Render("Instagram API Limits")
		usage := fmt.Sprintf("%d / %d posts used", m.quotaUsage, m.quotaTotal)
		if m.quotaTotal == 0 {
			usage = "Loading or unavailable..."
		}
		s = fmt.Sprintf("%s\n\n%s\n\n(24-hour moving window)\n\nPress 'q' to return to menu", title, usage)
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
		s = m.list.View()
	}

	// Bottom sections: Path and Footer
	var footer string

	// Highlighted Path (only if we have a context or in browser)
	currentPath := m.browserDir
	if m.currentView == "setup" && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" && (m.currentView == "browser" || m.currentView == "settings_dir") {
		footer += pathStyle.Render("📍 "+currentPath) + "\n"
	}

	// Manual Help Footer (only for views WITHOUT built-in help)
	// Help Footer
	if m.currentView != "menu" && m.currentView != "settings" {
		var km help.KeyMap
		switch m.currentView {
		case "browser":
			km = BrowserKeyMap{KeyMap: Keys, HasSelection: len(m.selectedMedia) > 0}
		case "composer":
			km = ComposerKeyMap{KeyMap: Keys}
		case "setup":
			km = SetupKeyMap{KeyMap: Keys, Step: m.setupStep}
		case "settings_dir":
			km = SettingsDirKeyMap{KeyMap: Keys}
		case "dashboard":
			// Simple back key for dashboard
			km = SettingsDirKeyMap{KeyMap: Keys} // Reuse or make simple
		}

		if km != nil {
			footer += "\n" + m.help.View(km)
		}
	}

	if m.statusMsg != "" {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMsg)
	}

	return docStyle.Render(s + "\n\n" + footer)
}
