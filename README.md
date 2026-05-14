# try - Temporary Project Directory Manager

A Go implementation of the `try` CLI tool for managing temporary project directories, with full feature parity to the original Ruby version.

## Features

- **Interactive TUI**: Fuzzy search and navigate your try directories with a beautiful terminal interface
- **Smart Search**: Fuzzy matching with intelligent scoring based on recency and naming patterns
- **Date Prefixes**: Automatically prefix directories with `YYYY-MM-DD-` for easy organization
- **Directory Operations**: Create, rename, and delete try directories
- **Git Integration**: Clone repositories and create worktrees directly into try directories
- **Shell Integration**: Seamless `cd` navigation through shell wrapper functions
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Installation

### Using Go Install

```bash
go install github.com/bscott/try/cmd/try@latest
```

### From Source

```bash
git clone https://github.com/bscott/try.git
cd try
go build -o try ./cmd/try
sudo mv try /usr/local/bin/
```

## Quick Start

1. **Initialize shell integration**:

```bash
# For bash/zsh (add to ~/.bashrc or ~/.zshrc)
eval "$(try init)"

# For fish (add to ~/.config/fish/config.fish)
try init | source
```

2. **Use the TUI**:

```bash
# Launch interactive TUI
try

# The TUI opens with fuzzy search - just start typing!
```

## Usage

### Commands

- `try` - Launch interactive TUI
- `try init [path]` - Generate shell integration script
- `try exec` - Execute TUI and output shell command (used by wrapper)
- `try config show` - Print the effective tries path and which source (env / config file / default) it came from

### Keyboard Shortcuts

#### Navigation
- `↑/↓` or `Ctrl+P/N` - Move selection up/down
- `Enter` - Select entry and navigate to directory

#### Search Input
- `Ctrl+A` - Move cursor to start
- `Ctrl+E` - Move cursor to end
- `Ctrl+B` - Move cursor back one character
- `Ctrl+F` - Move cursor forward one character
- `Ctrl+K` - Kill to end of line
- `Ctrl+W` - Delete word backward
- `Backspace` - Delete character backward

#### Actions
- `Ctrl+D` - Toggle delete mode / mark entry for deletion
- `Ctrl+R` - Rename selected entry
- `Ctrl+T` - Create new try directory
- `Ctrl+H` - Show help
- `Esc` - Cancel current action / return to normal mode
- `Ctrl+C` - Quit

### Delete Flow

1. Press `Ctrl+D` to enter delete mode
2. Press `Ctrl+D` again to mark/unmark entries (✗ indicator appears)
3. Press `Enter` to show confirmation prompt
4. Type `YES` exactly to confirm deletion
5. Press `Esc` to cancel at any point

### Create Flow

1. Press `Ctrl+T` to open create prompt
2. Enter name (without date prefix)
3. Press `Enter` to create
4. Directory is created as `YYYY-MM-DD-{name}` and you're navigated to it

### Rename Flow

1. Select entry and press `Ctrl+R`
2. Edit the name (date prefix is preserved)
3. Press `Enter` to confirm or `Esc` to cancel

## Configuration

The tries base directory is resolved with the following precedence (highest wins):

1. **`$TRY_PATH`** environment variable
2. **`tries_path`** in the config file at `~/.config/try/config.toml` (or `$XDG_CONFIG_HOME/try/config.toml` if set)
3. **Default**: `~/src/tries`

Inspect what's currently in effect with `try config show`.

### Config file

```toml
# ~/.config/try/config.toml
tries_path = "~/code/tries"
```

`~` is expanded to the user's home directory. The directory is created on first use if it doesn't already exist.

### Environment variable (overrides the config file)

```bash
export TRY_PATH="$HOME/experiments"
eval "$(try init)"
```

## Git Integration

Clone repositories directly into try directories:

```bash
try clone https://github.com/user/repo.git
```

Create worktrees:

```bash
try worktree feature-branch
```

## Fuzzy Matching Algorithm

The fuzzy matcher uses intelligent scoring:

1. **Recency**: Newer directories score higher
2. **Date Prefix Bonus**: Entries with `YYYY-MM-DD-` prefix get +10 points
3. **Character Matches**: +1 per matched character
4. **Word Boundaries**: +1 for matches at word start
5. **Proximity**: Closer matches score higher
6. **Density**: Tighter match spans score higher
7. **Length Penalty**: Shorter names score slightly higher

Empty queries show all entries sorted by modification time (newest first).

## Project Structure

```
try/
├── cmd/try/              # Main entry point
│   ├── main.go
│   ├── exec.go           # TUI launcher
│   └── init.go           # Shell integration
├── internal/
│   ├── config/           # Configuration management
│   ├── tries/            # Directory operations
│   ├── fuzzy/            # Fuzzy matching algorithm
│   ├── tui/              # Bubble Tea TUI
│   ├── git/              # Git operations
│   └── shell/            # Shell integration
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o try ./cmd/try
```

## License

MIT License - see LICENSE file for details

## Credits

- Original Ruby implementation: [tobi/try](https://github.com/tobi/try)
- TUI framework: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styling: [Lipgloss](https://github.com/charmbracelet/lipgloss)
