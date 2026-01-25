package tui

import (
	"fmt"

	"github.com/bscott/try/internal/tries"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case entriesLoadedMsg:
		m.allEntries = msg.entries
		m.updateSearch()
		return m, nil

	case errMsg:
		m.errorMsg = msg.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input based on the current mode
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit handler
	if msg.Type == tea.KeyCtrlC {
		m.shouldExit = true
		return m, tea.Quit
	}

	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeDelete:
		return m.handleDeleteMode(msg)
	case ModeConfirmDelete:
		return m.handleConfirmDeleteMode(msg)
	case ModeRename:
		return m.handleRenameMode(msg)
	case ModeCreate:
		return m.handleCreateMode(msg)
	}

	return m, nil
}

// handleNormalMode handles input in normal mode
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear status messages
	m.statusMsg = ""
	m.errorMsg = ""

	switch {
	// Navigation
	case key.Matches(msg, m.keys.Up):
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedIndex < len(m.filteredEntries)-1 {
			m.selectedIndex++
		}
		return m, nil

	// Selection
	case key.Matches(msg, m.keys.Select):
		entry := m.getSelectedEntry()
		if entry != nil {
			m.outputCommand(fmt.Sprintf("cd %s", shellQuote(entry.Path)))
			return m, tea.Quit
		}
		return m, nil

	// Mode switches
	case key.Matches(msg, m.keys.ToggleDelete):
		m.mode = ModeDelete
		m.markedForDelete = make(map[int]bool)
		m.statusMsg = "Delete mode - mark entries with Ctrl+D, Enter to confirm"
		return m, nil

	case key.Matches(msg, m.keys.Rename):
		entry := m.getSelectedEntry()
		if entry != nil {
			m.mode = ModeRename
			m.renameTarget = entry
			m.renameInput = entry.DisplayName()
			m.renameCursor = len(m.renameInput)
			m.statusMsg = "Rename mode - edit name and press Enter"
		}
		return m, nil

	case key.Matches(msg, m.keys.Create):
		m.mode = ModeCreate
		m.createInput = ""
		m.createCursor = 0
		m.statusMsg = "Create mode - enter name for new try directory"
		return m, nil

	// Search input
	case key.Matches(msg, m.keys.CursorStart):
		m.searchCursor = 0
		return m, nil

	case key.Matches(msg, m.keys.CursorEnd):
		m.searchCursor = len(m.searchQuery)
		return m, nil

	case key.Matches(msg, m.keys.CursorBack):
		m.searchCursor = moveCursor(m.searchCursor, -1, len(m.searchQuery))
		return m, nil

	case key.Matches(msg, m.keys.CursorForward):
		m.searchCursor = moveCursor(m.searchCursor, 1, len(m.searchQuery))
		return m, nil

	case key.Matches(msg, m.keys.DeleteWordBack):
		m.searchQuery, m.searchCursor = deleteWordBack(m.searchQuery, m.searchCursor)
		m.updateSearch()
		return m, nil

	case key.Matches(msg, m.keys.KillLine):
		m.searchQuery = killLine(m.searchQuery, m.searchCursor)
		m.updateSearch()
		return m, nil

	case key.Matches(msg, m.keys.Backspace):
		m.searchQuery, m.searchCursor = deleteChar(m.searchQuery, m.searchCursor)
		m.updateSearch()
		return m, nil

	// Regular character input
	default:
		if len(msg.Runes) > 0 {
			m.searchQuery, m.searchCursor = insertChar(m.searchQuery, m.searchCursor, msg.Runes[0])
			m.updateSearch()
		}
		return m, nil
	}
}

// handleDeleteMode handles input in delete mode
func (m Model) handleDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.markedForDelete = make(map[int]bool)
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedIndex < len(m.filteredEntries)-1 {
			m.selectedIndex++
		}
		return m, nil

	case key.Matches(msg, m.keys.ToggleDelete):
		// Toggle mark for current entry
		if m.markedForDelete[m.selectedIndex] {
			delete(m.markedForDelete, m.selectedIndex)
		} else {
			m.markedForDelete[m.selectedIndex] = true
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if len(m.markedForDelete) == 0 {
			m.errorMsg = "No entries marked for deletion"
			return m, nil
		}
		m.mode = ModeConfirmDelete
		m.deleteConfirm = ""
		m.statusMsg = fmt.Sprintf("Type YES to delete %d entries", len(m.markedForDelete))
		return m, nil
	}

	return m, nil
}

// handleConfirmDeleteMode handles input in confirm delete mode
func (m Model) handleConfirmDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.markedForDelete = make(map[int]bool)
		m.deleteConfirm = ""
		m.statusMsg = "Delete cancelled"
		return m, nil

	case key.Matches(msg, m.keys.Backspace):
		if len(m.deleteConfirm) > 0 {
			m.deleteConfirm = m.deleteConfirm[:len(m.deleteConfirm)-1]
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		if m.deleteConfirm == "YES" {
			// Perform deletion
			return m.performDeletion()
		}
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			m.deleteConfirm += string(msg.Runes[0])
		}
		return m, nil
	}
}

// handleRenameMode handles input in rename mode
func (m Model) handleRenameMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.renameTarget = nil
		m.renameInput = ""
		m.statusMsg = "Rename cancelled"
		return m, nil

	case key.Matches(msg, m.keys.Select):
		// Validate and perform rename
		if err := validateName(m.renameInput); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		// Perform rename
		if err := m.manager.Rename(*m.renameTarget, m.renameInput); err != nil {
			m.errorMsg = fmt.Sprintf("Rename failed: %v", err)
			return m, nil
		}

		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Renamed to %s", m.renameInput)
		m.renameTarget = nil
		m.renameInput = ""

		// Reload entries
		return m, func() tea.Msg {
			entries, err := m.manager.List()
			if err != nil {
				return errMsg{err}
			}
			return entriesLoadedMsg{entries}
		}

	case key.Matches(msg, m.keys.CursorStart):
		m.renameCursor = 0
		return m, nil

	case key.Matches(msg, m.keys.CursorEnd):
		m.renameCursor = len(m.renameInput)
		return m, nil

	case key.Matches(msg, m.keys.CursorBack):
		m.renameCursor = moveCursor(m.renameCursor, -1, len(m.renameInput))
		return m, nil

	case key.Matches(msg, m.keys.CursorForward):
		m.renameCursor = moveCursor(m.renameCursor, 1, len(m.renameInput))
		return m, nil

	case key.Matches(msg, m.keys.DeleteWordBack):
		m.renameInput, m.renameCursor = deleteWordBack(m.renameInput, m.renameCursor)
		return m, nil

	case key.Matches(msg, m.keys.KillLine):
		m.renameInput = killLine(m.renameInput, m.renameCursor)
		return m, nil

	case key.Matches(msg, m.keys.Backspace):
		m.renameInput, m.renameCursor = deleteChar(m.renameInput, m.renameCursor)
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			m.renameInput, m.renameCursor = insertChar(m.renameInput, m.renameCursor, msg.Runes[0])
		}
		return m, nil
	}
}

// handleCreateMode handles input in create mode
func (m Model) handleCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeNormal
		m.createInput = ""
		m.statusMsg = "Create cancelled"
		return m, nil

	case key.Matches(msg, m.keys.Select):
		// Validate and create
		if err := validateName(m.createInput); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		// Create new directory (manager adds date prefix)
		dirPath, err := m.manager.Create(m.createInput)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Create failed: %v", err)
			return m, nil
		}

		// Output cd command and exit
		m.outputCmd = fmt.Sprintf("cd %s", shellQuote(dirPath))
		m.shouldExit = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.CursorStart):
		m.createCursor = 0
		return m, nil

	case key.Matches(msg, m.keys.CursorEnd):
		m.createCursor = len(m.createInput)
		return m, nil

	case key.Matches(msg, m.keys.CursorBack):
		m.createCursor = moveCursor(m.createCursor, -1, len(m.createInput))
		return m, nil

	case key.Matches(msg, m.keys.CursorForward):
		m.createCursor = moveCursor(m.createCursor, 1, len(m.createInput))
		return m, nil

	case key.Matches(msg, m.keys.DeleteWordBack):
		m.createInput, m.createCursor = deleteWordBack(m.createInput, m.createCursor)
		return m, nil

	case key.Matches(msg, m.keys.KillLine):
		m.createInput = killLine(m.createInput, m.createCursor)
		return m, nil

	case key.Matches(msg, m.keys.Backspace):
		m.createInput, m.createCursor = deleteChar(m.createInput, m.createCursor)
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			m.createInput, m.createCursor = insertChar(m.createInput, m.createCursor, msg.Runes[0])
		}
		return m, nil
	}
}

// performDeletion deletes all marked entries
func (m Model) performDeletion() (tea.Model, tea.Cmd) {
	deletedCount := 0

	// Collect entries to delete
	toDelete := make([]tries.Entry, 0, len(m.markedForDelete))
	for idx := range m.markedForDelete {
		if idx < len(m.filteredEntries) {
			toDelete = append(toDelete, m.filteredEntries[idx].Entry)
		}
	}

	// Delete all marked entries
	if err := m.manager.Delete(toDelete); err != nil {
		m.errorMsg = fmt.Sprintf("Delete failed: %v", err)
	} else {
		deletedCount = len(toDelete)
		m.statusMsg = fmt.Sprintf("Deleted %d entries", deletedCount)
	}

	// Reset state
	m.mode = ModeNormal
	m.markedForDelete = make(map[int]bool)
	m.deleteConfirm = ""

	// Reload entries
	return m, func() tea.Msg {
		entries, err := m.manager.List()
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries}
	}
}
