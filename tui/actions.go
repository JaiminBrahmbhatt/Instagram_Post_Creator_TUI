package tui

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) checkMediaCount() {
	count, err := m.db.GetDirMediaCount(m.photosDir)
	if err != nil {
		m.statusMsg = "Error counting media: " + err.Error()
		return
	}
	m.mediaCount = count
	if count >= PhotoLimitThreshold {
		m.showLimitWarn = true
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

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.help.Width = msg.Width
	h, v := DocStyle.GetFrameSize()
	baseHeight := msg.Height - v - 1

	m.list.SetSize(msg.Width-h, baseHeight)
	m.settingsList.SetSize(msg.Width-h, baseHeight)

	manualFooterHeight := 4
	m.fp.SetHeight(msg.Height - v - manualFooterHeight - 2)
	m.table.SetHeight(msg.Height - v - manualFooterHeight - 4)
}

// getScheduledAt returns UTC datetime string for the selected schedule (preset or custom), or "" for draft.
func (m *Model) getScheduledAt(status db.PostStatus) (string, error) {
	if status == db.StatusDraft {
		return "", nil
	}
	now := time.Now()
	idx := m.scheduleChoiceIdx
	if idx < 0 || idx >= len(ScheduleOptions) {
		idx = 0
	}
	opt := ScheduleOptions[idx]
	if opt.Modifier != "" {
		// Preset: apply modifier to now
		d := parseModifierToDuration(opt.Modifier)
		t := now.Add(d).UTC()
		return t.Format("2006-01-02 15:04:05"), nil
	}
	// Custom: parse custom input
	raw := strings.TrimSpace(m.customScheduleInput.Value())
	if raw == "" {
		return "", nil
	}
	t, err := parseCustomSchedule(raw, now)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("2006-01-02 15:04:05"), nil
}

func parseModifierToDuration(mod string) time.Duration {
	switch mod {
	case "+0 minutes":
		return 0
	case "+1 hour":
		return time.Hour
	case "+3 hours":
		return 3 * time.Hour
	case "+6 hours":
		return 6 * time.Hour
	case "+12 hours":
		return 12 * time.Hour
	case "+1 day":
		return 24 * time.Hour
	}
	return 0
}

// parseCustomSchedule parses user input like "2026-02-10 14:30" or "tomorrow 9am" in local time.
func parseCustomSchedule(raw string, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return now, nil
	}
	loc := now.Location()
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 3:04 PM",
		"2006-01-02 3:04pm",
		"2 Jan 2006 15:04",
		"2 Jan 2006 3:04 PM",
		"02/01/2006 15:04",
		"01/02/2006 15:04",
		"01-02-2006 15:04",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, raw, loc); err == nil {
			if t.Before(now) {
				return t, errors.New("scheduled time must be in the future")
			}
			return t, nil
		}
	}
	// "tomorrow" or "tomorrow 9am" or "tomorrow 09:00"
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "tomorrow") {
		day := now.AddDate(0, 0, 1)
		rest := strings.TrimSpace(raw[8:])
		if rest == "" {
			return time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, loc), nil
		}
		// "9am" or "09:00" or "14:30"
		if t, err := time.ParseInLocation("3:04 PM", rest, loc); err == nil {
			return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
		}
		if t, err := time.ParseInLocation("3:04pm", rest, loc); err == nil {
			return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
		}
		if t, err := time.ParseInLocation("15:04", rest, loc); err == nil {
			return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
		}
		// "9" or "9am" single number hour
		if num := regexp.MustCompile(`^(\d{1,2})(?:\s*:?\s*(\d{2}))?\s*(am|pm)?$`).FindStringSubmatch(strings.ToLower(rest)); len(num) >= 2 {
			h, _ := strconv.Atoi(num[1])
			m := 0
			if len(num) > 2 && num[2] != "" {
				m, _ = strconv.Atoi(num[2])
			}
			if len(num) > 3 && num[3] == "pm" && h < 12 {
				h += 12
			}
			if len(num) > 3 && num[3] == "am" && h == 12 {
				h = 0
			}
			if h < 0 || h > 23 {
				h = 9
			}
			return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, loc), nil
		}
		return time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, loc), nil
	}
	return time.Time{}, errors.New("unrecognized format: use e.g. 2026-02-10 14:30 or tomorrow 9am")
}

func (m *Model) savePost(status db.PostStatus, successMsg string) {
	m.caption = m.input.Value()

	if len(m.caption) > 2200 {
		m.statusMsg = "Error: Caption exceeds 2200 character limit"
		return
	}

	if len(m.selectedMedia) == 0 {
		m.statusMsg = "Error: No media selected"
		return
	}

	scheduledAt, err := m.getScheduledAt(status)
	if err != nil {
		m.statusMsg = "Error: Invalid custom date/time — use e.g. 2026-02-10 14:30 or tomorrow 9am"
		return
	}
	if status == db.StatusScheduled && scheduledAt == "" {
		m.statusMsg = "Error: Choose a schedule time or enter custom date/time"
		return
	}

	// Post as Story: exactly one image or video (implementation_plan.md §3.2)
	postAsStory := false
	{
		storyRaw := strings.TrimSpace(strings.ToLower(m.postAsStoryInput.Value()))
		postAsStory = storyRaw == "y" || storyRaw == "yes" || storyRaw == "1" || storyRaw == "true"
	}
	if postAsStory && len(m.selectedMedia) != 1 {
		m.statusMsg = "Error: Post as Story requires exactly one image or video"
		return
	}

	// Validate location ID: numeric string (implementation_plan.md §1.2)
	if err := api.ValidateLocationID(strings.TrimSpace(m.locationIDInput.Value())); err != nil {
		m.statusMsg = "Error: " + err.Error()
		return
	}

	// Parse user tags: comma-separated, trim, drop empty (implementation_plan.md §2.2: username format)
	userTagsRaw := strings.TrimSpace(m.userTagsInput.Value())
	var userTags []string
	if userTagsRaw != "" {
		for _, s := range strings.Split(userTagsRaw, ",") {
			s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "@"))
			if s != "" {
				userTags = append(userTags, s)
			}
		}
		if err := api.ValidateUserTags(userTags); err != nil {
			m.statusMsg = "Error: " + err.Error()
			return
		}
	}

	// Alt text: only for images, max 300 chars (implementation_plan.md §3.1)
	altText := strings.TrimSpace(m.altTextInput.Value())
	if altText != "" {
		isImage := len(m.selectedMedia) != 1 || !api.IsVideoPath(m.selectedMedia[0])
		if err := api.ValidateAltText(altText, isImage); err != nil {
			m.statusMsg = "Error: " + err.Error()
			return
		}
	}

	// Share to feed: only used for single video; y/yes/1/true = true
	shareRaw := strings.TrimSpace(strings.ToLower(m.shareToFeedInput.Value()))
	shareToFeed := shareRaw == "y" || shareRaw == "yes" || shareRaw == "1" || shareRaw == "true"

	// Thumb offset: ms (0 if empty or invalid)
	thumbStr := strings.TrimSpace(m.thumbOffsetInput.Value())
	thumbOffset := 0
	if thumbStr != "" {
		if n, e := strconv.Atoi(thumbStr); e == nil && n >= 0 {
			thumbOffset = n
		}
	}

	// Collaborators: comma-separated usernames (same format as user tags)
	collabRaw := strings.TrimSpace(m.collaboratorsInput.Value())
	var collaborators []string
	if collabRaw != "" {
		for _, s := range strings.Split(collabRaw, ",") {
			s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "@"))
			if s != "" {
				collaborators = append(collaborators, s)
			}
		}
		if err := api.ValidateUserTags(collaborators); err != nil {
			m.statusMsg = "Error: Collaborators — " + err.Error()
			return
		}
	}

	options := db.PostOptions{
		AltText:       altText,
		LocationID:    strings.TrimSpace(m.locationIDInput.Value()),
		UserTags:      userTags,
		ShareToFeed:   shareToFeed,
		CoverURL:      strings.TrimSpace(m.coverURLInput.Value()),
		ThumbOffset:   thumbOffset,
		Collaborators: collaborators,
		AudioName:     strings.TrimSpace(m.audioNameInput.Value()),
		PostAsStory:   postAsStory,
	}
	_, err = m.db.SavePost(m.caption, m.selectedMedia, scheduledAt, status, options)
	if err != nil {
		m.statusMsg = "Error saving post: " + err.Error()
	} else {
		m.statusMsg = successMsg
		m.selectedMedia = nil
		m.altTextInput.SetValue("")
		m.locationIDInput.SetValue("")
		m.userTagsInput.SetValue("")
		m.shareToFeedInput.SetValue("")
		m.coverURLInput.SetValue("")
		m.thumbOffsetInput.SetValue("")
		m.collaboratorsInput.SetValue("")
		m.audioNameInput.SetValue("")
		m.postAsStoryInput.SetValue("")
		m.customScheduleInput.SetValue("")
		// Only show "uploading" and trigger scheduler when post is scheduled for now (publishes immediately).
		// For future times (custom or preset like "In 3 hours"), just go back to menu; scheduler will publish when time comes.
		if status == db.StatusScheduled && m.scheduleChoiceIdx == 0 {
			m.isProcessing = true
			select {
			case m.schedulerTrigger <- struct{}{}:
			default:
			}
			return
		}
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
