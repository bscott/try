package fuzzy

import (
	"testing"
	"time"

	"github.com/bscott/try/internal/tries"
)

func TestMatch_EmptyQuery(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "old-project",
			Path:          "/tmp/old-project",
			ModTime:       now.Add(-48 * time.Hour), // 2 days old
			HasDatePrefix: false,
		},
		{
			Name:          "recent-project",
			Path:          "/tmp/recent-project",
			ModTime:       now.Add(-1 * time.Hour), // 1 hour old
			HasDatePrefix: false,
		},
		{
			Name:          "medium-project",
			Path:          "/tmp/medium-project",
			ModTime:       now.Add(-24 * time.Hour), // 1 day old
			HasDatePrefix: false,
		},
	}

	matches := Match("", entries)

	// Should return all entries
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// Should be sorted by recency (most recent first)
	expectedOrder := []string{"recent-project", "medium-project", "old-project"}
	for i, expected := range expectedOrder {
		if matches[i].Entry.Name != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, matches[i].Entry.Name)
		}
	}

	// All scores should be positive
	for i, match := range matches {
		if match.Score <= 0 {
			t.Errorf("Match %d (%s) has non-positive score: %f", i, match.Entry.Name, match.Score)
		}
	}
}

func TestMatch_ExactMatch(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "redis-server",
			Path:          "/tmp/redis-server",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "redis-cli",
			Path:          "/tmp/redis-cli",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "postgres",
			Path:          "/tmp/postgres",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("redis-server", entries)

	// Exact match should score highest
	if len(matches) == 0 {
		t.Fatal("Expected at least one match")
	}

	if matches[0].Entry.Name != "redis-server" {
		t.Errorf("Expected 'redis-server' to rank first, got %s", matches[0].Entry.Name)
	}
}

func TestMatch_DatePrefixBonus(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "2024-01-25-redis-test",
			Path:          "/tmp/2024-01-25-redis-test",
			ModTime:       now.Add(-24 * time.Hour), // Same age, same length approximately
			HasDatePrefix: true,
			DatePrefix:    "2024-01-25",
		},
		{
			Name:          "redis-test-backup",
			Path:          "/tmp/redis-test-backup",
			ModTime:       now.Add(-24 * time.Hour), // Same age, similar length
			HasDatePrefix: false,
		},
	}

	matches := Match("redis", entries)

	if len(matches) < 2 {
		t.Fatal("Expected at least 2 matches")
	}

	// Entry with date prefix should score higher (has +10.0 bonus)
	// Even though names have similar length and match quality,
	// the date prefix bonus should tip the scales
	if matches[0].Entry.Name != "2024-01-25-redis-test" {
		t.Errorf("Expected dated entry to rank first, got %s (scores: dated=%f, undated=%f)",
			matches[0].Entry.Name,
			getScoreByName(matches, "2024-01-25-redis-test"),
			getScoreByName(matches, "redis-test-backup"))
	}
}

// Helper function for tests
func getScoreByName(matches []Result, name string) float64 {
	for _, m := range matches {
		if m.Entry.Name == name {
			return m.Score
		}
	}
	return 0.0
}

func TestMatch_WordBoundaryBonus(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "my-test-project", // 't' at word boundary after '-'
			Path:          "/tmp/my-test-project",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "contest-project", // 't' in middle of word
			Path:          "/tmp/contest-project",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("t", entries)

	if len(matches) < 2 {
		t.Fatal("Expected at least 2 matches")
	}

	// Entry with 't' at word boundary should score higher
	if matches[0].Entry.Name != "my-test-project" {
		t.Errorf("Expected 'my-test-project' to rank first, got %s", matches[0].Entry.Name)
	}
}

func TestMatch_ProximityBonus(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "redis-server", // 'rds' has gaps
			Path:          "/tmp/redis-server",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "rds-backup", // 'rds' is consecutive
			Path:          "/tmp/rds-backup",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("rds", entries)

	if len(matches) < 2 {
		t.Fatal("Expected at least 2 matches")
	}

	// Entry with consecutive characters should score higher
	if matches[0].Entry.Name != "rds-backup" {
		t.Errorf("Expected 'rds-backup' to rank first, got %s (score: %f vs %f)",
			matches[0].Entry.Name, matches[0].Score, matches[1].Score)
	}
}

func TestMatch_NoMatch(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "redis-server",
			Path:          "/tmp/redis-server",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "postgres",
			Path:          "/tmp/postgres",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("xyz", entries)

	// Should return no matches
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for non-matching query, got %d", len(matches))
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "Redis-Server",
			Path:          "/tmp/Redis-Server",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	testCases := []string{"redis", "REDIS", "ReDiS", "redis-server"}

	for _, query := range testCases {
		matches := Match(query, entries)
		if len(matches) == 0 {
			t.Errorf("Query '%s' should match 'Redis-Server'", query)
		}
	}
}

func TestMatch_CharacterOrder(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "connection-pool",
			Path:          "/tmp/connection-pool",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	// Should match: characters in order
	matches := Match("connpool", entries)
	if len(matches) == 0 {
		t.Error("Expected 'connpool' to match 'connection-pool'")
	}

	// Should not match: characters out of order
	matches = Match("poolconn", entries)
	if len(matches) != 0 {
		t.Error("Expected 'poolconn' NOT to match 'connection-pool'")
	}
}

func TestMatch_LengthPenalty(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "test",
			Path:          "/tmp/test",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "test-with-very-long-name-here",
			Path:          "/tmp/test-with-very-long-name-here",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("test", entries)

	if len(matches) < 2 {
		t.Fatal("Expected at least 2 matches")
	}

	// Shorter name should score higher
	if matches[0].Entry.Name != "test" {
		t.Errorf("Expected shorter name 'test' to rank first, got %s", matches[0].Entry.Name)
	}
}

func TestMatch_DensityMultiplier(t *testing.T) {
	now := time.Now()
	entries := []tries.Entry{
		{
			Name:          "abc-test", // 'abc' is compact (span=3)
			Path:          "/tmp/abc-test",
			ModTime:       now,
			HasDatePrefix: false,
		},
		{
			Name:          "a-big-container-test", // 'abc' is spread out (span=14)
			Path:          "/tmp/a-big-container-test",
			ModTime:       now,
			HasDatePrefix: false,
		},
	}

	matches := Match("abc", entries)

	if len(matches) < 2 {
		t.Fatal("Expected at least 2 matches")
	}

	// Compact match should score higher
	if matches[0].Entry.Name != "abc-test" {
		t.Errorf("Expected compact match 'abc-test' to rank first, got %s (score: %f vs %f)",
			matches[0].Entry.Name, matches[0].Score, matches[1].Score)
	}
}
