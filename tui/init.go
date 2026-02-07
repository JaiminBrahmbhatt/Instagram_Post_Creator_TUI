package tui

import (
	"path/filepath"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func InitialModel(database *db.Database, client *api.Client, reportChan chan string, triggerChan chan struct{}) *Model {
	m := &Model{
		authInputs:       NewEnvInputs(),
		browserTable:     NewBrowserTable(),
		client:           client,
		db:               database,
		fp:               NewFilePicker(),
		help:             help.New(),
		input:            NewCaptionInput(),
		altTextInput:       NewAltTextInput(),
		locationIDInput:    NewLocationIDInput(),
		customScheduleInput: NewCustomScheduleInput(),
		userTagsInput:       NewUserTagsInput(),
		shareToFeedInput:    NewShareToFeedInput(),
		coverURLInput:       NewCoverURLInput(),
		thumbOffsetInput:    NewThumbOffsetInput(),
		collaboratorsInput:  NewCollaboratorsInput(),
		audioNameInput:      NewAudioNameInput(),
		postAsStoryInput:    NewPostAsStoryInput(),
		composerViewport:    viewport.New(80, 20),
		ngrokInput:          NewNgrokInput(),
		domainInput:      NewDomainInput(),
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

	dir, _ := database.GetSetting("photos_dir")
	if dir == "" {
		m.currentView = SetupView
		m.setupStep = 0
		m.fp.DirAllowed = true
		m.fp.FileAllowed = false
		m.fp.ShowPermissions = false
		m.fp.AllowedTypes = nil
	} else {
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
