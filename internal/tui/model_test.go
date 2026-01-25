package tui

import (
	"testing"

	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
)

func TestNew(t *testing.T) {
	manager := tries.NewManager("/tmp/test")
	model := New(manager)

	// Verify initial state
	if model.mode != ModeNormal {
		t.Errorf("expected mode to be ModeNormal, got %v", model.mode)
	}

	if model.searchQuery != "" {
		t.Errorf("expected empty search query, got %q", model.searchQuery)
	}

	if model.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to be 0, got %d", model.selectedIndex)
	}

	if len(model.markedForDelete) != 0 {
		t.Errorf("expected empty markedForDelete, got %d items", len(model.markedForDelete))
	}

	if model.shouldExit {
		t.Error("expected shouldExit to be false")
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeNormal, "NORMAL"},
		{ModeDelete, "DELETE"},
		{ModeRename, "RENAME"},
		{ModeCreate, "CREATE"},
		{ModeConfirmDelete, "CONFIRM DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("Mode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetSelectedEntry(t *testing.T) {
	manager := tries.NewManager("/tmp/test")
	model := New(manager)

	// No entries - should return nil
	entry := model.getSelectedEntry()
	if entry != nil {
		t.Error("expected nil entry when no entries exist")
	}

	// Add some mock filtered entries
	model.filteredEntries = []fuzzy.Result{
		{Entry: tries.Entry{Name: "test-1", Path: "/tmp/test/test-1"}},
		{Entry: tries.Entry{Name: "test-2", Path: "/tmp/test/test-2"}},
	}
	model.selectedIndex = 0

	// Should return first entry
	entry = model.getSelectedEntry()
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Name != "test-1" {
		t.Errorf("expected name 'test-1', got %q", entry.Name)
	}

	// Select second entry
	model.selectedIndex = 1
	entry = model.getSelectedEntry()
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Name != "test-2" {
		t.Errorf("expected name 'test-2', got %q", entry.Name)
	}

	// Out of bounds - should return nil
	model.selectedIndex = 10
	entry = model.getSelectedEntry()
	if entry != nil {
		t.Error("expected nil entry when index out of bounds")
	}
}

func TestEnsureSelectionInBounds(t *testing.T) {
	manager := tries.NewManager("/tmp/test")
	model := New(manager)

	// No entries
	model.selectedIndex = 5
	model.ensureSelectionInBounds()
	if model.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to be 0, got %d", model.selectedIndex)
	}

	// Add entries
	model.filteredEntries = []fuzzy.Result{
		{Entry: tries.Entry{Name: "test-1"}},
		{Entry: tries.Entry{Name: "test-2"}},
		{Entry: tries.Entry{Name: "test-3"}},
	}

	// Index too high
	model.selectedIndex = 10
	model.ensureSelectionInBounds()
	if model.selectedIndex != 2 {
		t.Errorf("expected selectedIndex to be 2, got %d", model.selectedIndex)
	}

	// Index negative
	model.selectedIndex = -5
	model.ensureSelectionInBounds()
	if model.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to be 0, got %d", model.selectedIndex)
	}

	// Index valid
	model.selectedIndex = 1
	model.ensureSelectionInBounds()
	if model.selectedIndex != 1 {
		t.Errorf("expected selectedIndex to be 1, got %d", model.selectedIndex)
	}
}
