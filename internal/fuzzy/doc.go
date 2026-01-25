// Package fuzzy implements fuzzy matching and ranking for directory entries.
//
// This package provides smart search functionality that matches query strings
// against directory names using a sophisticated scoring algorithm. The algorithm
// considers multiple factors including recency, match quality, word boundaries,
// character proximity, and entry name length.
//
// # Basic Usage
//
//	entries := []tries.Entry{
//	    {Name: "redis-server", Path: "/tmp/redis-server", ModTime: time.Now()},
//	    {Name: "postgres", Path: "/tmp/postgres", ModTime: time.Now()},
//	}
//
//	matches := fuzzy.Match("rds", entries)
//	for _, match := range matches {
//	    fmt.Printf("%s (score: %.2f)\n", match.Entry.Name, match.Score)
//	}
//
// # Scoring Algorithm
//
// The scoring algorithm combines multiple components:
//
// 1. Base Score (Recency): More recent entries score higher
//   - Formula: 100.0 / (days_since_mtime + 1.0)
//   - Date prefix bonus: +10.0 if name starts with YYYY-MM-DD-
//
// 2. Character Match Score:
//   - +1.0 per matched character
//   - +1.0 bonus for word boundary matches (start or after non-alphanumeric)
//   - Proximity bonus: 2.0 / sqrt(gap + 1) for consecutive matches
//
// 3. Density Multiplier: query_length / span_width
//   - Rewards compact matches over spread-out ones
//
// 4. Length Penalty: 10.0 / (entry_name_length + 10.0)
//   - Prefers shorter names when match quality is similar
//
// Final score = (base_score + match_score) × density × length_penalty
//
// # Empty Query Behavior
//
// When the query string is empty, all entries are returned sorted by base score
// (recency) only. This provides a simple way to list all entries sorted by
// most recently modified.
//
// # Match Quality Examples
//
//   - "rds" matches "redis-server" (subsequence matching)
//   - "connpool" matches "connection-pool" (character-by-character)
//   - "cp" prefers "connection-pool" over "escape-pod" (word boundaries)
//   - Matching is case-insensitive
//   - Characters must appear in order but don't need to be consecutive
package fuzzy
