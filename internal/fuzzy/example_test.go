package fuzzy_test

import (
	"fmt"
	"time"

	"github.com/bscott/try/internal/fuzzy"
	"github.com/bscott/try/internal/tries"
)

// Example demonstrates basic fuzzy matching usage.
func Example() {
	// Create some sample entries
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "redis-server",
			Path:          "/tmp/redis-server",
			ModTime:       now.Add(-1 * time.Hour),
			HasDatePrefix: false,
		},
		{
			Name:          "redis-cli",
			Path:          "/tmp/redis-cli",
			ModTime:       now.Add(-2 * time.Hour),
			HasDatePrefix: false,
		},
		{
			Name:          "postgres-db",
			Path:          "/tmp/postgres-db",
			ModTime:       now.Add(-3 * time.Hour),
			HasDatePrefix: false,
		},
	}

	// Perform fuzzy matching
	matches := fuzzy.Match("rds", entries)

	// Display results
	for i, match := range matches {
		fmt.Printf("%d. %s (score: %.2f)\n", i+1, match.Entry.Name, match.Score)
	}

	// Output shows matches sorted by score
	// Exact output will vary based on current time
}

// Example_emptyQuery demonstrates that empty queries return all entries sorted by recency.
func Example_emptyQuery() {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:    "old-project",
			Path:    "/tmp/old-project",
			ModTime: now.Add(-48 * time.Hour),
		},
		{
			Name:    "new-project",
			Path:    "/tmp/new-project",
			ModTime: now.Add(-1 * time.Hour),
		},
	}

	matches := fuzzy.Match("", entries)

	// Empty query returns all entries sorted by recency
	fmt.Printf("Most recent: %s\n", matches[0].Entry.Name)
	fmt.Printf("Older: %s\n", matches[1].Entry.Name)

	// Output:
	// Most recent: new-project
	// Older: old-project
}

// Example_datePrefix demonstrates the date prefix bonus.
func Example_datePrefix() {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "2024-01-25-experiment",
			Path:          "/tmp/2024-01-25-experiment",
			ModTime:       now.Add(-24 * time.Hour),
			HasDatePrefix: true,
			DatePrefix:    "2024-01-25",
		},
		{
			Name:          "experiment-backup",
			Path:          "/tmp/experiment-backup",
			ModTime:       now.Add(-24 * time.Hour),
			HasDatePrefix: false,
		},
	}

	matches := fuzzy.Match("exp", entries)

	// Date-prefixed entries get a scoring bonus
	for _, match := range matches {
		fmt.Printf("%s: %.2f\n", match.Entry.Name, match.Score)
	}

	// Date-prefixed entry should rank higher (actual scores vary)
}
