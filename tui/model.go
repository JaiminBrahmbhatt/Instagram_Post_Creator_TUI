package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
)

type Model struct {
	authEditing      bool
	authFocusIndex   int
	authInputs       []textinput.Model
	ngrokFocusIndex  int
	browserDir       string
	browserTable     table.Model
	caption          string
	client           *api.Client
	currentView      ViewState
	db               *db.Database
	fp               filepicker.Model
	help             help.Model
	input            textinput.Model
	ngrokInput       textinput.Model
	domainInput      textinput.Model
	isProcessing     bool
	lastLogs         []string
	list             list.Model
	logSub           chan string
	schedulerTrigger chan struct{}
	mediaCount       int
	photosDir        string
	quitting         bool
	quotaTotal       int
	quotaUsage       int
	selectedMedia    []string
	settingsList     list.Model
	setupStep        int // 0: dir, 1: cleanup
	showLimitWarn    bool
	spinner          spinner.Model
	statusMsg        string
	table            table.Model
	width            int
	height           int
	showSuccess      bool
	lastResult       string
	currentStatus    string
	tableOffset      int
	tableLimit       int
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global Key Handling
	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.showSuccess {
			m.showSuccess = false
			return m, nil
		}
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
	case logMsg:
		m.handleLogMsg(msg)
		return m, watchLogsCmd(m.logSub)
	case quotaMsg:
		m.handleQuotaMsg(msg)
		return m, nil
	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// View-specific Update Logic
	cmd := m.updateViewLogic(msg)
	return m, cmd
}

func (m *Model) View() string {
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
	if m.isProcessing {
		// Processing view handles its own layout for now (centered)
		// We might want to wrap it in shell later, but for now keep as is or wrap it
		content = m.renderProcessingView()
		return DocStyle.Render(content)
	} else if m.showSuccess {
		content = m.renderSuccessView()
		return DocStyle.Render(content)
	} else {
		content = m.renderCurrentView()
	}

	footer := m.viewFooter()
	fullView := content + "\n\n" + footer
	
	// Wrap in App Shell
	return m.renderAppShell(fullView)
}

func (m *Model) handleLogMsg(msg logMsg) {
	m.lastLogs = append(m.lastLogs, string(msg))
	if len(m.lastLogs) > 8 {
		m.lastLogs = m.lastLogs[len(m.lastLogs)-8:]
	}

	content := string(msg)
	m.currentStatus = m.cleanLogLine(content)
	if strings.Contains(content, "Successfully published") || (strings.Contains(content, "[DRY RUN]") && strings.Contains(content, "complete")) {
		m.isProcessing = false
		m.showSuccess = true
		m.lastResult = "Process complete!"
		if strings.Contains(content, "Successfully published") {
			m.lastResult = "Post published successfully!"
		}
	} else if strings.Contains(content, "❌") {
		m.isProcessing = false
		m.showSuccess = true
		m.lastResult = content
	} else if strings.Contains(content, "Publishing post") {
		m.isProcessing = true
		m.showSuccess = false
	}
}

func (m *Model) handleQuotaMsg(msg quotaMsg) {
	if msg.err != nil {
		m.statusMsg = "Error fetching limits: " + msg.err.Error()
	} else {
		m.quotaUsage = msg.usage
		m.quotaTotal = msg.total
	}
}

func (m *Model) updateViewLogic(msg tea.Msg) tea.Cmd {
	switch m.currentView {
	case BrowserView:
		return m.updateBrowserView(msg)
	case ComposerView:
		return m.updateComposerView(msg)
	case MenuView:
		return m.updateMenuView(msg)
	case SchedulerView:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return cmd
	case SettingsAuthView:
		return m.updateSettingsAuthView(msg)
	case SettingsNgrokView:
		return m.updateSettingsNgrokView(msg)
	case SettingsDirView:
		return m.updateSettingsDirView(msg)
	case SettingsView:
		return m.updateSettingsView(msg)
	case SetupView:
		return m.updateSetupView(msg)
	default:
		return nil
	}
}
