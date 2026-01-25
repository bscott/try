package tries

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles operations on try directories
type Manager struct {
	basePath string
}

// NewManager creates a new Manager with the given base path
func NewManager(basePath string) *Manager {
	return &Manager{
		basePath: basePath,
	}
}

// List returns all try entries sorted by modification time (newest first)
func (m *Manager) List() ([]Entry, error) {
	return ScanEntries(m.basePath)
}

// Create creates a new try directory with date prefix (YYYY-MM-DD-name)
func (m *Manager) Create(name string) (string, error) {
	// Generate date prefix
	datePrefix := time.Now().Format("2006-01-02")
	fullName := fmt.Sprintf("%s-%s", datePrefix, name)
	dirPath := filepath.Join(m.basePath, fullName)

	// Create the directory (os.Mkdir fails atomically if it exists)
	if err := os.Mkdir(dirPath, 0755); err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("directory already exists: %s", fullName)
		}
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return dirPath, nil
}

// Delete deletes try directories
func (m *Manager) Delete(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}

	// Delete each entry
	for _, entry := range entries {
		if err := os.RemoveAll(entry.Path); err != nil {
			return fmt.Errorf("failed to delete %s: %w", entry.Name, err)
		}
	}

	return nil
}

// Rename renames a try directory
func (m *Manager) Rename(entry Entry, newName string) error {
	// Build new path - preserve date prefix if present
	var newFullName string
	if entry.HasDatePrefix {
		newFullName = fmt.Sprintf("%s-%s", entry.DatePrefix, newName)
	} else {
		newFullName = newName
	}

	newPath := filepath.Join(m.basePath, newFullName)

	// Rename the directory (os.Rename on Unix will fail if target exists)
	if err := os.Rename(entry.Path, newPath); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("directory already exists: %s", newFullName)
		}
		return fmt.Errorf("failed to rename directory: %w", err)
	}

	return nil
}
