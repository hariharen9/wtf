package search

import (
	"strings"
	"unicode"
)

type MatchResult struct {
	Path           string
	Score          int
	MatchedIndices []int // Indices in the Path string that matched the query (for terminal highlighting)
}

// SmartCaseCheck checks if the query contains any uppercase letters.
// If it does, we perform case-sensitive search. Otherwise, case-insensitive.
func IsCaseSensitive(query string) bool {
	for _, r := range query {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// SubstringMatch performs an ultra-fast substring check and scores the result.
func SubstringMatch(path, query string, caseSensitive bool) (MatchResult, bool) {
	origPath := path
	origQuery := query

	if !caseSensitive {
		path = strings.ToLower(path)
		query = strings.ToLower(query)
	}

	idx := strings.Index(path, query)
	if idx == -1 {
		return MatchResult{}, false
	}

	// Calculate score
	score := 100

	// Bonus: match is at the very end (filename match)
	filename := getFilename(origPath)
	filenameLower := strings.ToLower(filename)
	
	fileIdx := strings.Index(filenameLower, query)
	if fileIdx != -1 {
		score += 200
		
		// Perfect match for filename
		if len(filenameLower) == len(query) {
			score += 500
		}
		// Prefix match on filename
		if fileIdx == 0 {
			score += 100
		}
	}

	// Bonus: matches right after a separator
	if idx > 0 && (origPath[idx-1] == '/' || origPath[idx-1] == '\\') {
		score += 50
	}

	// Generate matched indices
	matchedIndices := make([]int, len(origQuery))
	for i := 0; i < len(origQuery); i++ {
		matchedIndices[i] = idx + i
	}

	return MatchResult{
		Path:           origPath,
		Score:          score,
		MatchedIndices: matchedIndices,
	}, true
}

// FuzzyMatch performs a highly optimized fuzzy search, scoring, and returns matched indices.
// Characters in the query must appear in order in the path, but can be non-consecutive.
func FuzzyMatch(path, query string, caseSensitive bool) (MatchResult, bool) {
	if len(query) == 0 {
		return MatchResult{Path: path, Score: 0}, true
	}

	origPath := path

	if !caseSensitive {
		path = strings.ToLower(path)
		query = strings.ToLower(query)
	}

	pathLen := len(path)
	queryLen := len(query)

	if queryLen > pathLen {
		return MatchResult{}, false
	}

	// Quick check if all characters exist in order
	pIdx := 0
	qIdx := 0
	matchedIndices := make([]int, 0, queryLen)

	for pIdx < pathLen && qIdx < queryLen {
		if path[pIdx] == query[qIdx] {
			matchedIndices = append(matchedIndices, pIdx)
			qIdx++
		}
		pIdx++
	}

	if qIdx < queryLen {
		return MatchResult{}, false
	}

	// Calculate dynamic score
	score := 0
	consecutiveCount := 0
	
	// Pre-find path boundary information
	lastSepIdx := strings.LastIndexAny(origPath, "/\\")

	for i, pPos := range matchedIndices {
		// 1. Consecutive character match bonus
		if i > 0 && pPos == matchedIndices[i-1]+1 {
			consecutiveCount++
			score += 15 + (consecutiveCount * 5)
		} else {
			consecutiveCount = 0
		}

		// 2. Separator boundary bonus (character directly after / or \)
		if pPos > 0 && (origPath[pPos-1] == '/' || origPath[pPos-1] == '\\') {
			score += 40
		}

		// 3. Filename matching bonus
		if pPos > lastSepIdx {
			score += 20 // Characters in filename are more important
			
			// Start of filename bonus
			if pPos == lastSepIdx+1 {
				score += 50
			}
		}

		// 4. Penalty for gaps
		if i > 0 {
			gap := pPos - matchedIndices[i-1] - 1
			if gap > 0 {
				score -= gap * 2
			}
		}
	}

	// Exact substring match check inside fuzzy
	if subIdx := strings.Index(path, query); subIdx != -1 {
		score += 150
		if subIdx > lastSepIdx {
			score += 200 // Substring match is completely within the filename
		}
	}

	return MatchResult{
		Path:           origPath,
		Score:          score,
		MatchedIndices: matchedIndices,
	}, true
}

func getFilename(path string) string {
	idx := strings.LastIndexAny(path, "/\\")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}
