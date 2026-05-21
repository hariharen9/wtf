package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wtf/config"
	"wtf/indexer"
	"wtf/tui"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
)

const Version = "1.0.0"

func main() {
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "WTF - Where's The File? (v%s)\n", Version)
		fmt.Fprintf(os.Stderr, "A blazing-fast, cross-platform CLI file searcher & indexer.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  wtf [query]           Launch interactive fuzzy-finder TUI (with optional query)\n")
		fmt.Fprintf(os.Stderr, "  wtf update            Rebuild/update the search index\n")
		fmt.Fprintf(os.Stderr, "  wtf search [query]    Search index directly inside the terminal and print paths\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  wtf                   Launch TUI\n")
		fmt.Fprintf(os.Stderr, "  wtf main.go           Launch TUI searching for 'main.go'\n")
		fmt.Fprintf(os.Stderr, "  wtf search app.js     Print all matching 'app.js' paths directly\n")
		fmt.Fprintf(os.Stderr, "  wtf -o main.go        Instantly open the first matching 'main.go'\n")
		fmt.Fprintf(os.Stderr, "  wtf -c config         Instantly copy first match for 'config' to clipboard\n")
	}

	// Define flags
	fuzzyFlag := flag.Bool("f", false, "Use fuzzy matching instead of simple substring matching")
	limitFlag := flag.Int("n", 20, "Number of results to display in direct CLI search")
	openFlag := flag.Bool("o", false, "Immediately open the first match")
	copyFlag := flag.Bool("c", false, "Immediately copy the first matching path to clipboard")
	versionFlag := flag.Bool("v", false, "Print version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("wtf version %s\n", Version)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()

	// Route subcommands
	if len(args) > 0 {
		subcommand := args[0]
		switch subcommand {
		case "update":
			runDirectUpdate(cfg)
			return
		case "search":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: search command requires a query. Example: wtf search main.go")
				os.Exit(1)
			}
			runDirectSearch(cfg, strings.Join(args[1:], " "), *limitFlag, *fuzzyFlag)
			return
		}
	}

	// If -o (open) or -c (copy) flags are set, run search immediately and perform action
	if *openFlag || *copyFlag {
		query := strings.Join(args, " ")
		if query == "" {
			fmt.Fprintln(os.Stderr, "Error: immediate actions require a search query")
			os.Exit(1)
		}
		runDirectAction(cfg, query, *openFlag, *copyFlag, *fuzzyFlag)
		return
	}

	// Default behavior: Launch interactive TUI
	// We pass any trailing arguments as the initial search query inside the TUI
	initialQuery := strings.Join(args, " ")
	
	// Launch the beautiful TUI
	if err := tui.StartTUI(cfg, initialQuery, *fuzzyFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
		os.Exit(1)
	}
}

// runDirectUpdate updates the index from the CLI with a loading indicator
func runDirectUpdate(cfg *config.Config) {
	fmt.Println("🌀 Scanning your filesystem in parallel...")
	fmt.Printf("   Roots:   %s\n", strings.Join(cfg.Roots, ", "))
	
	start := time.Now()
	
	progressChan := make(chan int, 10)
	
	// Go indexing task in background
	type result struct {
		count int
		err   error
	}
	resChan := make(chan result)
	go func() {
		count, err := indexer.UpdateIndex(cfg, progressChan)
		resChan <- result{count: count, err: err}
	}()

	// Monitor progress updates in main thread
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	var lastCount int
	var done bool
	var count int
	var updateErr error

	for !done {
		select {
		case r := <-resChan:
			count = r.count
			updateErr = r.err
			done = true
		case c := <-progressChan:
			lastCount = c
			fmt.Printf("\r   Scanned %s files...", formatNumber(lastCount))
		case <-ticker.C:
			// Print progress spin
			fmt.Print(".")
		}
	}

	fmt.Println()

	if updateErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to update index: %v\n", updateErr)
		os.Exit(1)
	}

	duration := time.Since(start)
	fmt.Printf("\n✨ Index built successfully!\n")
	fmt.Printf("   Indexed:  %s files/folders\n", formatNumber(count))
	fmt.Printf("   Duration: %v\n", duration)
}

// runDirectSearch prints the search results directly to stdout (useful for piping or script integration)
func runDirectSearch(cfg *config.Config, query string, limit int, fuzzy bool) {
	start := time.Now()
	results, err := indexer.Search(query, limit, fuzzy)
	duration := time.Since(start)

	if err != nil {
		if err == indexer.ErrNoIndex {
			fmt.Fprintln(os.Stderr, "Error: No search index found. Run 'wtf update' to scan your drive first.")
		} else {
			fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
		}
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No matching files found.")
		return
	}

	// Detect if output is being piped or redirected (to disable terminal ANSI formatting)
	isTerminal := true
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		isTerminal = false
	}

	for _, res := range results {
		if isTerminal {
			// Print beautifully colored path
			fmt.Println(HighlightCommandLinePath(res.Path, res.MatchedIndices))
		} else {
			// Print plain absolute path for scripts/pipes
			fmt.Println(res.Path)
		}
	}

	if isTerminal {
		fmt.Printf("\n🔍 Found %d files in %v (showing top %d)\n", len(results), duration, limit)
	}
}

// runDirectAction performs opening or copying without starting the TUI
func runDirectAction(cfg *config.Config, query string, open, copy, fuzzy bool) {
	results, err := indexer.Search(query, 1, fuzzy)
	if err != nil {
		if err == indexer.ErrNoIndex {
			fmt.Fprintln(os.Stderr, "Error: No search index found. Run 'wtf update' to index first.")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No matching files found.")
		os.Exit(1)
	}

	matchedPath := results[0].Path

	if copy {
		err := clipboard.WriteAll(matchedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📋 Copied to clipboard: %s\n", matchedPath)
	}

	if open {
		err := tui.OpenFile(matchedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🚀 Opening: %s\n", matchedPath)
	}
}

// HighlightCommandLinePath formats path matches for direct terminal printing
func HighlightCommandLinePath(path string, indices []int) string {
	matchedMap := make(map[int]bool)
	for _, idx := range indices {
		matchedMap[idx] = true
	}

	var sb strings.Builder
	dir, _ := filepath.Split(path)

	// Lipgloss styles for direct CLI search
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Bold(true)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)

	for i, r := range path {
		char := string(r)
		if matchedMap[i] {
			sb.WriteString(highlightStyle.Render(char))
		} else {
			if i >= len(dir) {
				sb.WriteString(fileStyle.Render(char))
			} else {
				sb.WriteString(dirStyle.Render(char))
			}
		}
	}

	return sb.String()
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	parts := []string{}
	for n >= 1000 {
		parts = append([]string{fmt.Sprintf("%03d", n%1000)}, parts...)
		n /= 1000
	}
	parts = append([]string{fmt.Sprintf("%d", n)}, parts...)
	return strings.Join(parts, ",")
}
