package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Clone clones a Git repository from the given URI to the destination path.
// This function shells out to the system git command.
func Clone(uri, destPath string) error {
	if uri == "" {
		return fmt.Errorf("empty URI provided")
	}
	if destPath == "" {
		return fmt.Errorf("empty destination path provided")
	}

	cmd := exec.Command("git", "clone", uri, destPath)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include stderr in error message for debugging
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// CreateWorktree creates a new Git worktree at the specified path.
// If the branch exists, it checks it out. If not, it creates a new branch.
func CreateWorktree(repoPath, worktreePath, branch string) error {
	if repoPath == "" {
		return fmt.Errorf("empty repository path provided")
	}
	if worktreePath == "" {
		return fmt.Errorf("empty worktree path provided")
	}
	if branch == "" {
		return fmt.Errorf("empty branch name provided")
	}

	// First, check if the branch exists
	checkCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", branch)
	err := checkCmd.Run()

	var cmd *exec.Cmd
	if err != nil {
		// Branch doesn't exist, create it
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath, "-b", branch)
	} else {
		// Branch exists, check it out
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath, branch)
	}

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include stderr in error message for debugging
		return fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// RemoveWorktree removes a Git worktree at the specified path.
func RemoveWorktree(repoPath, worktreePath string) error {
	if repoPath == "" {
		return fmt.Errorf("empty repository path provided")
	}
	if worktreePath == "" {
		return fmt.Errorf("empty worktree path provided")
	}

	cmd := exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include stderr in error message for debugging
		return fmt.Errorf("git worktree remove failed: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// ListWorktrees lists all worktrees for a repository.
func ListWorktrees(repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("empty repository path provided")
	}

	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	// Parse porcelain output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var worktrees []string

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			worktrees = append(worktrees, path)
		}
	}

	return worktrees, nil
}

// IsGitRepository checks if the given path is a Git repository.
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
