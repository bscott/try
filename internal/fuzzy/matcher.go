package fuzzy

import (
	"sort"

	"github.com/bscott/try/internal/tries"
)

// Result represents a fuzzy match result with its score.
type Result struct {
	Entry tries.Entry
	Score float64
}

// Match performs fuzzy matching on a list of entries.
// For empty queries, returns all entries sorted by base score (recency).
// For non-empty queries, returns matches sorted by score in descending order.
func Match(query string, entries []tries.Entry) []Result {
	matches := make([]Result, 0, len(entries))

	// Score all entries
	for _, entry := range entries {
		score := scoreEntry(query, entry)

		// For empty query, include all entries
		// For non-empty query, only include entries with matches
		if query == "" || score > 0 {
			matches = append(matches, Result{
				Entry: entry,
				Score: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}
