package main

import (
	"fmt"
	"os"

	"github.com/bscott/try/internal/config"
	"github.com/bscott/try/internal/shell"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Generate shell integration script",
	Long: `Generate shell integration script for try.

The init command generates a shell function that wraps the try binary,
allowing it to execute commands like 'cd' in your current shell session.

Usage:
  eval "$(try init)"                    # Use default path from TRY_PATH or ~/src/tries
  eval "$(try init ~/my/tries)"         # Use custom path
  try init ~/my/tries | source          # For fish shell

The shell function will:
1. Call the try binary with 'exec' subcommand
2. Capture the output (shell commands like 'cd /path')
3. Execute the commands in your current shell session

Add the eval line to your shell's RC file for permanent integration:
  Bash: ~/.bashrc or ~/.bash_profile
  Zsh:  ~/.zshrc
  Fish: ~/.config/fish/config.fish
`,
	RunE: runInit,
}

func init() {
	// Add init command to root
	rootCmd.AddCommand(initCmd)

	// Add shell flag for explicit shell type selection
	initCmd.Flags().StringP("shell", "s", "", "Shell type (bash, zsh, fish). Auto-detected if not specified.")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine the base path
	var basePath string
	var err error

	if len(args) > 0 {
		// Use provided path
		basePath = args[0]
	} else {
		// Use config default (checks TRY_PATH env var or uses ~/src/tries)
		basePath, err = config.GetBasePath()
		if err != nil {
			return fmt.Errorf("failed to determine base path: %w", err)
		}
	}

	// Get shell type from flag or auto-detect
	shellFlag, _ := cmd.Flags().GetString("shell")
	var shellType shell.ShellType

	if shellFlag != "" {
		// Use explicit shell type from flag
		shellType = parseShellFlag(shellFlag)
		if shellType == shell.ShellUnknown {
			return fmt.Errorf("unsupported shell type: %s (supported: bash, zsh, fish)", shellFlag)
		}
	} else {
		// Auto-detect from SHELL environment variable
		shellEnv := os.Getenv("SHELL")
		shellType = shell.DetectShell(shellEnv)

		// Handle unknown shell
		if shellType == shell.ShellUnknown {
			fmt.Fprintf(os.Stderr, "Warning: Could not detect shell type from SHELL=%s\n", shellEnv)
			fmt.Fprintf(os.Stderr, "Defaulting to bash. Use --shell flag to specify: bash, zsh, or fish\n\n")
			shellType = shell.ShellBash
		}
	}

	// Generate init script
	initScript, err := shell.GenerateInit(shellType, basePath)
	if err != nil {
		return fmt.Errorf("failed to generate init script: %w", err)
	}

	// Output to stdout (for eval)
	fmt.Print(initScript)

	return nil
}

func parseShellFlag(shellFlag string) shell.ShellType {
	switch shellFlag {
	case "bash":
		return shell.ShellBash
	case "zsh":
		return shell.ShellZsh
	case "fish":
		return shell.ShellFish
	default:
		return shell.ShellUnknown
	}
}
