package main

import (
	"fmt"
	"os"

	"github.com/bscott/try/internal/config"
	"github.com/bscott/try/internal/tries"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "try",
	Short: "Manage temporary try directories",
	Long: `try is a tool for managing temporary project directories.
It helps you create, list, and manage ephemeral workspaces for experimentation.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Get base path from config
	basePath, err := config.GetBasePath()
	if err != nil {
		return fmt.Errorf("failed to get base path: %w", err)
	}

	// Create manager
	manager := tries.NewManager(basePath)

	// List entries
	entries, err := manager.List()
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	// Print entries
	if len(entries) == 0 {
		fmt.Println("No try directories found.")
		fmt.Printf("Base path: %s\n", basePath)
		return nil
	}

	fmt.Printf("Try directories in %s:\n\n", basePath)
	for _, entry := range entries {
		displayName := entry.DisplayName()
		if entry.HasDatePrefix {
			fmt.Printf("  [%s] %s\n", entry.DatePrefix, displayName)
		} else {
			fmt.Printf("  %s\n", entry.Name)
		}
		fmt.Printf("    Path: %s\n", entry.Path)
		fmt.Printf("    Modified: %s\n\n", entry.ModTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
