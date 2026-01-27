package tries

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitStatus holds git repository status information
type GitStatus struct {
	IsRepo       bool      // Whether directory is a git repository
	Branch       string    // Current branch name
	LastCommit   time.Time // Last commit date
	Ahead        int       // Commits ahead of remote
	Behind       int       // Commits behind remote
	IsDirty      bool      // Whether working tree has uncommitted changes
	RemoteBranch string    // Remote tracking branch (e.g., "origin/main")
}

// GetGitStatus checks if a directory is a git repository and gathers status info
func GetGitStatus(path string) GitStatus {
	status := GitStatus{IsRepo: false}

	// Check if .git directory exists
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return status
	}

	status.IsRepo = true

	// Get current branch
	if branch := gitCommand(path, "rev-parse", "--abbrev-ref", "HEAD"); branch != "" {
		status.Branch = branch
	}

	// Get last commit date
	if commitDate := gitCommand(path, "log", "-1", "--format=%cI"); commitDate != "" {
		if t, err := time.Parse(time.RFC3339, commitDate); err == nil {
			status.LastCommit = t
		}
	}

	// Get remote tracking branch
	if remote := gitCommand(path, "rev-parse", "--abbrev-ref", "@{upstream}"); remote != "" {
		status.RemoteBranch = remote

		// Get ahead/behind counts
		if counts := gitCommand(path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}"); counts != "" {
			parts := strings.Fields(counts)
			if len(parts) == 2 {
				// Parse ahead/behind counts
				var ahead, behind int
				stringToInt(parts[0], &ahead)
				stringToInt(parts[1], &behind)
				status.Ahead = ahead
				status.Behind = behind
			}
		}
	}

	// Check if working tree is dirty
	if diff := gitCommand(path, "status", "--porcelain"); diff != "" {
		status.IsDirty = true
	}

	return status
}

// gitCommand executes a git command in the given directory and returns trimmed output
func gitCommand(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// stringToInt converts a string to int
func stringToInt(s string, dest *int) error {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err == nil {
		*dest = val
	}
	return err
}
