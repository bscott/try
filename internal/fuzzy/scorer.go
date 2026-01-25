package fuzzy

import (
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/bscott/try/internal/tries"
)

// Pre-computed square root table for proximity bonuses (matching Ruby implementation)
var sqrtTable [16]float64

func init() {
	for i := 0; i < 16; i++ {
		sqrtTable[i] = 2.0 / math.Sqrt(float64(i+1))
	}
}

// scoreEntry calculates the fuzzy match score for an entry against a query.
// The score combines multiple components:
// 1. Base score from recency: 100.0 / (days_since_mtime + 1.0)
// 2. Date prefix bonus: +10.0 if name starts with YYYY-MM-DD-
// 3. Character match bonuses
// 4. Word boundary bonuses
// 5. Proximity bonuses for consecutive matches
// 6. Density multiplier: query_len / span_width
// 7. Length penalty: 10.0 / (name_len + 10.0)
func scoreEntry(query string, entry tries.Entry) float64 {
	// Base score from recency
	daysSinceMtime := time.Since(entry.ModTime).Hours() / 24.0
	baseScore := 100.0 / (daysSinceMtime + 1.0)

	// Date prefix bonus
	if entry.HasDatePrefix {
		baseScore += 10.0
	}

	// Empty query: return base score only
	if query == "" {
		return baseScore
	}

	// Find character match positions
	positions := findMatches(query, entry.Name)
	if len(positions) == 0 {
		return 0.0 // No match
	}

	// Calculate character match and bonus scores
	matchScore := calculateMatchScore(query, entry.Name, positions)

	// Combine scores
	score := baseScore + matchScore

	// Apply density multiplier (query length / span width)
	queryLen := float64(len(query))
	spanWidth := float64(positions[len(positions)-1] - positions[0] + 1)
	densityMultiplier := queryLen / spanWidth

	// Apply length penalty
	nameLen := float64(len(entry.Name))
	lengthPenalty := 10.0 / (nameLen + 10.0)

	return score * densityMultiplier * lengthPenalty
}

// findMatches finds the positions of query characters in the name.
// Characters must appear in order but don't need to be consecutive.
// Matching is case-insensitive.
func findMatches(query, name string) []int {
	if query == "" {
		return []int{}
	}

	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(name)

	var positions []int
	nameIdx := 0

	for _, qChar := range queryLower {
		found := false
		for nameIdx < len(nameLower) {
			if rune(nameLower[nameIdx]) == qChar {
				positions = append(positions, nameIdx)
				nameIdx++
				found = true
				break
			}
			nameIdx++
		}
		if !found {
			return []int{} // Character not found in sequence
		}
	}

	return positions
}

// calculateMatchScore calculates bonuses for character matches.
// Returns the sum of:
// - +1.0 per matched character
// - +1.0 if match is at word boundary (start or after non-alphanumeric)
// - Proximity bonus for consecutive matches: 2.0 / sqrt(gap + 1)
func calculateMatchScore(query, name string, positions []int) float64 {
	score := 0.0
	nameLower := strings.ToLower(name)

	for i, pos := range positions {
		// +1.0 per character match
		score += 1.0

		// Word boundary bonus: +1.0 if at start or after non-alphanumeric
		if pos == 0 || !isAlphanumeric(rune(nameLower[pos-1])) {
			score += 1.0
		}

		// Proximity bonus for consecutive matches
		if i > 0 {
			gap := pos - positions[i-1] - 1
			if gap < 16 {
				score += sqrtTable[gap]
			} else {
				score += 2.0 / math.Sqrt(float64(gap+1))
			}
		}
	}

	return score
}

// isAlphanumeric returns true if the rune is a letter or digit.
func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
