package tui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const watchPollInterval = 8 * time.Second

// pollPhotoDirCmd sleeps for watchPollInterval then returns the current image
// count in dir. The model re-issues this command after every message to keep
// continuous polling alive.
func pollPhotoDirCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(watchPollInterval)
		return photosDirPollMsg{imageCount: countImagesInDir(dir)}
	}
}

// countImagesInDir counts supported image files (not videos, not subdirs) in dir.
func countImagesInDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return -1
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".jpg", ".jpeg", ".png", ".gif":
			n++
		}
	}
	return n
}
