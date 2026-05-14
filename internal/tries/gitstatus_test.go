package tries

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitInit shells out to git to build a real repo on disk. We deliberately use
// the system git (matching production behavior) instead of go-git.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	mustRun(t, dir, "git", "init", "-q", "-b", "main")
	mustRun(t, dir, "git", "config", "user.email", "test@example.com")
	mustRun(t, dir, "git", "config", "user.name", "Test")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}

func commitFile(t *testing.T, dir, name, body, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mustRun(t, dir, "git", "add", name)
	mustRun(t, dir, "git", "commit", "-q", "-m", msg)
}

func TestGetGitStatus_NonGitDir(t *testing.T) {
	dir := t.TempDir()
	s := GetGitStatus(dir)
	if s.IsRepo {
		t.Errorf("IsRepo = true for non-git dir %s", dir)
	}
	if s.Branch != "" || s.IsDirty || s.Ahead != 0 || s.Behind != 0 {
		t.Errorf("non-zero fields on non-git: %+v", s)
	}
}

func TestGetGitStatus_CleanRepo(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	commitFile(t, dir, "README.md", "hi", "initial")

	s := GetGitStatus(dir)
	if !s.IsRepo {
		t.Fatal("IsRepo = false on initialized repo")
	}
	if s.Branch != "main" {
		t.Errorf("Branch = %q, want main", s.Branch)
	}
	if s.IsDirty {
		t.Error("IsDirty = true on clean repo")
	}
	if s.LastCommit.IsZero() {
		t.Error("LastCommit is zero on repo with one commit")
	}
}

func TestGetGitStatus_DirtyWorkingTree(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	commitFile(t, dir, "a.txt", "first", "init")
	// Untracked file should also count as dirty.
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	s := GetGitStatus(dir)
	if !s.IsDirty {
		t.Error("IsDirty = false despite untracked file")
	}
}

func TestGetGitStatus_AheadOfRemote(t *testing.T) {
	// Set up a bare "remote" and a local repo tracking it.
	remote := t.TempDir()
	mustRun(t, remote, "git", "init", "-q", "--bare", "-b", "main")

	local := t.TempDir()
	gitInit(t, local)
	mustRun(t, local, "git", "remote", "add", "origin", remote)
	commitFile(t, local, "a.txt", "1", "first")
	mustRun(t, local, "git", "push", "-q", "-u", "origin", "main")
	commitFile(t, local, "b.txt", "2", "second")

	s := GetGitStatus(local)
	if s.Ahead != 1 {
		t.Errorf("Ahead = %d, want 1", s.Ahead)
	}
	if s.Behind != 0 {
		t.Errorf("Behind = %d, want 0", s.Behind)
	}
	if s.RemoteBranch != "origin/main" {
		t.Errorf("RemoteBranch = %q, want origin/main", s.RemoteBranch)
	}
}

func TestGetGitStatus_NoUpstream(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	commitFile(t, dir, "a.txt", "1", "init")

	s := GetGitStatus(dir)
	if s.RemoteBranch != "" {
		t.Errorf("RemoteBranch = %q on repo with no upstream, want empty", s.RemoteBranch)
	}
	if s.Ahead != 0 || s.Behind != 0 {
		t.Errorf("Ahead/Behind nonzero with no upstream: %+v", s)
	}
}
