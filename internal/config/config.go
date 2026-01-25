package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultTriesDir = "~/src/tries"
	envVarName      = "TRY_PATH"
)

// GetBasePath returns the base directory for tries.
// It checks the TRY_PATH environment variable first, falling back to ~/src/tries.
// The path is expanded and the directory is created if it doesn't exist.
func GetBasePath() (string, error) {
	path := os.Getenv(envVarName)
	if path == "" {
		path = defaultTriesDir
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Clean the path
	path = filepath.Clean(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create tries directory: %w", err)
	}

	return path, nil
}
