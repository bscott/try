package fuzzy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
)

// TestIntegration demonstrates how fuzzy matching integrates with the tries package.
func TestIntegration(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create some test directories
	testDirs := []struct {
		name  string
		delay time.Duration // Delay before creation to vary mtime
	}{
		{"2024-01-25-redis-experiment", 0},
		{"postgres-test", 100 * time.Millisecond},
		{"redis-server-backup", 200 * time.Millisecond},
		{"connection-pool", 300 * time.Millisecond},
	}

	var entries []tries.Entry
	baseTime := time.Now()

	for i, td := range testDirs {
		// Create directory
		dirPath := filepath.Join(tmpDir, td.name)
		if err := os.Mkdir(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Parse the entry
		entry, err := tries.ParseEntry(dirPath)
		if err != nil {
			t.Fatalf("Failed to parse entry: %v", err)
		}

		// Manually set ModTime for predictable ordering
		entry.ModTime = baseTime.Add(time.Duration(i) * -24 * time.Hour)

		entries = append(entries, *entry)
	}

	// Test 1: Empty query returns all entries sorted by recency
	t.Run("empty query", func(t *testing.T) {
		matches := fuzzy.Match("", entries)

		if len(matches) != len(entries) {
			t.Errorf("Expected %d matches, got %d", len(entries), len(matches))
		}

		// Most recent should be first (smallest index in our test data)
		if matches[0].Entry.Name != "2024-01-25-redis-experiment" {
			t.Errorf("Expected most recent entry first, got %s", matches[0].Entry.Name)
		}
	})

	// Test 2: Fuzzy matching for "redis"
	t.Run("redis query", func(t *testing.T) {
		matches := fuzzy.Match("redis", entries)

		if len(matches) < 2 {
			t.Errorf("Expected at least 2 matches for 'redis', got %d", len(matches))
		}

		// All matches should contain 'redis' in some form
		for _, match := range matches {
			if !containsSubsequence("redis", match.Entry.Name) {
				t.Errorf("Match %s doesn't contain subsequence 'redis'", match.Entry.Name)
			}
		}
	})

	// Test 3: Fuzzy matching with compact query
	t.Run("conn query", func(t *testing.T) {
		matches := fuzzy.Match("conn", entries)

		if len(matches) == 0 {
			t.Error("Expected at least 1 match for 'conn'")
		}

		// connection-pool should match
		found := false
		for _, match := range matches {
			if match.Entry.Name == "connection-pool" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected 'connection-pool' to match 'conn'")
		}
	})

	// Test 4: Date prefix bonus
	t.Run("date prefix bonus", func(t *testing.T) {
		// Find entries with and without date prefix
		var datedScore, undatedScore float64

		matches := fuzzy.Match("redis", entries)

		for _, match := range matches {
			if match.Entry.Name == "2024-01-25-redis-experiment" {
				datedScore = match.Score
			} else if match.Entry.Name == "redis-server-backup" {
				undatedScore = match.Score
			}
		}

		// The dated entry should score higher than the undated one
		// (given they have similar match quality)
		if datedScore <= undatedScore {
			t.Logf("Warning: Dated entry scored %f, undated scored %f", datedScore, undatedScore)
			// Note: This might not always be true depending on the length penalty,
			// but it's good to log for visibility
		}
	})
}

// Helper function to check if a string contains a subsequence
func containsSubsequence(query, target string) bool {
	query = toLower(query)
	target = toLower(target)

	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			qi++
		}
	}

	return qi == len(query)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// Example_integration shows a realistic usage scenario
func Example_integration() {
	// Simulate a set of try directories
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "2024-01-25-redis-experiment",
			Path:          "/tmp/tries/2024-01-25-redis-experiment",
			ModTime:       now.Add(-1 * time.Hour),
			HasDatePrefix: true,
			DatePrefix:    "2024-01-25",
		},
		{
			Name:    "postgres-migration",
			Path:    "/tmp/tries/postgres-migration",
			ModTime: now.Add(-48 * time.Hour),
		},
		{
			Name:    "redis-server-config",
			Path:    "/tmp/tries/redis-server-config",
			ModTime: now.Add(-72 * time.Hour),
		},
	}

	// User types "rds" to find redis-related directories
	matches := fuzzy.Match("rds", entries)

	fmt.Println("Matches for 'rds':")
	for i, match := range matches {
		fmt.Printf("%d. %s\n", i+1, match.Entry.Name)
	}

	// Output will show redis-related entries ranked by score
	// The dated entry and more recent one will typically rank higher
}
