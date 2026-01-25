# Git Package

The `git` package provides utilities for working with Git repositories by shelling out to the system git command.

## Features

- **URI Parsing**: Parse various Git URI formats (HTTPS, SSH, etc.)
- **Clone Operations**: Clone repositories to local paths
- **Worktree Management**: Create, remove, and list Git worktrees
- **Repository Detection**: Check if a path is a Git repository

## Installation

This package is part of the `try` CLI tool. No separate installation is required.

## URI Parser

Parse Git repository URIs in various formats:

```go
import "github.com/bscott/try/internal/git"

// Parse HTTPS URL
info, err := git.ParseGitURI("https://github.com/user/repo.git")
if err != nil {
    log.Fatal(err)
}

fmt.Println(info.Host)   // "github.com"
fmt.Println(info.Owner)  // "user"
fmt.Println(info.Repo)   // "repo"

// Parse SSH URL
info, err = git.ParseGitURI("git@github.com:user/repo.git")

// Get clone URLs
fmt.Println(info.CloneURL())  // "https://github.com/user/repo.git"
fmt.Println(info.SSHURL())    // "git@github.com:user/repo.git"
```

### Supported URI Formats

- `https://github.com/user/repo.git` - HTTPS with .git
- `https://github.com/user/repo` - HTTPS without .git
- `git@github.com:user/repo.git` - SSH with .git
- `git@github.com:user/repo` - SSH without .git
- `ssh://git@github.com:user/repo.git` - SSH protocol with .git
- Works with GitHub, GitLab, and any custom Git host

## Git Operations

### Clone a Repository

```go
err := git.Clone("https://github.com/user/repo.git", "/path/to/destination")
if err != nil {
    log.Fatal(err)
}
```

### Create a Worktree

```go
// Creates a new worktree with a new branch
err := git.CreateWorktree("/path/to/repo", "/path/to/worktree", "feature-branch")
if err != nil {
    log.Fatal(err)
}
```

If the branch already exists, it will be checked out. Otherwise, a new branch will be created.

### Remove a Worktree

```go
err := git.RemoveWorktree("/path/to/repo", "/path/to/worktree")
if err != nil {
    log.Fatal(err)
}
```

### List Worktrees

```go
worktrees, err := git.ListWorktrees("/path/to/repo")
if err != nil {
    log.Fatal(err)
}

for _, wt := range worktrees {
    fmt.Println(wt)
}
```

### Check if Path is a Git Repository

```go
if git.IsGitRepository("/path/to/check") {
    fmt.Println("This is a git repository")
} else {
    fmt.Println("Not a git repository")
}
```

## Error Handling

All functions return descriptive errors that include git command output when operations fail:

```go
err := git.Clone("invalid-uri", "/tmp/dest")
if err != nil {
    // Error will include:
    // - What went wrong
    // - Git command output/stderr
    log.Printf("Clone failed: %v", err)
}
```

## Implementation Details

This package shells out to the system `git` command using `os/exec` rather than using the go-git library. This approach:

- Ensures compatibility with all Git features and configurations
- Respects system Git configuration and credentials
- Provides familiar error messages from Git itself
- Requires Git to be installed on the system

## Testing

Run the test suite:

```bash
go test ./internal/git/
```

Run tests with verbose output:

```bash
go test -v ./internal/git/
```

The test suite includes:
- Unit tests for URI parsing (21 test cases)
- Unit tests for all operations
- Integration tests for the full Git workflow
- Example tests demonstrating usage

## License

Part of the try CLI tool - see LICENSE file in repository root.
