// Package git provides utilities for working with Git repositories.
//
// This package shells out to the system git command rather than using go-git
// library, ensuring compatibility with all git features and configurations.
//
// # Parser
//
// The parser subpackage provides URI parsing for Git repositories:
//
//	info, err := git.ParseGitURI("git@github.com:user/repo.git")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(info.Host)   // "github.com"
//	fmt.Println(info.Owner)  // "user"
//	fmt.Println(info.Repo)   // "repo"
//
// Supported URI formats:
//   - HTTPS: https://github.com/user/repo.git
//   - HTTPS without .git: https://github.com/user/repo
//   - SSH: git@github.com:user/repo.git
//   - SSH without .git: git@github.com:user/repo
//   - SSH protocol: ssh://git@github.com:user/repo.git
//
// # Operations
//
// The operations subpackage provides functions for common git operations:
//
// Clone a repository:
//
//	err := git.Clone("https://github.com/user/repo.git", "/path/to/dest")
//
// Create a worktree:
//
//	err := git.CreateWorktree("/path/to/repo", "/path/to/worktree", "feature-branch")
//
// Remove a worktree:
//
//	err := git.RemoveWorktree("/path/to/repo", "/path/to/worktree")
//
// List all worktrees:
//
//	worktrees, err := git.ListWorktrees("/path/to/repo")
//
// Check if a path is a git repository:
//
//	if git.IsGitRepository("/path/to/check") {
//		fmt.Println("This is a git repository")
//	}
//
// # Error Handling
//
// All functions return descriptive errors that include the git command output
// when operations fail, making debugging easier.
package git
