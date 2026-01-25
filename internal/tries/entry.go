package tries

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Entry represents a single try directory
type Entry struct {
	Name          string    // Full directory name
	Path          string    // Full path to directory
	ModTime       time.Time // Last modification time
	HasDatePrefix bool      // Whether the name has a date prefix
	DatePrefix    string    // Date prefix if present (e.g., "2024-01-25")
}

// Date prefix pattern: YYYY-MM-DD- at the start of the name
var datePrefixRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-`)

// ParseEntry creates an Entry from a directory path.
// It extracts the directory name, modification time, and any date prefix.
func ParseEntry(path string) (*Entry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", path)
	}

	name := filepath.Base(path)
	entry := &Entry{
		Name:    name,
		Path:    path,
		ModTime: info.ModTime(),
	}

	// Check for date prefix
	matches := datePrefixRegex.FindStringSubmatch(name)
	if len(matches) > 1 {
		entry.HasDatePrefix = true
		entry.DatePrefix = matches[1]
	}

	return entry, nil
}

// DisplayName returns the name without the date prefix if present
func (e *Entry) DisplayName() string {
	if e.HasDatePrefix {
		return datePrefixRegex.ReplaceAllString(e.Name, "")
	}
	return e.Name
}
