package promote

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bscott/try/internal/config"
)

// mkdir is a t.Fatal-on-failure helper for building test fixtures.
func mkdir(t *testing.T, parts ...string) string {
	t.Helper()
	p := filepath.Join(parts...)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
	return p
}

func TestCandidateDirs_Depth1(t *testing.T) {
	root := t.TempDir()
	mkdir(t, root, "personal")
	mkdir(t, root, "blstech")
	mkdir(t, root, "personal", "nested") // should NOT appear at depth 1
	mkdir(t, root, ".hidden")            // hidden dir, skip

	got, err := candidateDirs(root, 1, "/nonexistent-tries-path")
	if err != nil {
		t.Fatalf("candidateDirs: %v", err)
	}

	want := []string{
		filepath.Join(root, "blstech"),
		filepath.Join(root, "personal"),
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("depth=1 candidates:\n  got:  %v\n  want: %v", got, want)
	}
}

func TestCandidateDirs_Depth2(t *testing.T) {
	root := t.TempDir()
	mkdir(t, root, "personal")
	mkdir(t, root, "personal", "nested")
	mkdir(t, root, "blstech")

	got, err := candidateDirs(root, 2, "/nonexistent")
	if err != nil {
		t.Fatalf("candidateDirs: %v", err)
	}

	want := []string{
		filepath.Join(root, "blstech"),
		filepath.Join(root, "personal"),
		filepath.Join(root, "personal", "nested"),
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("depth=2 candidates:\n  got:  %v\n  want: %v", got, want)
	}
}

func TestCandidateDirs_SkipsTriesPath(t *testing.T) {
	// Promote root and tries path both under same parent — promote should
	// skip the tries dir entirely so the user can't promote-in-place.
	root := t.TempDir()
	mkdir(t, root, "personal")
	triesPath := mkdir(t, root, "tries")
	mkdir(t, triesPath, "2026-01-01-foo") // would be listed if we didn't skip

	got, err := candidateDirs(root, 2, triesPath)
	if err != nil {
		t.Fatalf("candidateDirs: %v", err)
	}
	for _, p := range got {
		if p == triesPath || strings.HasPrefix(p, triesPath+string(filepath.Separator)) {
			t.Errorf("candidateDirs returned a tries-path entry: %s", p)
		}
	}
}

func TestRun_DryRun_StripsDatePrefix(t *testing.T) {
	// Build a fake config layout:
	//   root/tries/2026-02-11-subtrackr-website/
	//   root/personal/
	root := t.TempDir()
	triesPath := mkdir(t, root, "tries")
	srcName := "2026-02-11-subtrackr-website"
	srcPath := mkdir(t, triesPath, srcName)
	mkdir(t, root, "personal")

	// Stat-touch the dir so mtime is sane (ScanEntries sorts by mtime).
	_ = os.Chtimes(srcPath, time.Now(), time.Now())

	cfg := config.Resolved{
		Path: triesPath,
		Promote: config.Promote{
			Root:   root,
			Depth:  1,
			Picker: "fzf",
		},
	}

	res, err := Run(cfg, Options{
		SourceName: srcName,
		DestDir:    "personal", // relative to root
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	wantDest := filepath.Join(root, "personal", "subtrackr-website")
	if res.DestPath != wantDest {
		t.Errorf("DestPath = %q, want %q", res.DestPath, wantDest)
	}
	if res.Performed {
		t.Errorf("Performed = true on DryRun")
	}
	// Source must still exist after a dry run.
	if _, err := os.Stat(srcPath); err != nil {
		t.Errorf("src vanished on DryRun: %v", err)
	}
}

func TestRun_KeepDatePreservesPrefix(t *testing.T) {
	root := t.TempDir()
	triesPath := mkdir(t, root, "tries")
	srcName := "2026-02-11-subtrackr-website"
	mkdir(t, triesPath, srcName)
	mkdir(t, root, "personal")

	cfg := config.Resolved{
		Path:    triesPath,
		Promote: config.Promote{Root: root, Depth: 1, Picker: "fzf"},
	}

	res, err := Run(cfg, Options{
		SourceName: srcName,
		DestDir:    "personal",
		KeepDate:   true,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := filepath.Join(root, "personal", srcName)
	if res.DestPath != want {
		t.Errorf("KeepDate DestPath = %q, want %q", res.DestPath, want)
	}
}

func TestRun_PerformsMove(t *testing.T) {
	root := t.TempDir()
	triesPath := mkdir(t, root, "tries")
	srcName := "2026-02-11-foo"
	srcPath := mkdir(t, triesPath, srcName)
	mkdir(t, root, "personal")

	// Drop a sentinel file inside src so we can confirm the move took the
	// contents along.
	sentinel := filepath.Join(srcPath, "marker.txt")
	if err := os.WriteFile(sentinel, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	cfg := config.Resolved{
		Path:    triesPath,
		Promote: config.Promote{Root: root, Depth: 1, Picker: "fzf"},
	}

	res, err := Run(cfg, Options{
		SourceName: srcName,
		DestDir:    "personal",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Performed {
		t.Errorf("Performed = false on non-dry-run")
	}
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Errorf("src still exists after move: err = %v", err)
	}
	wantDest := filepath.Join(root, "personal", "foo")
	if res.DestPath != wantDest {
		t.Errorf("DestPath = %q, want %q", res.DestPath, wantDest)
	}
	if _, err := os.Stat(filepath.Join(wantDest, "marker.txt")); err != nil {
		t.Errorf("sentinel did not survive move: %v", err)
	}
}

func TestRun_RefusesIfDestExists(t *testing.T) {
	root := t.TempDir()
	triesPath := mkdir(t, root, "tries")
	srcName := "2026-02-11-foo"
	mkdir(t, triesPath, srcName)
	// Pre-create the would-be destination.
	mkdir(t, root, "personal", "foo")

	cfg := config.Resolved{
		Path:    triesPath,
		Promote: config.Promote{Root: root, Depth: 1, Picker: "fzf"},
	}

	_, err := Run(cfg, Options{SourceName: srcName, DestDir: "personal", DryRun: true})
	if err == nil || !strings.Contains(err.Error(), "destination already exists") {
		t.Errorf("expected dest-exists error, got %v", err)
	}
}

func TestRun_UnknownSource(t *testing.T) {
	root := t.TempDir()
	triesPath := mkdir(t, root, "tries")
	mkdir(t, triesPath, "2026-02-11-real")

	cfg := config.Resolved{
		Path:    triesPath,
		Promote: config.Promote{Root: root, Depth: 1, Picker: "fzf"},
	}

	_, err := Run(cfg, Options{SourceName: "ghost", DestDir: "personal", DryRun: true})
	if err == nil || !strings.Contains(err.Error(), "no try directory matches") {
		t.Errorf("expected unknown-source error, got %v", err)
	}
}
