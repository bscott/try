// Package promote moves a try directory out of the tries dir into a
// destination under the configured promote root (typically ~/code), stripping
// the YYYY-MM-DD- date prefix in the process.
//
// The intent: a directory in `tries/` graduates from "scratch experiment" to
// "real project" without manual mv + prefix-trim + cd.
package promote

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bscott/try/internal/config"
	"github.com/bscott/try/internal/tries"
)

// Options controls a single promote operation. Zero values mean "use the
// resolved config defaults."
type Options struct {
	// SourceName, if non-empty, names the try entry to promote (by directory
	// basename inside the tries dir). Empty means "pick interactively."
	SourceName string

	// DestRoot overrides config.Promote.Root for this call. Empty = use config.
	DestRoot string

	// DestDir, if non-empty, skips destination fuzzy-pick and uses this dir
	// (resolved relative to DestRoot/config root if relative, otherwise
	// honored as-is).
	DestDir string

	// MaxDepth overrides config.Promote.Depth for candidate enumeration.
	// Zero = use config.
	MaxDepth int

	// KeepDate, if true, preserves the YYYY-MM-DD- date prefix on the final
	// directory name. Default behavior strips it.
	KeepDate bool

	// DryRun describes the planned move without performing it.
	DryRun bool
}

// Result describes the outcome of a successful promote. SrcPath is the
// original tries path (no longer exists after a non-dry-run promote);
// DestPath is where the directory now lives.
type Result struct {
	SrcPath   string
	DestPath  string
	Performed bool // false when DryRun was true
}

// Run executes the promote workflow against the given resolved config.
// It returns the source/destination paths so callers (the Cobra cmd) can
// print them and optionally emit a `cd` to stdout for the shell wrapper.
func Run(cfg config.Resolved, opts Options) (Result, error) {
	root := opts.DestRoot
	if root == "" {
		root = cfg.Promote.Root
	}
	depth := opts.MaxDepth
	if depth <= 0 {
		depth = cfg.Promote.Depth
	}
	picker := cfg.Promote.Picker

	// 1. Pick source.
	manager := tries.NewManager(cfg.Path)
	src, err := pickSource(manager, opts.SourceName, picker)
	if err != nil {
		return Result{}, err
	}

	// 2. Pick destination parent dir.
	destParent, err := pickDestParent(root, depth, picker, opts.DestDir, cfg.Path)
	if err != nil {
		return Result{}, err
	}

	// 3. Compute final destination path.
	finalName := src.DisplayName()
	if opts.KeepDate {
		finalName = src.Name
	}
	destPath := filepath.Join(destParent, finalName)

	if _, err := os.Stat(destPath); err == nil {
		return Result{}, fmt.Errorf("destination already exists: %s", destPath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Result{}, fmt.Errorf("checking destination: %w", err)
	}

	res := Result{SrcPath: src.Path, DestPath: destPath}
	if opts.DryRun {
		return res, nil
	}

	// 4. Make sure the destination's parent exists. We use MkdirAll because
	// the immediate parent (destParent) was selected from existing dirs, but
	// it's defensive against picker returning a path with a missing leg.
	if err := os.MkdirAll(destParent, 0o755); err != nil {
		return Result{}, fmt.Errorf("ensuring destination parent: %w", err)
	}

	if err := move(src.Path, destPath); err != nil {
		return Result{}, err
	}
	res.Performed = true
	return res, nil
}

// pickSource resolves which try entry to promote. With a name, it looks up
// the matching entry; without one, it shells out to fzf over all entries.
func pickSource(manager *tries.Manager, name, picker string) (tries.Entry, error) {
	entries, err := manager.List()
	if err != nil {
		return tries.Entry{}, fmt.Errorf("listing tries: %w", err)
	}
	if len(entries) == 0 {
		return tries.Entry{}, errors.New("no try directories to promote")
	}

	if name != "" {
		for _, e := range entries {
			if e.Name == name || e.DisplayName() == name {
				return e, nil
			}
		}
		return tries.Entry{}, fmt.Errorf("no try directory matches %q", name)
	}

	labels := make([]string, len(entries))
	for i, e := range entries {
		labels[i] = e.Name
	}
	picked, err := runPicker(picker, "Promote which try?", labels)
	if err != nil {
		return tries.Entry{}, err
	}
	for _, e := range entries {
		if e.Name == picked {
			return e, nil
		}
	}
	return tries.Entry{}, fmt.Errorf("picker returned unknown entry: %q", picked)
}

// pickDestParent resolves the directory under which the promoted dir will
// live. With an explicit destDir, it's joined onto root (relative) or used
// as-is (absolute). Without one, it shells out to a fuzzy picker over
// candidate dirs walked from root to depth.
func pickDestParent(root string, depth int, picker, destDir, triesPath string) (string, error) {
	if destDir != "" {
		if filepath.IsAbs(destDir) {
			return filepath.Clean(destDir), nil
		}
		return filepath.Join(root, destDir), nil
	}

	candidates, err := candidateDirs(root, depth, triesPath)
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no candidate destination dirs under %s (depth %d)", root, depth)
	}

	// Display relative paths to the user; map back to absolute.
	relToAbs := make(map[string]string, len(candidates))
	labels := make([]string, len(candidates))
	for i, abs := range candidates {
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			rel = abs
		}
		labels[i] = rel
		relToAbs[rel] = abs
	}
	picked, err := runPicker(picker, "Promote into: "+root, labels)
	if err != nil {
		return "", err
	}
	abs, ok := relToAbs[picked]
	if !ok {
		return "", fmt.Errorf("picker returned unknown destination: %q", picked)
	}
	return abs, nil
}

// candidateDirs walks root to maxDepth and returns directories the user
// could plausibly promote into. Hidden dirs (starting with ".") are skipped,
// as is the tries dir itself (don't promote-in-place).
//
// Depth semantics: depth=1 returns direct subdirs of root; depth=2 also
// returns their immediate children, etc.
func candidateDirs(root string, maxDepth int, triesPath string) ([]string, error) {
	root = filepath.Clean(root)
	triesPath = filepath.Clean(triesPath)
	rootInfo, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("promote root: %w", err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("promote root is not a directory: %s", root)
	}

	var out []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}
		if path == triesPath {
			return fs.SkipDir
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// depth = number of path separators in rel + 1 for first-level entries.
		depth := strings.Count(rel, string(filepath.Separator)) + 1
		out = append(out, path)
		if depth >= maxDepth {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// runPicker invokes the configured fuzzy picker (today: fzf) with the given
// labels on stdin and returns the selected line. The picker uses /dev/tty
// for its UI; stdin/stdout are pipes used to feed candidates and read the
// selection.
func runPicker(picker, prompt string, labels []string) (string, error) {
	if picker == "" {
		picker = config.DefaultPromotePicker
	}
	if picker != "fzf" {
		return "", fmt.Errorf("unsupported picker %q (only \"fzf\" is supported today)", picker)
	}
	bin, err := exec.LookPath("fzf")
	if err != nil {
		return "", fmt.Errorf("fzf not found on PATH — install it (`brew install fzf`) or pass a destination flag")
	}

	cmd := exec.Command(bin,
		"--prompt="+prompt+"> ",
		"--height=40%",
		"--reverse",
		"--no-multi",
	)
	cmd.Stdin = strings.NewReader(strings.Join(labels, "\n") + "\n")
	// fzf uses /dev/tty for its UI — we just need stderr passed through so
	// any error messages from fzf reach the user.
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		// fzf returns exit code 130 on Ctrl-C / no selection. Surface that as
		// a clean "cancelled" error rather than a noisy exec error.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return "", errors.New("cancelled")
		}
		return "", fmt.Errorf("fzf: %w", err)
	}
	picked := strings.TrimSpace(string(out))
	if picked == "" {
		return "", errors.New("cancelled")
	}
	return picked, nil
}

// move renames src to dst, falling back to `mv` for cross-filesystem moves
// (which os.Rename rejects with EXDEV). We shell out to mv rather than
// copy-and-delete in Go so we inherit its handling of permissions, hard
// links, sparse files, and partial-failure cleanup.
func move(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !isCrossDevice(err) {
		return fmt.Errorf("rename %s → %s: %w", src, dst, err)
	}

	// Cross-filesystem: defer to system mv. `-n` refuses to overwrite (we
	// already stat-checked, but belt-and-suspenders).
	cmd := exec.Command("mv", "-n", src, dst)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mv %s → %s: %w", src, dst, err)
	}
	return nil
}

func isCrossDevice(err error) bool {
	var linkErr *os.LinkError
	if !errors.As(err, &linkErr) {
		return false
	}
	return strings.Contains(linkErr.Err.Error(), "cross-device")
}
