package promote

import (
	"strings"
	"testing"
)

// TestRunPicker_UnsupportedPicker verifies the picker dispatch rejects unknown
// picker names with a helpful error (the interactive fzf/builtin paths require
// a TTY and are exercised manually / in the picker package's own tests).
func TestRunPicker_UnsupportedPicker(t *testing.T) {
	_, err := runPicker("nano", "pick", []string{"a", "b"})
	if err == nil {
		t.Fatal("expected an error for an unsupported picker")
	}
	if !strings.Contains(err.Error(), "unsupported picker") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "fzf") || !strings.Contains(err.Error(), "builtin") {
		t.Errorf("error should mention both valid pickers, got: %v", err)
	}
}
