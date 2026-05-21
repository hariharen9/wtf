package indexer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"wtf/config"
	"wtf/search"
)

var ErrNoIndex = errors.New("search index does not exist. Please run 'wtf update' first")

// GetIndexPath returns the path to the search index database file.
func GetIndexPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "wtf", "index.db"), nil
}

// UpdateIndex performs a parallel directory scan and saves all paths to the index file.
func UpdateIndex(cfg *config.Config, progressChan chan<- int) (int, error) {
	indexPath, err := GetIndexPath()
	if err != nil {
		return 0, err
	}

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return 0, err
	}

	// Write to a temporary file first for atomic replacement
	tempFile, err := os.CreateTemp(filepath.Dir(indexPath), "wtf-index-*.tmp")
	if err != nil {
		return 0, err
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)
	defer tempFile.Close()

	// High buffer size for rapid disk write
	writer := bufio.NewWriterSize(tempFile, 2*1024*1024) // 2MB buffer
	
	outChan := make(chan WalkItem, 50000)
	var walkErr error

	// Run parallel walk in background
	go func() {
		walkErr = ParallelWalk(cfg, outChan)
		close(outChan)
	}()

	separator := string(filepath.Separator)
	var count int

	for item := range outChan {
		path := item.Path
		if item.IsDir && !strings.HasSuffix(path, separator) {
			path += separator
		}
		
		_, err := writer.WriteString(path + "\n")
		if err != nil {
			return 0, err
		}
		count++

		// Send progress updates if channel is provided
		if progressChan != nil && count%10000 == 0 {
			select {
			case progressChan <- count:
			default:
			}
		}
	}

	if walkErr != nil {
		return 0, fmt.Errorf("walk error: %w", walkErr)
	}

	if err := writer.Flush(); err != nil {
		return 0, err
	}
	tempFile.Close()

	// Atomic rename
	if err := os.Rename(tempName, indexPath); err != nil {
		return 0, fmt.Errorf("failed to finalize index: %w", err)
	}

	return count, nil
}

// Search queries the index file using multiple CPUs in parallel.
func Search(query string, limit int, fuzzy bool) ([]search.MatchResult, error) {
	indexPath, err := GetIndexPath()
	if err != nil {
		return nil, err
	}

	// Read index file fully into RAM
	content, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoIndex
		}
		return nil, err
	}

	if len(content) == 0 {
		return nil, nil
	}

	// Setup parallel scanner chunks
	numCPU := runtime.NumCPU()
	fileLen := len(content)
	if fileLen < 500000 {
		numCPU = 1 // Don't multithread for small files
	}

	chunkSize := fileLen / numCPU
	type Chunk struct {
		start int
		end   int
	}

	chunks := make([]Chunk, numCPU)
	curr := 0
	for i := 0; i < numCPU; i++ {
		start := curr
		end := curr + chunkSize
		if end > fileLen || i == numCPU-1 {
			end = fileLen
		} else {
			// Align end to nearest newline
			for end < fileLen && content[end] != '\n' {
				end++
			}
			if end < fileLen {
				end++ // include newline
			}
		}
		chunks[i] = Chunk{start: start, end: end}
		curr = end
	}

	caseSensitive := search.IsCaseSensitive(query)
	var wg sync.WaitGroup
	resultsChan := make(chan []search.MatchResult, numCPU)

	// Spawn parallel workers
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(c Chunk) {
			defer wg.Done()
			var localMatches []search.MatchResult
			
			buf := content[c.start:c.end]
			bufLen := len(buf)
			lineStart := 0

			for i := 0; i < bufLen; i++ {
				if buf[i] == '\n' {
					lineEnd := i
					// Handle CRLF on Windows
					if lineEnd > lineStart && buf[lineEnd-1] == '\r' {
						lineEnd--
					}

					lineStr := string(buf[lineStart:lineEnd])
					
					var match search.MatchResult
					var matched bool

					if fuzzy {
						match, matched = search.FuzzyMatch(lineStr, query, caseSensitive)
					} else {
						match, matched = search.SubstringMatch(lineStr, query, caseSensitive)
					}

					if matched {
						localMatches = append(localMatches, match)
					}

					lineStart = i + 1
				}
			}

			// Catch last line if not ending in newline (rare but possible)
			if lineStart < bufLen {
				lineStr := string(buf[lineStart:])
				if len(lineStr) > 0 && lineStr[len(lineStr)-1] == '\r' {
					lineStr = lineStr[:len(lineStr)-1]
				}

				var match search.MatchResult
				var matched bool
				if fuzzy {
					match, matched = search.FuzzyMatch(lineStr, query, caseSensitive)
				} else {
					match, matched = search.SubstringMatch(lineStr, query, caseSensitive)
				}
				if matched {
					localMatches = append(localMatches, match)
				}
			}

			resultsChan <- localMatches
		}(chunks[i])
	}

	// Close channel once workers finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Merge all results
	var allResults []search.MatchResult
	for localMatches := range resultsChan {
		allResults = append(allResults, localMatches...)
	}

	// Sort results by score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].Score == allResults[j].Score {
			// Tie-breaker: sort alphabetically or by shorter path
			if len(allResults[i].Path) == len(allResults[j].Path) {
				return allResults[i].Path < allResults[j].Path
			}
			return len(allResults[i].Path) < len(allResults[j].Path)
		}
		return allResults[i].Score > allResults[j].Score
	})

	// Cap results to limit
	if limit > 0 && len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}
