# Fuzzy Matching Package

This package implements the fuzzy matching algorithm from the original Ruby `try` tool, providing smart searching and ranking of directory entries.

## Features

- **Case-insensitive matching**: Queries match regardless of case
- **Subsequence matching**: Characters don't need to be consecutive (e.g., `rds` matches `redis-server`)
- **Recency weighting**: Recently modified entries score higher
- **Date prefix bonus**: Entries with `YYYY-MM-DD-` prefix get a scoring boost
- **Word boundary detection**: Matches at word boundaries score higher
- **Proximity bonus**: Consecutive character matches score higher
- **Density scoring**: Compact matches score higher than spread-out ones
- **Length penalty**: Shorter names score higher for similar matches

## Usage

```go
import (
    "github.com/bscott/try/internal/fuzzy"
    "github.com/bscott/try/internal/tries"
)

// Create entries
entries := []tries.Entry{
    {Name: "redis-server", Path: "/tmp/redis-server", ModTime: time.Now()},
    {Name: "postgres", Path: "/tmp/postgres", ModTime: time.Now()},
}

// Perform fuzzy matching
matches := fuzzy.Match("rds", entries)

// Results are sorted by score (highest first)
for _, match := range matches {
    fmt.Printf("%s: %.2f\n", match.Entry.Name, match.Score)
}
```

## Scoring Algorithm

The algorithm combines multiple scoring components to produce a final score for each match:

### 1. Base Score (Recency)

```
base_score = 100.0 / (days_since_mtime + 1.0)
```

More recent entries get higher base scores.

### 2. Date Prefix Bonus

```
if entry.Name starts with YYYY-MM-DD-:
    base_score += 10.0
```

Entries with date prefixes are prioritized.

### 3. Character Match Scoring

For each matched character:
- **+1.0** for the character match itself
- **+1.0** if the character is at a word boundary (start of string or after non-alphanumeric)
- **Proximity bonus** for consecutive matches:
  ```
  gap = current_position - previous_position - 1
  bonus = 2.0 / sqrt(gap + 1)
  ```

### 4. Density Multiplier

```
density = query_length / span_width
```

Where `span_width` is the distance from the first matched character to the last. Compact matches get higher density scores.

### 5. Length Penalty

```
length_penalty = 10.0 / (entry_name_length + 10.0)
```

Shorter entry names are preferred over longer ones.

### Final Score

```
final_score = (base_score + match_score) * density * length_penalty
```

## Empty Query Behavior

When the query is empty, all entries are returned sorted by base score (recency) only.

## Examples

### Basic Fuzzy Matching

```go
// "rds" matches "redis-server"
matches := fuzzy.Match("rds", entries)
```

### Empty Query (List All)

```go
// Returns all entries sorted by recency
matches := fuzzy.Match("", entries)
```

### Word Boundary Matching

```go
// "cp" prefers "connection-pool" over "escape-pod"
// because 'c' and 'p' are at word boundaries
```

### Proximity Matching

```go
// "redis" scores higher on "redis-server" than "r-e-d-i-s-spread"
// because characters are closer together
```

## Testing

Run tests with:

```bash
go test ./internal/fuzzy/...
```

Run with coverage:

```bash
go test ./internal/fuzzy -cover
```

## Implementation Details

- **Pre-computed square roots**: For performance, square roots for gaps 0-15 are pre-computed
- **Greedy matching**: Characters are matched greedily from left to right
- **Character order**: Query characters must appear in order in the target string
- **Unicode support**: Uses Go's `unicode` package for proper character classification

## Port Notes

This is a faithful port of the Ruby implementation from [tobi/try](https://github.com/tobi/try), with the following adjustments:

1. Base score uses days instead of hours for better scaling over longer time periods
2. Date prefix bonus increased to +10.0 for better visibility in scoring
3. Go-idiomatic naming and error handling
4. Leverages Go's type system for clarity and safety
