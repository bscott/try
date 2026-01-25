package tui

import (
	"fmt"
	"strings"

	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
)

// shellQuote quotes a string for safe use in shell commands
// Uses single quotes and escapes any single quotes in the string
func shellQuote(s string) string {
	// Replace ' with '\''
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return fmt.Sprintf("'%s'", escaped)
}

// updateSearch re-filters the entries based on the current search query
func (m *Model) updateSearch() {
	m.filteredEntries = fuzzy.Match(m.searchQuery, m.allEntries)
	m.ensureSelectionInBounds()
}

// ensureSelectionInBounds keeps the selected index within valid range
func (m *Model) ensureSelectionInBounds() {
	if len(m.filteredEntries) == 0 {
		m.selectedIndex = 0
		return
	}

	if m.selectedIndex >= len(m.filteredEntries) {
		m.selectedIndex = len(m.filteredEntries) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// getSelectedEntry returns the currently selected entry, or nil if none
func (m *Model) getSelectedEntry() *tries.Entry {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.filteredEntries) {
		return nil
	}
	return &m.filteredEntries[m.selectedIndex].Entry
}

// outputCommand sets the command to output when exiting
func (m *Model) outputCommand(cmd string) {
	m.outputCmd = cmd
	m.shouldExit = true
}

// deleteWordBack deletes the word before the cursor in a string
func deleteWordBack(s string, cursor int) (string, int) {
	if cursor == 0 {
		return s, 0
	}

	// Find the start of the word
	wordStart := cursor - 1

	// Skip any trailing spaces
	for wordStart > 0 && s[wordStart] == ' ' {
		wordStart--
	}

	// Find the actual word start
	for wordStart > 0 && s[wordStart-1] != ' ' {
		wordStart--
	}

	// Delete from wordStart to cursor
	newStr := s[:wordStart] + s[cursor:]
	return newStr, wordStart
}

// killLine deletes from cursor to end of line
func killLine(s string, cursor int) string {
	if cursor >= len(s) {
		return s
	}
	return s[:cursor]
}

// insertChar inserts a character at the cursor position
func insertChar(s string, cursor int, ch rune) (string, int) {
	before := s[:cursor]
	after := s[cursor:]
	return before + string(ch) + after, cursor + 1
}

// deleteChar deletes the character before the cursor
func deleteChar(s string, cursor int) (string, int) {
	if cursor == 0 {
		return s, 0
	}
	before := s[:cursor-1]
	after := s[cursor:]
	return before + after, cursor - 1
}

// moveCursor moves the cursor in the string, ensuring it stays in bounds
func moveCursor(cursor, delta, maxLen int) int {
	newCursor := cursor + delta
	if newCursor < 0 {
		return 0
	}
	if newCursor > maxLen {
		return maxLen
	}
	return newCursor
}

// validateName checks if a name is valid for a try directory
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("name cannot contain slashes")
	}
	if strings.Contains(name, "\\") {
		return fmt.Errorf("name cannot contain backslashes")
	}
	if name == "." || name == ".." || strings.HasPrefix(name, "..") {
		return fmt.Errorf("name cannot be or start with '..'")
	}
	if strings.ContainsAny(name, "\x00") {
		return fmt.Errorf("name cannot contain null bytes")
	}
	return nil
}
