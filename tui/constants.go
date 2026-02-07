package tui

const (
	MenuTitleDashboard      = "📊 Dashboard"
	MenuTitleMediaBrowser   = "📁 Media Browser"
	MenuTitleScheduledPosts = "📅 Scheduled Posts"
	MenuTitleSettings       = "⚙️  Settings"

	// Settings Sub-Menu Items
	SettingsTitlePhotosDir = "Change Photos Directory"
	SettingsTitleCleanup   = "Auto Cleanup"
	SettingsTitleNgrok     = "Ngrok Configuration"
	SettingsTitleEnv       = "Environment Configuration"
)

// ScheduleOptions define when a post can be scheduled (label + SQLite modifier; empty modifier = custom).
var ScheduleOptions = []struct {
	Label    string
	Modifier string // e.g. "+1 hour"; empty means use custom datetime input
}{
	{"Post now", "+0 minutes"},
	{"In 1 hour", "+1 hour"},
	{"In 3 hours", "+3 hours"},
	{"In 6 hours", "+6 hours"},
	{"In 12 hours", "+12 hours"},
	{"In 1 day", "+1 day"},
	{"Custom date & time", ""},
}
