package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding // ?
	Select      key.Binding // 's' for directory selection
	Continue    key.Binding // 'c'
	Draft       key.Binding // 'd'
	AutoCleanup key.Binding // 'y'/'n'
	Tab         key.Binding // tab
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
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("q", "esc"),
		key.WithHelp("q/esc", "back"),
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
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "next field"),
	),
}

// BrowserKeyMap implements help.KeyMap for the media browser.
type BrowserKeyMap struct {
	KeyMap
	HasSelection bool
}

func (k BrowserKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Help}
	if k.HasSelection {
		kb = append(kb, k.Continue)
	}
	return kb
}

func (k BrowserKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Back, k.Continue, k.Help},
	}
}

// ComposerKeyMap implements help.KeyMap for the post composer.
type ComposerKeyMap struct{ KeyMap }

func (k ComposerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Draft, k.Back, k.Help}
}

func (k ComposerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Enter, k.Draft},
		{k.Back, k.Help},
	}
}

// SetupKeyMap implements help.KeyMap for setup screens.
type SetupKeyMap struct {
	KeyMap
	Step int // 0: dir, 1: cleanup
}

func (k SetupKeyMap) ShortHelp() []key.Binding {
	if k.Step == 0 {
		return []key.Binding{k.Up, k.Down, k.Enter, k.Select, k.Quit}
	}
	return []key.Binding{k.AutoCleanup, k.Quit}
}

func (k SetupKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

// SettingsDirKeyMap for the directory selector in settings.
type SettingsDirKeyMap struct{ KeyMap }

func (k SettingsDirKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Select, k.Back}
}

func (k SettingsDirKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Select, k.Back},
	}
}

// AuthKeyMap for credentials management.
type AuthKeyMap struct{ KeyMap }

func (k AuthKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Back}
}

func (k AuthKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Enter},
		{k.Back},
	}
}
