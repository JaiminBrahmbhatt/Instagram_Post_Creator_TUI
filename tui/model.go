package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

type Model struct {
	authEditing    bool
	authFocusIndex int
	authInputs     []textinput.Model
	browserDir     string
	browserTable   table.Model
	caption        string
	client         *api.Client
	currentView    ViewState
	db             *db.Database
	fp             filepicker.Model
	help           help.Model
	input          textinput.Model
	list           list.Model
	mediaCount     int
	photosDir      string
	quitting       bool
	quotaTotal     int
	quotaUsage     int
	selectedMedia  []string
	settingsList   list.Model
	setupStep      int // 0: dir, 1: cleanup
	showLimitWarn  bool
	statusMsg      string
	table          table.Model
}

func InitialModel(database *db.Database, client *api.Client) Model {
	m := Model{
		authInputs:   NewAuthInputs(),
		browserTable: NewBrowserTable(),
		client:       client,
		db:           database,
		fp:           NewFilePicker(),
		help:         help.New(),
		input:        NewCaptionInput(),
		list:         NewMenu(),
		settingsList: NewSettingsList(),
		table:        NewPostsTable(),

		currentView: MenuView,
	}

	// Check for first-time setup
	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = SetupView
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

func (m Model) Init() tea.Cmd {
	return m.fp.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global Key Handling
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, Keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, Keys.Back):
			return m.handleBackKey()
		case key.Matches(msg, Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}

	// Handle specific messages
	switch msg := msg.(type) {
	case quotaMsg:
		if msg.err != nil {
			m.statusMsg = "Error fetching limits: " + msg.err.Error()
		} else {
			m.quotaUsage = msg.usage
			m.quotaTotal = msg.total
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
		return m, nil
	}

	// View-specific Update Logic
	var cmd tea.Cmd
	switch m.currentView {
	case BrowserView:
		cmd = m.updateBrowserView(msg)
	case ComposerView:
		cmd = m.updateComposerView(msg)
	case DashboardView:
		// No specific update logic unrelated to global keys
	case MenuView:
		cmd = m.updateMenuView(msg)
	case SchedulerView:
		m.table, cmd = m.table.Update(msg)
	case SettingsAuthView:
		cmd = m.updateSettingsAuthView(msg)
	case SettingsDirView:
		cmd = m.updateSettingsDirView(msg)
	case SettingsView:
		cmd = m.updateSettingsView(msg)
	case SetupView:
		cmd = m.updateSetupView(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showLimitWarn {
		return DocStyle.Render(WarnStyle.Render(fmt.Sprintf(
			"⚠️  PHOTO LIMIT REACHED\n\nYou have %d photos in your directory.\nPlease remove older photos if they are irrelevant or already posted.\n\nPress Enter to continue...",
			m.mediaCount,
		)))
	}

	var content string
	switch m.currentView {
	case BrowserView:
		content = m.viewBrowser()
	case ComposerView:
		content = m.viewComposer()
	case DashboardView:
		content = m.viewDashboard()
	case MenuView:
		content = m.list.View()
	case SchedulerView:
		content = "Scheduled Posts & History (q: back)\n\n" + m.table.View()
	case SettingsAuthView:
		content = m.viewSettingsAuth()
	case SettingsDirView:
		content = m.viewSettingsDir()
	case SettingsView:
		content = m.settingsList.View()
	case SetupView:
		content = m.viewSetup()
	}

	return DocStyle.Render(content + "\n\n" + m.viewFooter())
}

// =========================================================================
// Update Helpers
// =========================================================================

func (m *Model) handleBackKey() (tea.Model, tea.Cmd) {
	if m.currentView == MenuView && m.list.FilterState() == list.Filtering {
		return m, nil // Let list handle filtering escape
	}
	if m.currentView == MenuView || m.currentView == SetupView {
		m.quitting = true
		return m, tea.Quit
	}
	if m.currentView == SettingsDirView || m.currentView == SettingsAuthView {
		m.currentView = SettingsView
		return m, nil
	}
	m.currentView = MenuView
	return m, nil
}

func (m *Model) updateBrowserView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.browserTable, cmd = m.browserTable.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Enter):
			m.handleBrowserSelection()
		case key.Matches(keyMsg, Keys.Continue):
			if len(m.selectedMedia) > 0 {
				m.currentView = ComposerView
				m.input.Focus()
			}
		}
	}
	return cmd
}

func (m *Model) updateComposerView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, Keys.Enter): // Schedule
			m.savePost(db.StatusScheduled, "+0 minutes", "Post scheduled for now!")
		case key.Matches(keyMsg, Keys.Draft):
			m.savePost(db.StatusDraft, "", "Post saved as draft!")
		}
	}
	return cmd
}

func (m *Model) updateMenuView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Enter) {
		if m.showLimitWarn {
			m.showLimitWarn = false
			return nil
		}

		it := m.list.SelectedItem()
		if it == nil {
			return nil
		}
		selectedItem, ok := it.(item)
		if !ok {
			return nil
		}

		switch selectedItem.title {
		case "Dashboard":
			m.currentView = DashboardView
			if m.client != nil {
				return fetchQuotaCmd(m.client)
			}
		case "Media Browser":
			m.checkMediaCount()
			if m.showLimitWarn {
				return nil
			}
			m.currentView = BrowserView
			m.browserDir = m.photosDir
			m.fp.AllowedTypes = api.SupportedExtensions
			m.refreshBrowserTable()
		case "Scheduled Posts":
			m.refreshTable()
			m.currentView = SchedulerView
		case "Settings":
			m.currentView = SettingsView
		}
	}
	return cmd
}

func (m *Model) updateSettingsDirView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.browserTable, cmd = m.browserTable.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, Keys.Enter) {
			m.enterBrowserDirectory()
		} else if key.Matches(keyMsg, Keys.Select) {
			// Select current folder
			absPath, _ := filepath.Abs(m.browserDir)
			m.updatePhotoDir(absPath)
		}
	}
	return cmd
}

func (m *Model) updateSettingsView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.settingsList, cmd = m.settingsList.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Enter) {
		it := m.settingsList.SelectedItem()
		if it == nil {
			return nil
		}
		selectedItem, ok := it.(item)
		if !ok {
			return nil
		}

		switch selectedItem.title {
		case "Change Photos Directory":
			m.currentView = SettingsDirView
			m.browserDir = m.photosDir
			if m.browserDir == "" {
				m.browserDir, _ = os.Getwd()
			}
			m.refreshBrowserTable()
			m.browserTable.GotoTop()
		case "Auto Cleanup":
			m.setupStep = 1 // Reuse setup cleanup view
			m.currentView = SetupView
		case "Manage API Credentials":
			m.currentView = SettingsAuthView
			m.authFocusIndex = 0
			m.authEditing = false
			// Pre-fill with current values
			m.authInputs[0].SetValue(api.GetCredential("INSTA_ACCESS_TOKEN"))
			m.authInputs[1].SetValue(api.GetCredential("INSTA_IG_ID"))
			for i := range m.authInputs {
				m.authInputs[i].Blur()
				m.authInputs[i].EchoMode = textinput.EchoPassword
			}
		}
	}
	return cmd
}

func (m *Model) updateSettingsAuthView(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if !m.authEditing {
			switch {
			case key.Matches(keyMsg, Keys.Up):
				m.authFocusIndex--
				if m.authFocusIndex < 0 {
					m.authFocusIndex = len(m.authInputs) - 1
				}
			case key.Matches(keyMsg, Keys.Down):
				m.authFocusIndex = (m.authFocusIndex + 1) % len(m.authInputs)
			case key.Matches(keyMsg, Keys.Enter):
				m.authEditing = true
				cmds = append(cmds, m.authInputs[m.authFocusIndex].Focus())
				m.authInputs[m.authFocusIndex].EchoMode = textinput.EchoNormal
			}
			return tea.Batch(cmds...)
		}

		// Editing Mode Logic
		switch {
		case key.Matches(keyMsg, Keys.Back): // Escape/q to stop editing
			m.authEditing = false
			for i := range m.authInputs {
				m.authInputs[i].Blur()
				m.authInputs[i].EchoMode = textinput.EchoPassword
			}
			return nil
		case key.Matches(keyMsg, Keys.Tab):
			m.authFocusIndex = (m.authFocusIndex + 1) % len(m.authInputs)
			for i := range m.authInputs {
				if i == m.authFocusIndex {
					cmds = append(cmds, m.authInputs[i].Focus())
					m.authInputs[i].EchoMode = textinput.EchoNormal
				} else {
					m.authInputs[i].Blur()
					m.authInputs[i].EchoMode = textinput.EchoPassword
				}
			}
		case key.Matches(keyMsg, Keys.Enter):
			if m.authFocusIndex < len(m.authInputs)-1 {
				m.authFocusIndex++
				for i := range m.authInputs {
					if i == m.authFocusIndex {
						cmds = append(cmds, m.authInputs[i].Focus())
						m.authInputs[i].EchoMode = textinput.EchoNormal
					} else {
						m.authInputs[i].Blur()
						m.authInputs[i].EchoMode = textinput.EchoPassword
					}
				}
			} else {
				// Save credentials
				token := m.authInputs[0].Value()
				igID := m.authInputs[1].Value()

				if token != "" && igID != "" {
					api.SetCredential("INSTA_ACCESS_TOKEN", token)
					api.SetCredential("INSTA_IG_ID", igID)
					// Update client
					m.client.AccessToken = token
					m.client.IGID = igID
					m.statusMsg = "Credentials saved to Keychain!"
					m.currentView = SettingsView
					m.authEditing = false
				} else {
					m.statusMsg = "Error: Both fields are required"
				}
			}
		}
	}

	for i := range m.authInputs {
		var cmd tea.Cmd
		m.authInputs[i], cmd = m.authInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateSetupView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle Auto-Cleanup Confirmation
	if m.setupStep == 1 {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.AutoCleanup) {
			val := "0"
			if keyMsg.String() == "y" {
				val = "1"
			}
			m.db.SetSetting("auto_cleanup", val)
			m.currentView = MenuView
			if val == "1" {
				m.db.RunCleanup()
			}
			m.checkMediaCount()
		}
		return nil
	}

	// Handle File Picker
	m.fp, cmd = m.fp.Update(msg)
	if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
		absPath, _ := filepath.Abs(path)
		m.updatePhotoDir(absPath)
	}
	// Fallback to Select key for directories if file picker didn't catch it naturally
	if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Select) {
		path := m.fp.CurrentDirectory
		absPath, _ := filepath.Abs(path)
		m.updatePhotoDir(absPath)
	}

	return cmd
}

// =========================================================================
// Logic Helpers
// =========================================================================

func (m *Model) checkMediaCount() {
	count, err := m.db.GetDirMediaCount(m.photosDir)
	if err != nil {
		m.statusMsg = "Error counting media: " + err.Error()
		return
	}
	m.mediaCount = count
	if count >= 1000 {
		m.showLimitWarn = true
	}
}

func (m *Model) enterBrowserDirectory() {
	selectedRow := m.browserTable.SelectedRow()
	if len(selectedRow) < 2 {
		return
	}
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
	}
}

func (m *Model) formatLocalTime(utcStr string) string {
	if utcStr == "" || utcStr == "NULL" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", utcStr)
		if err != nil {
			return utcStr
		}
	}
	return t.Local().Format("2006-01-02 15:04 MST")
}

func (m *Model) handleBrowserSelection() {
	selectedRow := m.browserTable.SelectedRow()
	if len(selectedRow) < 2 {
		return
	}
	name := selectedRow[1]
	isDir := strings.HasPrefix(name, "📁")
	cleanName := strings.TrimPrefix(name, "📁 ")
	cleanName = strings.TrimSuffix(cleanName, "/")
	cleanName = strings.TrimPrefix(cleanName, "📄 ")

	if cleanName == ".." {
		m.browserDir = filepath.Dir(m.browserDir)
		m.refreshBrowserTable()
		return
	}

	fullPath := filepath.Join(m.browserDir, cleanName)
	if isDir {
		m.browserDir = fullPath
		m.refreshBrowserTable()
		m.browserTable.GotoTop()
	} else {
		// Toggle selection
		idx := slices.Index(m.selectedMedia, fullPath)
		if idx >= 0 {
			m.selectedMedia = append(m.selectedMedia[:idx], m.selectedMedia[idx+1:]...)
		} else {
			m.selectedMedia = append(m.selectedMedia, fullPath)
		}
		m.refreshBrowserTable()
	}
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
	m.help.Width = msg.Width
	h, v := DocStyle.GetFrameSize()
	baseHeight := msg.Height - v - 1

	m.list.SetSize(msg.Width-h, baseHeight)
	m.settingsList.SetSize(msg.Width-h, baseHeight)

	manualFooterHeight := 4
	m.fp.SetHeight(msg.Height - v - manualFooterHeight - 2)
	m.table.SetHeight(msg.Height - v - manualFooterHeight - 4)
}

func (m *Model) refreshBrowserTable() {
	files, err := os.ReadDir(m.browserDir)
	if err != nil {
		m.statusMsg = "Error reading dir: " + err.Error()
		return
	}

	var rows []table.Row
	if m.browserDir != "/" {
		rows = append(rows, table.Row{" ", "..", "", ""})
	}

	onlyDirs := m.currentView == SettingsDirView || (m.currentView == SetupView && m.setupStep == 0)

	for _, f := range files {
		info, _ := f.Info()
		name := f.Name()
		size, mod, icon := "", "", "📄"

		if f.IsDir() {
			icon = "📁"
			name = name + "/"
		} else {
			if onlyDirs {
				continue
			}
			ext := strings.ToLower(filepath.Ext(name))
			if !slices.Contains(api.SupportedExtensions, ext) {
				continue
			}
			size = fmt.Sprintf("%.1f KB", float64(info.Size())/1024)
			mod = info.ModTime().Format("2006-01-02 15:04")
		}

		selected := " "
		if !onlyDirs {
			if slices.Contains(m.selectedMedia, filepath.Join(m.browserDir, f.Name())) {
				selected = "x"
			}
		}

		rows = append(rows, table.Row{selected, icon + " " + name, size, mod})
	}
	m.browserTable.SetRows(rows)
}

func (m *Model) refreshTable() {
	posts, err := m.db.GetPosts()
	if err != nil {
		m.statusMsg = "Error fetching posts: " + err.Error()
		return
	}
	var rows []table.Row
	for _, p := range posts {
		timeToShow := ""
		switch p.Status {
		case db.StatusDraft:
			timeToShow = m.formatLocalTime(p.CreatedAt)
		case db.StatusPublished:
			timeToShow = m.formatLocalTime(p.PublishedAt)
		default:
			timeToShow = m.formatLocalTime(p.ScheduledAt)
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			strings.ToUpper(string(p.Status)),
			timeToShow,
			fmt.Sprintf("%d", p.MediaCount),
			p.Caption,
		})
	}
	m.table.SetRows(rows)
}

func (m *Model) savePost(status db.PostStatus, scheduleTime, successMsg string) {
	m.caption = m.input.Value()
	_, err := m.db.SavePost(m.caption, m.selectedMedia, scheduleTime, status)
	if err != nil {
		m.statusMsg = "Error saving post: " + err.Error()
	} else {
		m.statusMsg = successMsg
		m.selectedMedia = nil
	}
	m.currentView = MenuView
}

func (m *Model) updatePhotoDir(path string) {
	m.photosDir = path
	m.browserDir = path
	m.db.SetSetting("photos_dir", path)
	if m.currentView == SetupView {
		m.setupStep = 1
		m.fp.DirAllowed = false
		m.fp.FileAllowed = true
		m.fp.CurrentDirectory = path
	} else {
		m.currentView = SettingsView
		m.statusMsg = "Photo directory updated!"
	}
}

// =========================================================================
// View Helpers
// =========================================================================

func (m Model) viewBrowser() string {
	s := "Select Media (Enter: toggle selection • c: continue • q: back)\n\n"
	if len(m.selectedMedia) > 0 {
		s += fmt.Sprintf("Selected (%d): ", len(m.selectedMedia))
		var names []string
		for _, p := range m.selectedMedia {
			names = append(names, filepath.Base(p))
		}
		s += strings.Join(names, ", ") + "\n\n"
	}
	return s + m.browserTable.View()
}

func (m Model) viewComposer() string {
	return fmt.Sprintf(
		"Composer (Enter: schedule NOW • d: save draft • q: cancel)\n\nSelected: %d files\n\n%s",
		len(m.selectedMedia),
		m.input.View(),
	)
}

func (m Model) viewDashboard() string {
	title := TitleStyle.Render("Instagram API Limits")
	usage := fmt.Sprintf("%d / %d posts used", m.quotaUsage, m.quotaTotal)
	if m.quotaTotal == 0 {
		usage = "Loading or unavailable..."
	}
	return fmt.Sprintf("%s\n\n%s\n\n(24-hour moving window)\n\nPress 'q' to return to menu", title, usage)
}

func (m Model) viewFooter() string {
	var footer string

	// Path Footer
	currentPath := m.browserDir
	if m.currentView == SetupView && m.setupStep == 0 {
		currentPath = m.fp.CurrentDirectory
	}
	if currentPath != "" && (m.currentView == BrowserView || m.currentView == SettingsDirView) {
		footer += PathStyle.Render("📍 "+currentPath) + "\n"
	}

	// Help Footer
	if m.currentView != MenuView && m.currentView != SettingsView {
		var km help.KeyMap
		switch m.currentView {
		case BrowserView:
			km = BrowserKeyMap{KeyMap: Keys, HasSelection: len(m.selectedMedia) > 0}
		case ComposerView:
			km = ComposerKeyMap{KeyMap: Keys}
		case SetupView:
			km = SetupKeyMap{KeyMap: Keys, Step: m.setupStep}
		case SettingsAuthView:
			km = AuthKeyMap{KeyMap: Keys}
		case SettingsDirView:
			km = SettingsDirKeyMap{KeyMap: Keys}
		case DashboardView:
			km = SettingsDirKeyMap{KeyMap: Keys} // Basic help
		}

		if km != nil {
			footer += "\n" + m.help.View(km)
		}
	}

	if m.statusMsg != "" {
		footer += "\n" + StatusMsgStyle.Render(m.statusMsg)
	}
	return footer
}

func (m Model) viewSettingsDir() string {
	return TitleStyle.Render("Change Photos Directory") + "\n\n" +
		"Navigation: Enter to open folder • Esc/q: Back to settings\n" +
		"Selection:  Press 's' to select THE CURRENT folder\n\n" +
		"Current: " + m.browserDir + "\n\n" +
		m.browserTable.View()
}

func (m Model) viewSettingsAuth() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Manage API Credentials") + "\n\n")
	b.WriteString("These values are stored SECURELY in your OS Keychain.\n\n")

	for i := range m.authInputs {
		prefix := "  "
		if i == m.authFocusIndex {
			prefix = lipgloss.NewStyle().Foreground(ColorPrimary).Render("> ")
		}
		b.WriteString(prefix + m.authInputs[i].View() + "\n\n")
	}

	if m.authEditing {
		b.WriteString("\n(Tab: switch fields • Enter: next/save • q: cancel)")
	} else {
		b.WriteString("\n(↑/↓: select field • Enter: VIEW & EDIT • q: back)")
	}
	return b.String()
}

func (m Model) viewSetup() string {
	title := TitleStyle.Render("First Time Setup")
	if m.setupStep == 0 {
		return title + "\n\nPick a directory for your photos:\n\n" + m.fp.View()
	}
	return title + "\n\nAuto Cleanup\n\nWould you like to automatically remove photos after 30 days if they have been posted?\n\n(y/n)"
}
