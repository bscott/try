package picker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Coverage categories for the built-in picker:
//
//   - Security     : covered — the built-in picker never shells out; typed
//                    input (even shell metacharacters) is treated purely as
//                    filter text.
//   - Performance  : covered — filtering a large candidate set and the visible
//                    window stay bounded.
//   - Retry        : N/A — pure in-memory UI state machine, nothing to retry.
//   - Unit         : covered — filter, New, window, and accessor behavior.
//   - Integration  : covered — a full keystroke sequence (type → navigate →
//                    select) drives the model to a choice.
//   - Functional   : covered — cancel paths and enter-on-empty behavior.
//   - Frame        : N/A — no wire/protocol framing at this layer.

// send feeds a key message to the model and returns the updated Model plus
// whether a command (e.g. tea.Quit) was produced.
func send(m Model, msg tea.Msg) (Model, bool) {
	next, cmd := m.Update(msg)
	return next.(Model), cmd != nil
}

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// --- Unit -----------------------------------------------------------------

func TestUnit_NewInitialState(t *testing.T) {
	m := New("pick one", []string{"alpha", "beta", "gamma"})
	if m.prompt != "pick one" {
		t.Errorf("prompt not stored: %q", m.prompt)
	}
	if len(m.filtered) != 3 {
		t.Errorf("expected all 3 labels visible initially, got %d", len(m.filtered))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
	if m.Choice() != "" || m.Cancelled() {
		t.Error("fresh model should have no choice and not be cancelled")
	}
}

func TestUnit_FilterEmptyPreservesOrder(t *testing.T) {
	labels := []string{"zeta", "alpha", "mid"}
	got := filter("", labels)
	if len(got) != len(labels) {
		t.Fatalf("expected %d, got %d", len(labels), len(got))
	}
	for i := range labels {
		if got[i] != labels[i] {
			t.Errorf("order changed at %d: got %q want %q", i, got[i], labels[i])
		}
	}
}

func TestUnit_FilterNonEmptyMatchesSubset(t *testing.T) {
	labels := []string{"subtrackr", "chat-tails", "getm", "subtle"}
	got := filter("sub", labels)
	if len(got) == 0 {
		t.Fatal("expected some matches for 'sub'")
	}
	for _, g := range got {
		if !strings.Contains(strings.ToLower(g), "s") {
			t.Errorf("unexpected match %q for query 'sub'", g)
		}
	}
	// "getm" and "chat-tails" have no fuzzy 'sub' subsequence; ensure filtering
	// actually narrowed the set.
	if len(got) >= len(labels) {
		t.Errorf("expected query to narrow results, got %d of %d", len(got), len(labels))
	}
}

func TestUnit_WindowBounds(t *testing.T) {
	labels := make([]string, 100)
	for i := range labels {
		labels[i] = strings.Repeat("x", i%5+1)
	}
	m := New("p", labels)
	m.cursor = 50
	start, end := m.window()
	if end-start > maxVisible {
		t.Errorf("window larger than maxVisible: %d", end-start)
	}
	if m.cursor < start || m.cursor >= end {
		t.Errorf("cursor %d not within window [%d,%d)", m.cursor, start, end)
	}
}

// --- Integration ----------------------------------------------------------

// TestIntegration_TypeNavigateSelect drives the model through a realistic
// keystroke sequence and asserts it lands on the expected choice.
func TestIntegration_TypeNavigateSelect(t *testing.T) {
	m := New("Promote which try?", []string{
		"2026-01-01-alpha",
		"2026-02-02-beta",
		"2026-03-03-betamax",
	})

	// Filter down to the "beta" entries.
	m, _ = send(m, runes("beta"))
	if len(m.filtered) < 2 {
		t.Fatalf("expected at least 2 'beta' matches, got %d: %v", len(m.filtered), m.filtered)
	}

	// Move down one and select.
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	want := m.filtered[1]
	m, quit := send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !quit {
		t.Error("enter should produce a quit command")
	}
	if m.Choice() != want {
		t.Errorf("choice = %q, want %q", m.Choice(), want)
	}
	if m.Cancelled() {
		t.Error("selection should not be marked cancelled")
	}
}

// TestIntegration_BackspaceRewidensAndClampsCursor verifies editing the query
// keeps the cursor valid.
func TestIntegration_BackspaceRewidensAndClampsCursor(t *testing.T) {
	m := New("p", []string{"alpha", "alto", "beta"})
	m, _ = send(m, runes("al"))
	// Move cursor to the last match, then broaden the query.
	for i := 0; i < len(m.filtered)-1; i++ {
		m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyBackspace}) // "al" -> "a"
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		t.Errorf("cursor %d out of range after backspace (len %d)", m.cursor, len(m.filtered))
	}
}

// --- Functional -----------------------------------------------------------

func TestFunctional_EscCancels(t *testing.T) {
	m := New("p", []string{"a", "b"})
	m, quit := send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if !quit {
		t.Error("esc should quit")
	}
	if !m.Cancelled() || m.Choice() != "" {
		t.Errorf("esc should cancel with no choice; cancelled=%v choice=%q", m.Cancelled(), m.Choice())
	}
}

func TestFunctional_CtrlCCancels(t *testing.T) {
	m := New("p", []string{"a", "b"})
	m, quit := send(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !quit || !m.Cancelled() {
		t.Errorf("ctrl+c should quit and cancel; quit=%v cancelled=%v", quit, m.Cancelled())
	}
}

func TestFunctional_EnterOnEmptyDoesNothing(t *testing.T) {
	m := New("p", []string{"alpha"})
	m, _ = send(m, runes("zzz")) // no matches
	if len(m.filtered) != 0 {
		t.Fatalf("expected zero matches for 'zzz', got %v", m.filtered)
	}
	m, quit := send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if quit {
		t.Error("enter with no matches should not quit")
	}
	if m.Choice() != "" {
		t.Errorf("expected no choice, got %q", m.Choice())
	}
}

func TestFunctional_ViewRendersPromptAndItems(t *testing.T) {
	m := New("Promote into: ~/code", []string{"personal", "work"})
	out := m.View()
	for _, want := range []string{"Promote into: ~/code", "personal", "work"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q; got:\n%s", want, out)
		}
	}
}

// --- Security -------------------------------------------------------------

// TestSecurity_MetacharactersAreLiteralFilterText confirms shell metacharacters
// typed into the query are treated purely as filter text. Unlike the fzf path,
// the built-in picker never spawns a subprocess, so there is no command
// injection surface; this guards that typing such input neither panics nor is
// interpreted specially.
func TestSecurity_MetacharactersAreLiteralFilterText(t *testing.T) {
	m := New("p", []string{"safe-entry", "rm -rf /", "normal"})
	m, _ = send(m, runes("rm -rf /"))
	// The exact-match label should still be selectable as plain text.
	found := false
	for _, l := range m.filtered {
		if l == "rm -rf /" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected literal label 'rm -rf /' to remain a filterable option, got %v", m.filtered)
	}
}

// --- Performance ----------------------------------------------------------

func TestPerformance_FilterLargeSetBounded(t *testing.T) {
	labels := make([]string, 5000)
	for i := range labels {
		labels[i] = "proj-" + strings.Repeat("a", i%7)
	}
	m := New("p", labels)
	m, _ = send(m, runes("proj"))
	// Visible window must stay capped regardless of match count.
	start, end := m.window()
	if end-start > maxVisible {
		t.Errorf("visible window exceeded cap: %d", end-start)
	}
}

// --- Retry / Frame --------------------------------------------------------

func TestRetry_NotApplicable(t *testing.T) {
	t.Skip("N/A: picker is a pure in-memory UI state machine with no retryable I/O")
}

func TestFrame_NotApplicable(t *testing.T) {
	t.Skip("N/A: no wire/protocol framing at the picker layer")
}
