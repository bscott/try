package shell

import (
	"fmt"
	"strings"
)

// GenerateBashInit generates the bash shell function for try integration
func GenerateBashInit(basePath string) string {
	return fmt.Sprintf(`# try - Shell integration for bash
# Add this to your ~/.bashrc or ~/.bash_profile:
# eval "$(try init %s)"

try() {
    local result
    result=$(command try exec "$@" 2>/dev/tty)
    if [[ $? -eq 0 ]]; then
        eval "$result"
    fi
}

export TRY_PATH="%s"
`, basePath, basePath)
}

// GenerateZshInit generates the zsh shell function for try integration
func GenerateZshInit(basePath string) string {
	return fmt.Sprintf(`# try - Shell integration for zsh
# Add this to your ~/.zshrc:
# eval "$(try init %s)"

try() {
    local result
    result=$(command try exec "$@" 2>/dev/tty)
    if [[ $? -eq 0 ]]; then
        eval "$result"
    fi
}

export TRY_PATH="%s"
`, basePath, basePath)
}

// GenerateFishInit generates the fish shell function for try integration
func GenerateFishInit(basePath string) string {
	return fmt.Sprintf(`# try - Shell integration for fish
# Add this to your ~/.config/fish/config.fish:
# try init %s | source

function try
    set result (command try exec $argv 2>/dev/tty)
    if test $status -eq 0
        eval $result
    end
end

set -gx TRY_PATH "%s"
`, basePath, basePath)
}

// ShellType represents the type of shell
type ShellType int

const (
	ShellUnknown ShellType = iota
	ShellBash
	ShellZsh
	ShellFish
)

// DetectShell detects the shell type from the SHELL environment variable
func DetectShell(shellEnv string) ShellType {
	shellEnv = strings.ToLower(shellEnv)

	if strings.Contains(shellEnv, "bash") {
		return ShellBash
	}
	if strings.Contains(shellEnv, "zsh") {
		return ShellZsh
	}
	if strings.Contains(shellEnv, "fish") {
		return ShellFish
	}

	return ShellUnknown
}

// GenerateInit generates the appropriate shell init script based on shell type
func GenerateInit(shellType ShellType, basePath string) (string, error) {
	switch shellType {
	case ShellBash:
		return GenerateBashInit(basePath), nil
	case ShellZsh:
		return GenerateZshInit(basePath), nil
	case ShellFish:
		return GenerateFishInit(basePath), nil
	default:
		return "", fmt.Errorf("unsupported shell type")
	}
}
