package tui

import (
	"path/filepath"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func InitialModel(database *db.Database, client *api.Client, reportChan chan string, triggerChan chan struct{}) *Model {
	m := &Model{
		authInputs:       NewEnvInputs(),
		browserTable:     NewBrowserTable(),
		client:           client,
		db:               database,
		help:             help.New(),
		input:            NewCaptionInput(),
		ngrokInput:       NewNgrokInput(),
		domainInput:      NewDomainInput(),
		list:             NewMenu(),
		settingsList:     NewSettingsList(),
		table:            NewPostsTable(),
		logSub:           reportChan,
		schedulerTrigger: triggerChan,
		lastLogs:         []string{},
		authFieldVisible:  make([]bool, 3),
		ngrokFieldVisible: make([]bool, 2),
		spinner:          spinner.New(spinner.WithSpinner(spinner.Points), spinner.WithStyle(SpinnerStyle)),
		currentStatus:    "Sharing your story...",
		tableLimit:       20,
		tableOffset:      0,

		currentView: MenuView,
	}

	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = SetupView
		m.setupStep = 0
	} else {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		m.photosDir = absDir
		m.browserDir = absDir
		database.RunCleanup()
		m.checkMediaCount()
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		watchLogsCmd(m.logSub),
	}
	if m.photosDir != "" {
		m.lastImageCount = countImagesInDir(m.photosDir)
		cmds = append(cmds, pollPhotoDirCmd(m.photosDir))
	}
	return tea.Batch(cmds...)
}
