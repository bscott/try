package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap contains all key bindings for the TUI
type KeyMap struct {
	// Navigation
	Up   key.Binding
	Down key.Binding

	// Selection
	Select key.Binding

	// Emacs-style cursor movement
	CursorStart   key.Binding
	CursorEnd     key.Binding
	CursorBack    key.Binding
	CursorForward key.Binding

	// Emacs-style editing
	DeleteWordBack key.Binding
	KillLine       key.Binding

	// Actions
	ToggleDelete key.Binding
	Rename       key.Binding
	Create       key.Binding
	Help         key.Binding

	// General
	Cancel    key.Binding
	Backspace key.Binding
}

// NewKeyMap creates a new KeyMap with default bindings
func NewKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑/ctrl+p", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓/ctrl+n", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		CursorStart: key.NewBinding(
			key.WithKeys("ctrl+a", "home"),
			key.WithHelp("ctrl+a", "start of line"),
		),
		CursorEnd: key.NewBinding(
			key.WithKeys("ctrl+e", "end"),
			key.WithHelp("ctrl+e", "end of line"),
		),
		CursorBack: key.NewBinding(
			key.WithKeys("ctrl+b", "left"),
			key.WithHelp("ctrl+b", "back"),
		),
		CursorForward: key.NewBinding(
			key.WithKeys("ctrl+f", "right"),
			key.WithHelp("ctrl+f", "forward"),
		),
		DeleteWordBack: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("ctrl+w", "delete word back"),
		),
		KillLine: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "kill line"),
		),
		ToggleDelete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete mode"),
		),
		Rename: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "rename"),
		),
		Create: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "create new"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+h", "?"),
			key.WithHelp("ctrl+h/?", "help"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
		),
	}
}

// ShortHelp returns a short help string for the current mode
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.ToggleDelete, k.Rename, k.Create, k.Help}
}

// FullHelp returns a full help string for the current mode
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.ToggleDelete, k.Rename, k.Create},
		{k.CursorStart, k.CursorEnd, k.CursorBack, k.CursorForward},
		{k.DeleteWordBack, k.KillLine},
		{k.Cancel, k.Help},
	}
}
