# Shell Integration

This package provides shell integration for the `try` CLI tool, enabling seamless interaction with the TUI output.

## Overview

The shell integration works by wrapping the `try` binary in a shell function that:
1. Executes the actual `try` binary with the `exec` subcommand
2. Captures stdout (which contains shell commands like `cd /path/to/dir`)
3. Evaluates the captured output in the current shell session
4. Redirects stderr to `/dev/tty` so TUI output is visible

## Supported Shells

- **Bash** - POSIX-compatible shell
- **Zsh** - Enhanced bash-compatible shell
- **Fish** - User-friendly shell with different syntax

## Installation

### Quick Setup

```bash
# Bash/Zsh
eval "$(try init)"

# Fish
try init | source
```

### Permanent Setup

Add the eval line to your shell's configuration file:

**Bash** (`~/.bashrc` or `~/.bash_profile`):
```bash
eval "$(try init ~/src/tries)"
```

**Zsh** (`~/.zshrc`):
```bash
eval "$(try init ~/src/tries)"
```

**Fish** (`~/.config/fish/config.fish`):
```fish
try init ~/src/tries | source
```

## Usage

### Default Path
Uses `$TRY_PATH` environment variable or defaults to `~/src/tries`:
```bash
try init
```

### Custom Path
Specify a custom base path:
```bash
try init ~/my/custom/tries
```

### Explicit Shell Type
Force a specific shell type (useful in edge cases):
```bash
try init --shell bash
try init --shell zsh
try init --shell fish
```

## How It Works

### Shell Function Template (Bash/Zsh)
```bash
try() {
    local result
    result=$(command try exec "$@" 2>/dev/tty)
    if [[ $? -eq 0 ]]; then
        eval "$result"
    fi
}
```

### Shell Function Template (Fish)
```fish
function try
    set result (command try exec $argv 2>/dev/tty)
    if test $status -eq 0
        eval $result
    end
end
```

### Key Components

1. **`command try exec`** - Bypasses the shell function and calls the actual binary
2. **`2>/dev/tty`** - Redirects stderr (TUI output) to the terminal
3. **`eval "$result"`** - Executes the captured stdout (e.g., `cd` commands)
4. **Exit code check** - Only evaluates output if the command succeeded

## API

### Functions

#### `DetectShell(shellEnv string) ShellType`
Detects the shell type from the `$SHELL` environment variable.

```go
shellType := shell.DetectShell("/bin/zsh")
// Returns: ShellZsh
```

#### `GenerateBashInit(basePath string) string`
Generates the bash shell integration script.

```go
script := shell.GenerateBashInit("/home/user/tries")
// Returns: bash function + export statement
```

#### `GenerateZshInit(basePath string) string`
Generates the zsh shell integration script.

```go
script := shell.GenerateZshInit("/home/user/tries")
// Returns: zsh function + export statement
```

#### `GenerateFishInit(basePath string) string`
Generates the fish shell integration script.

```go
script := shell.GenerateFishInit("/home/user/tries")
// Returns: fish function + set statement
```

#### `GenerateInit(shellType ShellType, basePath string) (string, error)`
Generates the appropriate shell init script based on shell type.

```go
script, err := shell.GenerateInit(shell.ShellBash, "/home/user/tries")
if err != nil {
    return err
}
fmt.Print(script)
```

## Environment Variables

- **`TRY_PATH`** - Base directory for tries (set by the init script)
- **`SHELL`** - Current shell path (used for auto-detection)

## Testing

Run the test suite:
```bash
go test ./internal/shell/...
```

Run with verbose output:
```bash
go test ./internal/shell/... -v
```

## Examples

### Auto-detect Shell
```bash
$ eval "$(try init)"
$ try
# TUI appears, select a directory
$ pwd
/Users/username/src/tries/my-experiment
```

### Force Bash on Zsh
```bash
$ eval "$(try init --shell bash)"
```

### Custom Path
```bash
$ eval "$(try init ~/projects/experiments)"
$ echo $TRY_PATH
/Users/username/projects/experiments
```

## Integration with TUI

The TUI (implemented separately) will:
1. Accept the `exec` subcommand
2. Display the interactive interface on stderr
3. Output shell commands to stdout when the user makes a selection
4. Exit with code 0 on success

Example TUI output flow:
```
stderr: [Interactive TUI appears]
stdout: cd /path/to/selected/directory
exit code: 0
```

The shell function captures stdout and executes it, resulting in the shell changing to the selected directory.
