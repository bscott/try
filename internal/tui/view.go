package tui

import (
	"fmt"
	"strings"

	"github.com/bscott/try/internal/tries"
)

// View renders the TUI based on current state
func (m Model) View() string {
	if m.shouldExit {
		// Don't render anything when exiting
		// The command will be output separately by the caller
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render("try - Manage temporary directories"))
	b.WriteString("\n\n")

	// Render based on mode
	switch m.mode {
	case ModeNormal, ModeDelete:
		m.renderNormalView(&b)
	case ModeConfirmDelete:
		m.renderConfirmDeleteView(&b)
	case ModeRename:
		m.renderRenameView(&b)
	case ModeCreate:
		m.renderCreateView(&b)
	}

	// Status bar
	b.WriteString("\n")
	m.renderStatusBar(&b)

	return b.String()
}

// renderNormalView renders the search and list view
func (m *Model) renderNormalView(b *strings.Builder) {
	// Search input
	b.WriteString(m.styles.SearchPrefix.Render("> "))
	m.renderSearchInput(b)
	b.WriteString("\n\n")

	// Entry list
	if len(m.filteredEntries) == 0 {
		if m.searchQuery == "" {
			b.WriteString(m.styles.HelpText.Render("No try directories found. Press Ctrl+T to create one."))
		} else {
			b.WriteString(m.styles.HelpText.Render(fmt.Sprintf("No matches for '%s'", m.searchQuery)))
		}
		return
	}

	// Calculate visible range (for scrolling)
	maxVisible := m.height - 10 // Reserve space for header, search, status
	if maxVisible < 5 {
		maxVisible = 5
	}

	startIdx := m.selectedIndex - maxVisible/2
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.filteredEntries) {
		endIdx = len(m.filteredEntries)
		startIdx = endIdx - maxVisible
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Render visible entries
	for i := startIdx; i < endIdx; i++ {
		entry := m.filteredEntries[i].Entry
		isSelected := i == m.selectedIndex
		isMarked := m.markedForDelete[i]

		m.renderEntry(b, entry, isSelected, isMarked)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filteredEntries) > maxVisible {
		scrollInfo := fmt.Sprintf("(%d/%d)", m.selectedIndex+1, len(m.filteredEntries))
		b.WriteString(m.styles.HelpText.Render(scrollInfo))
		b.WriteString("\n")
	}
}

// renderSearchInput renders the search input with cursor
func (m *Model) renderSearchInput(b *strings.Builder) {
	before := m.searchQuery[:m.searchCursor]
	var cursor, after string

	if m.searchCursor < len(m.searchQuery) {
		cursor = string(m.searchQuery[m.searchCursor])
		after = m.searchQuery[m.searchCursor+1:]
	} else {
		cursor = " "
		after = ""
	}

	b.WriteString(m.styles.SearchInput.Render(before))
	b.WriteString(m.styles.SearchCursor.Render(cursor))
	b.WriteString(m.styles.SearchInput.Render(after))
}

// renderEntry renders a single entry in the list
func (m *Model) renderEntry(b *strings.Builder, entry tries.Entry, isSelected, isMarked bool) {
	var line strings.Builder

	// Mark indicator
	if isMarked {
		line.WriteString("✗ ")
	} else {
		line.WriteString("  ")
	}

	// Date prefix and name
	if entry.HasDatePrefix {
		datePart := m.styles.DatePrefix.Render(fmt.Sprintf("[%s]", entry.DatePrefix))
		line.WriteString(datePart)
		line.WriteString(" ")
		line.WriteString(entry.DisplayName())
	} else {
		line.WriteString(entry.Name)
	}

	// Apply styling
	text := line.String()
	if isMarked {
		text = m.styles.ListItemMarked.Render(text)
	} else if isSelected {
		text = m.styles.ListItemSelected.Render(text)
	} else {
		text = m.styles.ListItem.Render(text)
	}

	b.WriteString(text)
}

// renderConfirmDeleteView renders the delete confirmation prompt
func (m *Model) renderConfirmDeleteView(b *strings.Builder) {
	b.WriteString(m.styles.Prompt.Render(fmt.Sprintf("Delete %d entries?", len(m.markedForDelete))))
	b.WriteString("\n\n")

	// List entries to be deleted
	b.WriteString("The following entries will be deleted:\n\n")
	for idx := range m.markedForDelete {
		if idx < len(m.filteredEntries) {
			entry := m.filteredEntries[idx].Entry
			b.WriteString("  ")
			b.WriteString(m.styles.ListItemMarked.Render("✗ " + entry.Name))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Prompt.Render("Type YES to confirm: "))
	b.WriteString(m.styles.PromptInput.Render(m.deleteConfirm))
	b.WriteString(m.styles.SearchCursor.Render(" "))
	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpText.Render("Press Esc to cancel"))
}

// renderRenameView renders the rename input
func (m *Model) renderRenameView(b *strings.Builder) {
	if m.renameTarget == nil {
		return
	}

	b.WriteString(m.styles.Prompt.Render(fmt.Sprintf("Rename: %s", m.renameTarget.Name)))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Prompt.Render("New name: "))

	// Render input with cursor
	before := m.renameInput[:m.renameCursor]
	var cursor, after string

	if m.renameCursor < len(m.renameInput) {
		cursor = string(m.renameInput[m.renameCursor])
		after = m.renameInput[m.renameCursor+1:]
	} else {
		cursor = " "
		after = ""
	}

	b.WriteString(m.styles.PromptInput.Render(before))
	b.WriteString(m.styles.SearchCursor.Render(cursor))
	b.WriteString(m.styles.PromptInput.Render(after))
	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpText.Render("Press Enter to confirm, Esc to cancel"))
}

// renderCreateView renders the create input
func (m *Model) renderCreateView(b *strings.Builder) {
	b.WriteString(m.styles.Prompt.Render("Create new try directory"))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Prompt.Render("Name: "))

	// Render input with cursor
	before := m.createInput[:m.createCursor]
	var cursor, after string

	if m.createCursor < len(m.createInput) {
		cursor = string(m.createInput[m.createCursor])
		after = m.createInput[m.createCursor+1:]
	} else {
		cursor = " "
		after = ""
	}

	b.WriteString(m.styles.PromptInput.Render(before))
	b.WriteString(m.styles.SearchCursor.Render(cursor))
	b.WriteString(m.styles.PromptInput.Render(after))
	b.WriteString("\n\n")
	b.WriteString(m.styles.HelpText.Render("Will be created as: YYYY-MM-DD-name"))
	b.WriteString("\n")
	b.WriteString(m.styles.HelpText.Render("Press Enter to confirm, Esc to cancel"))
}

// renderStatusBar renders the status bar with mode and messages
func (m *Model) renderStatusBar(b *strings.Builder) {
	// Mode indicator
	modeStr := fmt.Sprintf("[%s]", m.mode.String())
	b.WriteString(m.styles.StatusBar.Render(modeStr))

	// Error message takes precedence
	if m.errorMsg != "" {
		b.WriteString(" ")
		b.WriteString(m.styles.ErrorMessage.Render(m.errorMsg))
		return
	}

	// Status message
	if m.statusMsg != "" {
		b.WriteString(" ")
		b.WriteString(m.styles.StatusMessage.Render(m.statusMsg))
		return
	}

	// Help text based on mode
	b.WriteString(" ")
	switch m.mode {
	case ModeNormal:
		help := "Ctrl+D: delete | Ctrl+R: rename | Ctrl+T: create | Ctrl+C: quit"
		b.WriteString(m.styles.HelpText.Render(help))
	case ModeDelete:
		help := "Ctrl+D: mark | Enter: confirm | Esc: cancel"
		b.WriteString(m.styles.HelpText.Render(help))
	}
}
