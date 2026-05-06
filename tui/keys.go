package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Back        key.Binding  // Esc only in sub-views
	Quit        key.Binding  // q + ctrl+c on main menu
	Help        key.Binding  // ?
	Select      key.Binding  // s for directory selection
	Continue    key.Binding  // c
	Draft       key.Binding  // d
	AutoCleanup key.Binding  // y/n
	Tab         key.Binding  // tab only (next field)
	ShiftTab    key.Binding  // shift+tab only (toggle secret visibility)
	Filter      key.Binding  // / (enter filter mode)
	SortCycle   key.Binding  // s (cycle sort in browser)
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "up dir"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Select: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "pick dir"),
	),
	Continue: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "continue"),
	),
	Draft: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "save draft"),
	),
	AutoCleanup: key.NewBinding(
		key.WithKeys("y", "n"),
		key.WithHelp("y/n", "confirm"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "toggle visibility"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	SortCycle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
}

// BrowserKeyMap for the media browser.
type BrowserKeyMap struct {
	KeyMap
	HasSelection bool
}

func (k BrowserKeyMap) ShortHelp() []key.Binding {
	bindings := []key.Binding{k.Up, k.Down, k.Enter, k.Filter, k.SortCycle}
	if k.HasSelection {
		bindings = append(bindings, k.Continue)
	}
	return append(bindings, k.Back)
}

func (k BrowserKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Left},
		{k.Filter, k.SortCycle, k.Continue, k.Back},
	}
}

// ComposerKeyMap for the post composer.
type ComposerKeyMap struct{ KeyMap }

func (k ComposerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Draft, k.Back, k.Help}
}

func (k ComposerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Enter, k.Draft}, {k.Back, k.Help}}
}

// SetupKeyMap for setup screens.
type SetupKeyMap struct {
	KeyMap
	Step int
}

func (k SetupKeyMap) ShortHelp() []key.Binding {
	if k.Step == 0 {
		return []key.Binding{k.Up, k.Down, k.Enter, k.Select, k.Quit}
	}
	return []key.Binding{k.AutoCleanup, k.Quit}
}

func (k SetupKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// SettingsDirKeyMap for directory selector.
type SettingsDirKeyMap struct{ KeyMap }

func (k SettingsDirKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Select, k.Back}
}

func (k SettingsDirKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Enter}, {k.Select, k.Back}}
}

// AuthKeyMap for credentials management.
type AuthKeyMap struct{ KeyMap }

func (k AuthKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShiftTab, k.Enter, k.Back}
}

func (k AuthKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Tab, k.ShiftTab}, {k.Enter, k.Back}}
}

// AIGroupKeyMap for the AI photo grouping view.
type AIGroupKeyMap struct{ KeyMap }

func (k AIGroupKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back}
}

func (k AIGroupKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down}, {k.Enter, k.Back}}
}

// NgrokKeyMap for ngrok configuration.
type NgrokKeyMap struct{ KeyMap }

func (k NgrokKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShiftTab, k.Enter, k.Back}
}

func (k NgrokKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Tab, k.ShiftTab}, {k.Enter, k.Back}}
}

// DashboardKeyMap for the dashboard view.
type DashboardKeyMap struct{ KeyMap }

func (k DashboardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back}
}

func (k DashboardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Back}}
}
