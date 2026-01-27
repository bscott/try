package tui

import (
	"fmt"
	"strings"
	"time"

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

	// Column headers
	m.renderColumnHeaders(b)
	b.WriteString("\n")

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

// renderColumnHeaders renders the column headers for the entry list
func (m *Model) renderColumnHeaders(b *strings.Builder) {
	const nameColWidth = 50 // Match the renderEntry width

	var header strings.Builder

	// Name column header
	nameHeader := m.styles.ColumnHeader.Render("NAME")
	header.WriteString(nameHeader)

	// Pad name column to fixed width (accounting for styling)
	// The visual width is just "NAME" (4 chars) so pad to nameColWidth
	padding := nameColWidth - 4 // "NAME" is 4 chars visible
	if padding > 0 {
		header.WriteString(strings.Repeat(" ", padding))
	}

	// Git status column header (no extra spacing needed)
	header.WriteString(m.styles.ColumnHeader.Render("GIT STATUS"))

	b.WriteString(header.String())
	b.WriteString("\n")

	// Separator line - match column widths
	separator := strings.Repeat("─", nameColWidth) + strings.Repeat("─", 30)
	b.WriteString(m.styles.Muted.Render(separator))
}

// renderEntry renders a single entry in the list
func (m *Model) renderEntry(b *strings.Builder, entry tries.Entry, isSelected, isMarked bool) {
	const nameColWidth = 50

	var nameCol, gitCol strings.Builder

	// Build name column content (without mark indicator)
	var displayText string
	if entry.HasDatePrefix {
		displayText = fmt.Sprintf("[%s] %s", entry.DatePrefix, entry.DisplayName())
	} else {
		displayText = entry.Name
	}

	// Pad or truncate name to fixed width
	if len(displayText) < nameColWidth {
		displayText += strings.Repeat(" ", nameColWidth-len(displayText))
	} else if len(displayText) > nameColWidth {
		displayText = displayText[:nameColWidth-3] + "..."
	}

	nameCol.WriteString(displayText)

	// Build git status column
	if entry.Git.IsRepo {
		// Branch name
		if entry.Git.Branch != "" {
			gitCol.WriteString(m.styles.GitBranch.Render(entry.Git.Branch))
		}

		// Ahead/behind
		if entry.Git.Ahead > 0 {
			gitCol.WriteString(m.styles.GitAhead.Render(fmt.Sprintf(" ↑%d", entry.Git.Ahead)))
		}
		if entry.Git.Behind > 0 {
			gitCol.WriteString(m.styles.GitBehind.Render(fmt.Sprintf(" ↓%d", entry.Git.Behind)))
		}

		// Dirty indicator
		if entry.Git.IsDirty {
			gitCol.WriteString(m.styles.GitDirty.Render(" •"))
		}

		// Last commit time
		if !entry.Git.LastCommit.IsZero() {
			ago := formatTimeAgo(entry.Git.LastCommit)
			gitCol.WriteString(m.styles.Muted.Render(fmt.Sprintf(" %s", ago)))
		}
	}

	// Combine columns with proper styling
	var line strings.Builder

	// Mark indicator (before the columns)
	if isMarked {
		line.WriteString(m.styles.ListItemMarked.Render("✗ "))
	} else {
		line.WriteString("  ")
	}

	// Style the name column based on selection/mark status
	nameText := nameCol.String()
	if entry.HasDatePrefix {
		// Apply date prefix styling
		datePart := fmt.Sprintf("[%s]", entry.DatePrefix)
		restPart := nameText[len(datePart):]

		if isSelected {
			line.WriteString(m.styles.DatePrefix.Render(datePart))
			line.WriteString(m.styles.ListItemSelected.Render(restPart))
		} else {
			line.WriteString(m.styles.DatePrefix.Render(datePart))
			line.WriteString(m.styles.ListItem.Render(restPart))
		}
	} else {
		if isSelected {
			line.WriteString(m.styles.ListItemSelected.Render(nameText))
		} else {
			line.WriteString(m.styles.ListItem.Render(nameText))
		}
	}

	line.WriteString(gitCol.String())

	b.WriteString(line.String())
}

// formatTimeAgo formats a time as a human-readable "ago" string
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	}
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
	weeks := int(duration.Hours() / 24 / 7)
	if weeks == 1 {
		return "1w ago"
	}
	return fmt.Sprintf("%dw ago", weeks)
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
