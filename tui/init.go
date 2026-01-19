package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
)

func InitialModel(database *db.Database, client *api.Client, reportChan chan string, triggerChan chan struct{}) *Model {
	m := &Model{
		authInputs:       NewAuthInputs(),
		browserTable:     NewBrowserTable(),
		client:           client,
		db:               database,
		fp:               NewFilePicker(),
		help:             help.New(),
		input:            NewCaptionInput(),
		list:             NewMenu(),
		settingsList:     NewSettingsList(),
		table:            NewPostsTable(),
		logSub:           reportChan,
		schedulerTrigger: triggerChan,
		lastLogs:         []string{},
		spinner:          spinner.New(spinner.WithSpinner(spinner.Points), spinner.WithStyle(SpinnerStyle)),
		currentStatus:    "Sharing your story...",
		tableLimit:       20,
		tableOffset:      0,

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

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fp.Init(),
		m.spinner.Tick,
		watchLogsCmd(m.logSub),
	)
}
