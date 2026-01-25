package tui

import (
	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current mode of the TUI
type Mode int

const (
	ModeNormal Mode = iota
	ModeDelete
	ModeRename
	ModeCreate
	ModeConfirmDelete
)

// String returns a human-readable string for the mode
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeDelete:
		return "DELETE"
	case ModeRename:
		return "RENAME"
	case ModeCreate:
		return "CREATE"
	case ModeConfirmDelete:
		return "CONFIRM DELETE"
	default:
		return "UNKNOWN"
	}
}

// Model represents the state of the Bubble Tea application
type Model struct {
	manager         *tries.Manager
	allEntries      []tries.Entry
	filteredEntries []fuzzy.Result

	// Search state
	searchQuery  string
	searchCursor int

	// List selection
	selectedIndex int

	// Mode and state
	mode            Mode
	markedForDelete map[int]bool
	deleteConfirm   string

	// Rename state
	renameTarget *tries.Entry
	renameInput  string
	renameCursor int

	// Create state
	createInput  string
	createCursor int

	// Display
	width      int
	height     int
	statusMsg  string
	errorMsg   string
	outputCmd  string // Shell command to output on exit

	// Styling
	keys   KeyMap
	styles Styles

	// Flag to exit
	shouldExit bool
}

// New creates a new TUI model with the given manager
func New(manager *tries.Manager) Model {
	return Model{
		manager:         manager,
		allEntries:      []tries.Entry{},
		filteredEntries: []fuzzy.Result{},
		searchQuery:     "",
		searchCursor:    0,
		selectedIndex:   0,
		mode:            ModeNormal,
		markedForDelete: make(map[int]bool),
		deleteConfirm:   "",
		renameTarget:    nil,
		renameInput:     "",
		renameCursor:    0,
		createInput:     "",
		createCursor:    0,
		width:           80,
		height:          24,
		statusMsg:       "",
		errorMsg:        "",
		outputCmd:       "",
		keys:            NewKeyMap(),
		styles:          NewStyles(),
		shouldExit:      false,
	}
}

// Init initializes the model and loads entries
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.manager.List()
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries}
	}
}

// Message types
type entriesLoadedMsg struct {
	entries []tries.Entry
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

// ShouldExit returns whether the TUI should exit
func (m Model) ShouldExit() bool {
	return m.shouldExit
}

// OutputCommand returns the command to output on exit
func (m Model) OutputCommand() string {
	return m.outputCmd
}
