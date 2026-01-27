# Claude Code Guidelines for `try` CLI

## Project Overview

The `try` CLI is a Go implementation of a temporary directory manager with an interactive TUI. It helps developers quickly create, manage, and navigate ephemeral workspaces for experiments.

**Module:** `github.com/bscott/try`

## Architecture

```
try/
├── cmd/try/              # Main entry point (go install target)
│   ├── main.go          # Basic list command, Cobra setup
│   ├── exec.go          # TUI launcher with stdout/stderr handling
│   └── init.go          # Shell integration generator
├── internal/
│   ├── config/          # TRY_PATH configuration
│   ├── tries/           # Directory operations (Entry, Manager, Scanner, GitStatus)
│   ├── fuzzy/           # Fuzzy matching algorithm (98.3% coverage)
│   ├── tui/             # Bubble Tea TUI (Model, View, Update, Styles, Keys)
│   ├── git/             # Git operations (clone, worktree, parser)
│   └── shell/           # Shell integration (bash/zsh/fish)
```

## Key Design Principles

### 1. Security First
- **Shell command injection prevention**: All paths passed to shell are quoted using `shellQuote()` function
- **Path traversal protection**: Entry names validated to reject `..`, `.`, `/`, `\`, and null bytes
- **Atomic operations**: Use `os.Mkdir` (not `os.MkdirAll`) to prevent TOCTOU races
- **Input validation**: All user input sanitized before file operations

### 2. TUI Output Isolation
**Critical**: The TUI must never leak ANSI escape sequences to stdout when running in shell command substitution.

In `cmd/try/exec.go`:
```go
// Save original stdout and redirect stdout to stderr during TUI
origStdout := os.Stdout
os.Stdout = os.Stderr

// Run Bubble Tea TUI
p := tea.NewProgram(model, tea.WithAltScreen())
finalModel, err := p.Run()

// Restore stdout BEFORE writing command
os.Stdout = origStdout

// Only the cd command goes to stdout
fmt.Fprint(os.Stdout, output)
```

### 3. Fuzzy Matching Algorithm
The fuzzy matcher in `internal/fuzzy/` is a direct port of the Ruby implementation with exact scoring:

- Base score: `100.0 / (days_since_mtime + 1.0)`
- Date prefix bonus: `+10.0`
- Character match bonuses with proximity scoring
- Maintains 98.3% test coverage

**Do not modify** the scoring algorithm without corresponding tests.

### 4. Git Status Integration
Git status checking in `internal/tries/gitstatus.go`:
- Shells out to system `git` command (no go-git library)
- Checks: branch, ahead/behind, dirty status, last commit time
- Gracefully handles non-git directories
- Results cached in Entry struct during scan

## Common Tasks

### Adding a New TUI Mode

1. Add mode constant to `internal/tui/model.go`:
   ```go
   const (
       ModeNormal Mode = iota
       ModeYourNew
   )
   ```

2. Add handler in `internal/tui/update.go`:
   ```go
   func (m Model) handleYourNewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd)
   ```

3. Add view in `internal/tui/view.go`:
   ```go
   func (m *Model) renderYourNewView(b *strings.Builder)
   ```

4. Update mode switch in `handleKeyPress()` and `renderNormalView()`

### Adding Git Operations

1. Add function to `internal/git/operations.go`
2. Use `exec.Command("git", ...)` - capture stderr for errors
3. Add tests in `internal/git/operations_test.go`
4. Document in `internal/git/README.md`

### Modifying Directory Entry Display

- Column widths defined in `internal/tui/view.go` constants
- NAME column: 50 chars (with truncation)
- Headers in `renderColumnHeaders()`
- Entry rendering in `renderEntry()`
- Styles defined in `internal/tui/styles.go`

## Testing Guidelines

### Required Tests
- All new fuzzy matching logic must have corresponding tests
- Git operations should have both unit and integration tests
- TUI helper functions need test coverage

### Running Tests
```bash
go test ./...                    # All tests
go test -v -race ./...          # With race detection
go test -coverprofile=c.out ./...  # Coverage report
```

### Test Structure
```go
func TestFeature(t *testing.T) {
    // Setup
    // Execute
    // Assert
    // Cleanup if needed
}
```

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Keep functions small and focused
- Prefer explicit error handling over panic
- Use meaningful variable names (not single letters except in loops)
- Add comments for non-obvious logic

## Common Pitfalls

### ❌ Don't
```go
// Don't use MkdirAll for new entries (TOCTOU)
os.MkdirAll(dirPath, 0755)

// Don't output directly from TUI when in exec mode
fmt.Println("message")

// Don't use strings.Split on git output without checking length
parts := strings.Split(output, " ")
parts[1] // Could panic!
```

### ✅ Do
```go
// Use Mkdir for atomic creation
if err := os.Mkdir(dirPath, 0755); err != nil {
    if os.IsExist(err) {
        return fmt.Errorf("already exists")
    }
    return err
}

// Set model state, output in View()
m.statusMsg = "message"

// Check length before accessing
parts := strings.Split(output, " ")
if len(parts) < 2 {
    return "", fmt.Errorf("invalid format")
}
```

## Debugging

### TUI Not Showing Output
- Check that `os.Stdout` is redirected to `os.Stderr` during TUI run
- Verify `View()` returns empty string on exit
- Ensure final command goes to original stdout

### Git Status Not Appearing
- Check if `.git` directory exists
- Verify git commands work in terminal: `cd <dir> && git status`
- Look for errors in git command stderr output

### Fuzzy Search Not Working
- Verify entries are loaded: check `m.allEntries` length
- Check `updateSearch()` is called on input change
- Ensure `fuzzy.Match()` returns non-empty results

## Release Process

1. Test changes locally: `go build ./cmd/try && ./try`
2. Run all tests: `go test ./...`
3. Update version in commit message
4. Commit: `git commit -m "description"`
5. Tag: `git tag -a vX.Y.Z -m "message"`
6. Push: `git push origin main && git push origin vX.Y.Z`

## Dependencies

Key dependencies (see `go.mod`):
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - Reusable components
- `github.com/spf13/cobra` - CLI framework
- `github.com/adrg/xdg` - XDG directory support

## Environment Variables

- `TRY_PATH` - Base directory for try directories (default: `~/src/tries`)
- `SHELL` - Used by init command to detect shell type

## Shell Integration

The wrapper function in bash/zsh:
```bash
try() {
    local result
    result=$(command try exec "$@" 2>/dev/tty)
    if [[ $? -eq 0 ]]; then
        eval "$result"
    fi
}
export TRY_PATH="$HOME/src/tries"
```

- `2>/dev/tty` - Redirects TUI (stderr) to terminal
- `$()` - Captures stdout only (the cd command)
- `eval "$result"` - Executes the cd command in current shell

## Future Improvements

Potential enhancements (not prioritized):
- [ ] Config file support (~/.config/try/config.toml)
- [ ] Remote sync (git push/pull for try directories)
- [ ] Tags/labels for organizing tries
- [ ] Search history persistence
- [ ] Template system for new tries
- [ ] Integration with task managers
- [ ] Export/import try directories
