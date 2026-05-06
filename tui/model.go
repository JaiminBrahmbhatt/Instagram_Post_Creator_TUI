package tui

import (
	"fmt"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	authEditing      bool
	authFocusIndex   int
	authInputs       []textinput.Model
	ngrokFocusIndex  int
	browserDir       string
	browserTable     table.Model
	client           *api.Client
	currentView      ViewState
	db               *db.Database
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

	// Browser state
	filterMode  bool
	filterQuery string
	sortMode    SortMode

	// AI grouping state
	aiGroups      []api.PhotoGroup
	aiGroupIndex  int
	lastImageCount int // baseline for directory change detection

	authFieldVisible  []bool
	ngrokFieldVisible []bool
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.showSuccess {
			m.showSuccess = false
			return m, nil
		}
		switch {
		case m.currentView == MenuView && key.Matches(msg, Keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case m.currentView != MenuView && key.Matches(msg, Keys.Back):
			return m.handleBackKey()
		case key.Matches(msg, Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case logMsg:
		m.handleLogMsg(msg)
		return m, watchLogsCmd(m.logSub)
	case quotaMsg:
		m.handleQuotaMsg(msg)
		return m, nil
	case aiGroupMsg:
		m.isProcessing = false
		if msg.err != nil {
			m.showSuccess = true
			m.lastResult = "❌ " + msg.err.Error()
		} else if len(msg.groups) == 0 {
			m.showSuccess = true
			m.lastResult = "No distinct groups found — try with more varied photos."
		} else {
			m.aiGroups = msg.groups
			m.aiGroupIndex = 0
			m.currentView = AIGroupView
		}
		return m, nil
	case photosDirPollMsg:
		return m.handlePhotoDirPoll(msg)
	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	cmd := m.updateViewLogic(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showLimitWarn {
		return AppContainerStyle.Render(WarningBoxStyle.Render(fmt.Sprintf(
			"⚠️  PHOTO LIMIT REACHED\n\nYou have %d photos in your directory.\nPlease remove older photos if they are irrelevant or already posted.\n\nPress Enter to continue...",
			m.mediaCount,
		)))
	}

	var content string
	if m.isProcessing {
		content = m.renderProcessingView()
		return AppContainerStyle.Render(content)
	} else if m.showSuccess {
		content = m.renderSuccessView()
		return AppContainerStyle.Render(content)
	} else {
		content = m.renderCurrentView()
	}

	footer := m.viewFooter()
	fullView := content + "\n\n" + footer

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

func (m *Model) handlePhotoDirPoll(msg photosDirPollMsg) (tea.Model, tea.Cmd) {
	prev := m.lastImageCount
	m.lastImageCount = msg.imageCount

	var nextPoll tea.Cmd
	if m.photosDir != "" {
		nextPoll = pollPhotoDirCmd(m.photosDir)
	}

	// Trigger only when count strictly increases (photos added, not removed/error).
	if msg.imageCount > prev && prev >= 0 && !m.isProcessing {
		added := msg.imageCount - prev
		if m.currentView == MenuView {
			m.isProcessing = true
			m.currentStatus = fmt.Sprintf("Detected %d new photo(s) — grouping with Claude...", added)
			return m, tea.Batch(nextPoll, groupPhotosCmd(m.photosDir))
		}
		// User is busy elsewhere — surface a gentle notification.
		m.statusMsg = fmt.Sprintf("✨ %d new photo(s) detected — go to 'AI Group Photos' to group them", added)
	}
	return m, nextPoll
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
	case AIGroupView:
		return m.updateAIGroupView(msg)
	default:
		return nil
	}
}
