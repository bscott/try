// Package picker provides a small, single-purpose interactive fuzzy picker:
// "show these labels, let the user filter and choose one." It is the built-in
// fallback used by `try promote` when fzf is not installed (or when the picker
// is explicitly set to "builtin").
//
// Unlike internal/tui (which is a full manager: create/rename/delete), this is
// deliberately select-only. It reuses internal/fuzzy for match ranking so the
// filtering behaves like the rest of the tool.
//
// The Bubble Tea UI renders to stderr (see Select) so that stdout stays clean
// for the shell wrapper's `cd` output.
package picker

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrCancelled is returned when the user aborts the picker with ESC or Ctrl-C,
// mirroring the "cancelled" path fzf takes on exit code 130.
var ErrCancelled = errors.New("cancelled")

// maxVisible caps how many rows the list renders at once so a huge candidate
// set doesn't scroll the prompt off-screen.
const maxVisible = 15

// Model is the Bubble Tea model for the single-select picker. It is exported so
// its behavior can be unit-tested without spinning up a real terminal.
type Model struct {
	prompt string
	labels []string // all choices, in caller-provided order

	query    string
	filtered []string // labels currently shown, in display order
	cursor   int      // index into filtered

	choice    string
	cancelled bool

	width  int
	height int

	styles styles
}

type styles struct {
	prompt   lipgloss.Style
	query    lipgloss.Style
	selected lipgloss.Style
	item     lipgloss.Style
	help     lipgloss.Style
}

func newStyles() styles {
	return styles{
		prompt: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")), // bright blue
		query: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // light gray
		selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("25")).
			Bold(true),
		item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// New builds a picker model for the given prompt and labels.
func New(prompt string, labels []string) Model {
	m := Model{
		prompt: prompt,
		labels: labels,
		width:  80,
		height: 24,
		styles: newStyles(),
	}
	m.filtered = filter(m.query, m.labels)
	return m
}

// Choice returns the selected label (empty if cancelled or nothing chosen).
func (m Model) Choice() string { return m.choice }

// Cancelled reports whether the user aborted the picker.
func (m Model) Cancelled() bool { return m.cancelled }

// filter returns the labels matching query, in ranked order. An empty query
// preserves the caller-provided order; a non-empty query defers to the shared
// fuzzy matcher for scoring/ordering.
func filter(query string, labels []string) []string {
	if strings.TrimSpace(query) == "" {
		out := make([]string, len(labels))
		copy(out, labels)
		return out
	}
	entries := make([]tries.Entry, len(labels))
	for i, l := range labels {
		entries[i] = tries.Entry{Name: l}
	}
	results := fuzzy.Match(query, entries)
	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, r.Entry.Name)
	}
	return out
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 {
				m.choice = m.filtered[m.cursor]
				return m, tea.Quit
			}
			return m, nil

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "backspace":
			if r := []rune(m.query); len(r) > 0 {
				m.query = string(r[:len(r)-1])
				m.refilter()
			}
			return m, nil

		case "ctrl+u":
			m.query = ""
			m.refilter()
			return m, nil

		default:
			switch msg.Type {
			case tea.KeyRunes:
				m.query += string(msg.Runes)
				m.refilter()
			case tea.KeySpace:
				m.query += " "
				m.refilter()
			}
			return m, nil
		}
	}
	return m, nil
}

// refilter recomputes the visible list and clamps the cursor into range.
func (m *Model) refilter() {
	m.filtered = filter(m.query, m.labels)
	if m.cursor > len(m.filtered)-1 {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.styles.prompt.Render(m.prompt))
	b.WriteString("\n")
	b.WriteString(m.styles.prompt.Render("> "))
	b.WriteString(m.styles.query.Render(m.query))
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(m.styles.help.Render("  (no matches)"))
		b.WriteString("\n")
	} else {
		start, end := m.window()
		for i := start; i < end; i++ {
			label := m.filtered[i]
			if i == m.cursor {
				b.WriteString(m.styles.selected.Render("> " + label))
			} else {
				b.WriteString(m.styles.item.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.help.Render("↑/↓ or ctrl+p/n move · type to filter · enter select · esc cancel"))
	return b.String()
}

// window returns the [start, end) slice of filtered to render, keeping the
// cursor visible within maxVisible rows.
func (m Model) window() (int, int) {
	n := len(m.filtered)
	if n <= maxVisible {
		return 0, n
	}
	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > n {
		end = n
		start = end - maxVisible
	}
	return start, end
}

// Select runs the interactive picker over labels and returns the chosen label.
// ESC/Ctrl-C returns ErrCancelled. The UI renders to stderr so stdout stays
// available for the caller's machine-readable output (e.g. a `cd` command).
func Select(prompt string, labels []string) (string, error) {
	if len(labels) == 0 {
		return "", errors.New("no options to choose from")
	}

	p := tea.NewProgram(
		New(prompt, labels),
		tea.WithOutput(os.Stderr),
		tea.WithAltScreen(),
	)
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker: %w", err)
	}

	m, ok := final.(Model)
	if !ok {
		return "", fmt.Errorf("picker: unexpected model type %T", final)
	}
	if m.cancelled || m.choice == "" {
		return "", ErrCancelled
	}
	return m.choice, nil
}
