package main

import (
	"fmt"
	"os"

	"github.com/bscott/try/internal/config"
	"github.com/bscott/try/internal/tries"
	"github.com/bscott/try/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [query]",
	Short: "Interactive TUI for selecting a try directory",
	Long: `Launch an interactive TUI for browsing and selecting try directories.

The exec command starts a terminal user interface that allows you to:
- Search for try directories with fuzzy matching
- Navigate with arrow keys or Ctrl+P/Ctrl+N
- Create new try directories (Ctrl+T)
- Rename existing directories (Ctrl+R)
- Delete directories (Ctrl+D)
- Select a directory to change into (Enter)

When you select a directory, the command outputs "cd /path" to stdout,
which the shell wrapper function evaluates to change your current directory.

Usage:
  try [query]              # Launch TUI with optional initial search query

Key bindings:
  ↑/↓ or Ctrl+P/N         Navigate list
  Enter                    Select directory
  Ctrl+D                   Delete mode
  Ctrl+R                   Rename directory
  Ctrl+T                   Create new directory
  Ctrl+C                   Quit
  Esc                      Cancel current action

Search input supports Emacs-style editing:
  Ctrl+A/E                 Start/end of line
  Ctrl+B/F                 Back/forward character
  Ctrl+W                   Delete word back
  Ctrl+K                   Kill to end of line
`,
	RunE: runExec,
}

func init() {
	// Add exec command to root
	rootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	// Get base path from config
	basePath, err := config.GetBasePath()
	if err != nil {
		return fmt.Errorf("failed to get base path: %w", err)
	}

	// Create manager
	manager := tries.NewManager(basePath)

	// Create TUI model
	model := tui.New(manager)

	// Set initial query if provided
	if len(args) > 0 {
		// Note: Initial query support can be added to the model later
		// For now, just launch the TUI
	}

	// Open /dev/tty for input and output
	// This is necessary when running in a shell command substitution $()
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer tty.Close()

	// Create Bubble Tea program with explicit TTY
	// View() returns empty on exit, so only our manual output goes to stdout
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if we have a command to output
	if m, ok := finalModel.(tui.Model); ok {
		output := m.OutputCommand()
		if output != "" {
			// Output shell command to stdout for wrapper to eval
			fmt.Fprint(os.Stdout, output)
		}
	}

	return nil
}
