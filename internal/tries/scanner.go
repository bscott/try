package tries

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ScanEntries scans the base directory for try directories and returns them sorted by modification time (newest first).
func ScanEntries(basePath string) ([]Entry, error) {
	// Read directory contents
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	// Parse each subdirectory into an Entry
	var parsed []Entry
	for _, entry := range entries {
		if !entry.IsDir() {
			// Skip non-directories
			continue
		}

		fullPath := filepath.Join(basePath, entry.Name())
		parsedEntry, err := ParseEntry(fullPath)
		if err != nil {
			// Log warning but continue processing other entries
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", fullPath, err)
			continue
		}

		parsed = append(parsed, *parsedEntry)
	}

	// Sort by modification time, newest first
	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].ModTime.After(parsed[j].ModTime)
	})

	return parsed, nil
}
