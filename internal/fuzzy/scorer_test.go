package fuzzy

import (
	"testing"
	"time"

	"github.com/bscott/try/internal/tries"
)

func TestFindMatches_EmptyQuery(t *testing.T) {
	positions := findMatches("", "test-name")
	if len(positions) != 0 {
		t.Errorf("Expected no positions for empty query, got %d", len(positions))
	}
}

func TestFindMatches_AllCharsMatch(t *testing.T) {
	positions := findMatches("test", "test-case")
	expected := []int{0, 1, 2, 3}

	if len(positions) != len(expected) {
		t.Fatalf("Expected %d positions, got %d", len(expected), len(positions))
	}

	for i, pos := range positions {
		if pos != expected[i] {
			t.Errorf("Position %d: expected %d, got %d", i, expected[i], pos)
		}
	}
}

func TestFindMatches_ScatteredMatch(t *testing.T) {
	positions := findMatches("rds", "redis-server")
	// redis-server: r=0, e=1, d=2, i=3, s=4
	expected := []int{0, 2, 4} // r, d, s

	if len(positions) != len(expected) {
		t.Fatalf("Expected %d positions, got %d", len(expected), len(positions))
	}

	for i, pos := range positions {
		if pos != expected[i] {
			t.Errorf("Position %d: expected %d, got %d", i, expected[i], pos)
		}
	}
}

func TestFindMatches_NoMatch(t *testing.T) {
	positions := findMatches("xyz", "test-case")
	if len(positions) != 0 {
		t.Errorf("Expected no positions for non-matching query, got %d", len(positions))
	}
}

func TestFindMatches_OutOfOrder(t *testing.T) {
	// Characters must appear in order - this tests characters that can't be found in sequence
	positions := findMatches("zyx", "test-case")
	if len(positions) != 0 {
		t.Errorf("Expected no match for impossible sequence, got %d positions", len(positions))
	}

	// Test another case: reversed characters that exist but wrong order
	positions = findMatches("tset", "test")
	// t=0, s=2, e=1, t=3
	// We'd find t=0, s=2, but then can't find 'e' after position 2 in "test"
	// Actually wait - in "test" after finding s at 2, there IS a t at 3, so this would match!
	// Let me use a better example
	positions = findMatches("ba", "abc")
	// Looking for 'b' then 'a', but in "abc" we have a=0, b=1, c=2
	// So we'd find b=1, then look for 'a' starting from position 2, which doesn't exist
	if len(positions) != 0 {
		t.Errorf("Expected no match for reversed sequence 'ba' in 'abc', got %d positions", len(positions))
	}
}

func TestFindMatches_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		query    string
		name     string
		expected int
	}{
		{"TEST", "test-case", 4},
		{"TeSt", "test-case", 4},
		{"test", "TEST-CASE", 4},
		{"TeSt", "TeSt-CaSe", 4},
	}

	for _, tc := range testCases {
		positions := findMatches(tc.query, tc.name)
		if len(positions) != tc.expected {
			t.Errorf("Query '%s' in '%s': expected %d positions, got %d",
				tc.query, tc.name, tc.expected, len(positions))
		}
	}
}

func TestCalculateMatchScore_SingleChar(t *testing.T) {
	// Single character at start
	score := calculateMatchScore("t", "test", []int{0})

	// Should have:
	// +1.0 for character match
	// +1.0 for word boundary (start of string)
	// No proximity bonus (only one character)
	expected := 2.0

	if score != expected {
		t.Errorf("Expected score %f, got %f", expected, score)
	}
}

func TestCalculateMatchScore_ConsecutiveChars(t *testing.T) {
	// Consecutive characters
	score := calculateMatchScore("tes", "test", []int{0, 1, 2})

	// Should have:
	// +1.0 for each char (3)
	// +1.0 for word boundary at position 0
	// +2.0 for proximity between 0-1 (gap=0, sqrtTable[0])
	// +2.0 for proximity between 1-2 (gap=0, sqrtTable[0])
	// Total: 3 + 1 + 2 + 2 = 8.0

	if score < 7.0 || score > 9.0 {
		t.Errorf("Expected score around 8.0, got %f", score)
	}
}

func TestCalculateMatchScore_WordBoundary(t *testing.T) {
	// Character after hyphen (word boundary)
	score1 := calculateMatchScore("s", "test-server", []int{5}) // 's' after hyphen

	// Character in middle of word (no boundary)
	score2 := calculateMatchScore("s", "testing", []int{2}) // 's' in middle

	// Word boundary should score higher
	if score1 <= score2 {
		t.Errorf("Word boundary score (%f) should be higher than non-boundary (%f)",
			score1, score2)
	}
}

func TestScoreEntry_BaseScore(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		entry       tries.Entry
		expectHigher bool
	}{
		{
			name: "recent entry",
			entry: tries.Entry{
				Name:    "test",
				ModTime: now.Add(-1 * time.Hour),
			},
			expectHigher: true,
		},
		{
			name: "old entry",
			entry: tries.Entry{
				Name:    "test",
				ModTime: now.Add(-100 * 24 * time.Hour),
			},
			expectHigher: false,
		},
	}

	recentScore := scoreEntry("", tests[0].entry)
	oldScore := scoreEntry("", tests[1].entry)

	if recentScore <= oldScore {
		t.Errorf("Recent entry should score higher: recent=%f, old=%f",
			recentScore, oldScore)
	}
}

func TestScoreEntry_DatePrefixBonus(t *testing.T) {
	now := time.Now()

	withPrefix := tries.Entry{
		Name:          "2024-01-25-test",
		ModTime:       now,
		HasDatePrefix: true,
	}

	withoutPrefix := tries.Entry{
		Name:          "test",
		ModTime:       now,
		HasDatePrefix: false,
	}

	// Empty query to test base score only
	scoreWith := scoreEntry("", withPrefix)
	scoreWithout := scoreEntry("", withoutPrefix)

	// Date prefix should add +10.0 to base score
	diff := scoreWith - scoreWithout

	if diff < 9.5 || diff > 10.5 {
		t.Errorf("Date prefix bonus should be ~10.0, got diff: %f", diff)
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'5', true},
		{'-', false},
		{'_', false},
		{' ', false},
		{'.', false},
	}

	for _, tc := range tests {
		result := isAlphanumeric(tc.char)
		if result != tc.expected {
			t.Errorf("isAlphanumeric('%c'): expected %v, got %v",
				tc.char, tc.expected, result)
		}
	}
}
