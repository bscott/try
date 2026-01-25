package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestClone(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	tests := []struct {
		name        string
		uri         string
		destPath    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Empty URI",
			uri:         "",
			destPath:    "/tmp/test",
			wantErr:     true,
			errContains: "empty URI",
		},
		{
			name:        "Empty destination",
			uri:         "https://github.com/user/repo.git",
			destPath:    "",
			wantErr:     true,
			errContains: "empty destination path",
		},
		{
			name:        "Invalid URI",
			uri:         "not-a-valid-uri",
			destPath:    "/tmp/test",
			wantErr:     true,
			errContains: "git clone failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Clone(tt.uri, tt.destPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Clone() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestCreateWorktree(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	tests := []struct {
		name         string
		repoPath     string
		worktreePath string
		branch       string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "Empty repo path",
			repoPath:     "",
			worktreePath: "/tmp/worktree",
			branch:       "test-branch",
			wantErr:      true,
			errContains:  "empty repository path",
		},
		{
			name:         "Empty worktree path",
			repoPath:     "/tmp/repo",
			worktreePath: "",
			branch:       "test-branch",
			wantErr:      true,
			errContains:  "empty worktree path",
		},
		{
			name:         "Empty branch name",
			repoPath:     "/tmp/repo",
			worktreePath: "/tmp/worktree",
			branch:       "",
			wantErr:      true,
			errContains:  "empty branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateWorktree(tt.repoPath, tt.worktreePath, tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CreateWorktree() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestRemoveWorktree(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	tests := []struct {
		name         string
		repoPath     string
		worktreePath string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "Empty repo path",
			repoPath:     "",
			worktreePath: "/tmp/worktree",
			wantErr:      true,
			errContains:  "empty repository path",
		},
		{
			name:         "Empty worktree path",
			repoPath:     "/tmp/repo",
			worktreePath: "",
			wantErr:      true,
			errContains:  "empty worktree path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RemoveWorktree(tt.repoPath, tt.worktreePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("RemoveWorktree() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestListWorktrees(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	tests := []struct {
		name        string
		repoPath    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Empty repo path",
			repoPath:    "",
			wantErr:     true,
			errContains: "empty repository path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ListWorktrees(tt.repoPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListWorktrees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ListWorktrees() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestIsGitRepository(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	t.Run("Non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		if IsGitRepository(tmpDir) {
			t.Error("IsGitRepository() should return false for non-git directory")
		}
	})

	t.Run("Git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize a git repo
		cmd := exec.Command("git", "init", tmpDir)
		if err := cmd.Run(); err != nil {
			t.Skipf("Could not initialize test git repo: %v", err)
		}

		if !IsGitRepository(tmpDir) {
			t.Error("IsGitRepository() should return true for git directory")
		}
	})
}

// Integration test for the full workflow
func TestGitWorkflow_Integration(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	// Create a temporary directory for our test
	tmpDir := t.TempDir()

	// Create a test repository
	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test repo directory: %v", err)
	}

	// Initialize the repository
	cmd := exec.Command("git", "init", repoDir)
	if err := cmd.Run(); err != nil {
		t.Skipf("Could not initialize test git repo: %v", err)
	}

	// Create an initial commit
	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "-C", repoDir, "add", "README.md")
	if err := cmd.Run(); err != nil {
		t.Skipf("Could not add file: %v", err)
	}

	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Skipf("Could not create initial commit: %v", err)
	}

	// Test IsGitRepository
	if !IsGitRepository(repoDir) {
		t.Error("IsGitRepository() should return true for initialized repo")
	}

	// Test CreateWorktree with new branch
	worktreeDir := filepath.Join(tmpDir, "test-worktree")
	if err := CreateWorktree(repoDir, worktreeDir, "feature-branch"); err != nil {
		t.Errorf("CreateWorktree() failed: %v", err)
	}

	// Verify worktree was created
	if !IsGitRepository(worktreeDir) {
		t.Error("Worktree should be a valid git repository")
	}

	// Test ListWorktrees
	worktrees, err := ListWorktrees(repoDir)
	if err != nil {
		t.Errorf("ListWorktrees() failed: %v", err)
	}

	if len(worktrees) != 2 { // Main repo + worktree
		t.Errorf("Expected 2 worktrees, got %d", len(worktrees))
	}

	// Test RemoveWorktree
	if err := RemoveWorktree(repoDir, worktreeDir); err != nil {
		t.Errorf("RemoveWorktree() failed: %v", err)
	}

	// Verify worktree was removed
	worktrees, err = ListWorktrees(repoDir)
	if err != nil {
		t.Errorf("ListWorktrees() after removal failed: %v", err)
	}

	if len(worktrees) != 1 { // Only main repo should remain
		t.Errorf("Expected 1 worktree after removal, got %d", len(worktrees))
	}
}
