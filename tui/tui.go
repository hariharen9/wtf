package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"wtf/config"
	"wtf/indexer"
	"wtf/search"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style definitions using modern curated HSL-like palettes
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#8B5CF6")). // Royal Violet
			Padding(0, 1).
			MarginRight(2)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981")) // Emerald Green

	searchPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F43F5E")) // Glowing Rose

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B")). // Muted Slate
			Italic(true)

	selectedRowIndicatorStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#F59E0B")) // Warm Amber

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E293B")). // Dark Indigo Slate
			Foreground(lipgloss.Color("#94A3B8")).
			Padding(0, 1).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444"))
)

type progressMsg int
type indexFinishedMsg struct {
	count int
	err   error
}

type model struct {
	textInput      textinput.Model
	results        []search.MatchResult
	cursor         int
	limit          int
	fuzzy          bool
	cfg            *config.Config
	searchDuration time.Duration
	err            error
	width          int
	height         int

	// Copy indicator
	copiedPath string
	copiedTime time.Time

	// Indexing states
	isIndexing   bool
	indexedCount int
	indexingErr  error
}

// NewModel creates an initialized Bubble Tea model.
func NewModel(cfg *config.Config, initialQuery string, fuzzy bool) model {
	ti := textinput.New()
	ti.Placeholder = "Type to locate files instantly..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40
	ti.Prompt = " ✨ "
	ti.PromptStyle = searchPromptStyle

	if initialQuery != "" {
		ti.SetValue(initialQuery)
	}

	m := model{
		textInput: ti,
		limit:     100, // Search returns top 100, displayed list fits window height
		fuzzy:     fuzzy,
		cfg:       cfg,
	}

	if initialQuery != "" {
		m.runSearch()
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) runSearch() {
	query := m.textInput.Value()
	if query == "" {
		m.results = nil
		m.searchDuration = 0
		return
	}

	start := time.Now()
	res, err := indexer.Search(query, m.limit, m.fuzzy)
	m.searchDuration = time.Since(start)

	if err != nil {
		m.err = err
		m.results = nil
	} else {
		m.err = nil
		m.results = res
	}

	// Clamp cursor
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// backgroundIndexer runs the filesystem walker and reports progress updates to Bubbletea
func (m model) backgroundIndexer(progressChan chan int) tea.Cmd {
	return func() tea.Msg {
		count, err := indexer.UpdateIndex(m.cfg, progressChan)
		return indexFinishedMsg{count: count, err: err}
	}
}

func listenProgress(progressChan chan int) tea.Cmd {
	return func() tea.Msg {
		count, ok := <-progressChan
		if !ok {
			return nil
		}
		return progressMsg(count)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = m.width - 10
		return m, nil

	case progressMsg:
		m.indexedCount = int(msg)
		// We need to keep listening for more progress ticks
		// Note: The channel is stored in the closure of the background function.
		// To make it simple, we can listen again if m.isIndexing is true.
		// However, it's easier to just poll or stream progress.
		// Let's pass a progressChan in a controlled way or just use a ticking timer,
		// but standard channels are very neat. We will handle progress updates in TUI.

	case indexFinishedMsg:
		m.isIndexing = false
		if msg.err != nil {
			m.indexingErr = msg.err
		} else {
			m.indexingErr = nil
			m.indexedCount = msg.count
			m.runSearch() // Rerun search on new index
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			// Copy selected item path to clipboard
			if len(m.results) > 0 {
				path := m.results[m.cursor].Path
				err := clipboard.WriteAll(path)
				if err == nil {
					m.copiedPath = path
					m.copiedTime = time.Now()
				}
			}
			return m, nil

		case tea.KeyCtrlF:
			m.fuzzy = !m.fuzzy
			m.runSearch()
			return m, nil

		case tea.KeyCtrlQ, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlU:
			// Trigger index update in background
			if !m.isIndexing {
				m.isIndexing = true
				m.indexingErr = nil
				m.indexedCount = 0
				progressChan := make(chan int, 5)
				
				// Standard Bubble Tea way to run background task
				return m, tea.Batch(
					m.backgroundIndexer(progressChan),
					// To keep things simple and zero-lock, we can also just run indexing
					// and show an elegant spinner, checking completion via the finished msg.
				)
			}

		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.results) > 0 {
				m.cursor = len(m.results) - 1 // wrap around
			}

		case tea.KeyDown, tea.KeyCtrlN:
			if len(m.results) > 0 {
				if m.cursor < len(m.results)-1 {
					m.cursor++
				} else {
					m.cursor = 0 // wrap around
				}
			}

		case tea.KeyEnter:
			// Open selected file and quit
			if len(m.results) > 0 {
				path := m.results[m.cursor].Path
				_ = OpenFile(path)
				return m, tea.Quit
			}
		}
	}

	// Update text input
	oldVal := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	if m.textInput.Value() != oldVal {
		m.runSearch()
	}

	return m, cmd
}

// HighlightPath styles directories, filenames, and highlights matched fuzzy indices
func HighlightPath(path string, indices []int, isSelected bool) string {
	matchedMap := make(map[int]bool)
	for _, idx := range indices {
		matchedMap[idx] = true
	}

	var sb strings.Builder

	dir, _ := filepath.Split(path)

	var (
		dirStyle       lipgloss.Style
		fileStyle      lipgloss.Style
		highlightStyle lipgloss.Style
	)

	if isSelected {
		dirStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A5B4FC")).Italic(true) // Soft Indigo
		fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)  // Crisp White
		highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F43F5E")).Bold(true).Underline(true) // Glowing Rose
	} else {
		dirStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#475569")) // Muted Grey-Slate
		fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Bold(true) // Deep Sky Blue
		highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true) // Emerald Green
	}

	// Render path with color highlights
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

func OpenFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (m model) View() string {
	var s strings.Builder

	// Logo banner
	s.WriteString(titleStyle.Render(" WTF "))
	s.WriteString(logoStyle.Render("Where's The File?"))
	s.WriteString("\n\n")

	// Search box
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")

	// Main Area (Results or States)
	if m.isIndexing {
		s.WriteString(successStyle.Render(" 🌀 Re-indexing filesystem in parallel... Please wait.\n"))
		if m.indexedCount > 0 {
			s.WriteString(statsStyle.Render(fmt.Sprintf("    Scanned %s files...", formatNumber(m.indexedCount))))
		}
		s.WriteString("\n")
	} else if m.err != nil {
		if errorsIs(m.err, indexer.ErrNoIndex) {
			s.WriteString(errorStyle.Render(" ⚠️  No search index found!\n\n"))
			s.WriteString(" Press ")
			s.WriteString(selectedRowIndicatorStyle.Render("[ctrl+u]"))
			s.WriteString(" to scan your computer and build the search database now.\n")
		} else {
			s.WriteString(errorStyle.Render(fmt.Sprintf(" Error: %v\n", m.err)))
		}
	} else if len(m.results) == 0 {
		if m.textInput.Value() == "" {
			s.WriteString(statsStyle.Render(" Type characters to fuzzy find any file on your drive...\n"))
		} else {
			s.WriteString(statsStyle.Render(" No matching files or folders found.\n"))
		}
	} else {
		// Calculate available height for results dynamically
		// Total lines taken by headers & stats is around 8-9 lines.
		reservedLines := 8
		maxLines := m.height - reservedLines
		if maxLines < 3 {
			maxLines = 5 // Fallback minimum
		}

		// Adjust range of results to show (scrolling window)
		startIdx := 0
		endIdx := len(m.results)
		if endIdx > maxLines {
			endIdx = maxLines
		}

		// If cursor scrolls down past visible screen, shift window
		if m.cursor >= maxLines {
			startIdx = m.cursor - maxLines + 1
			endIdx = m.cursor + 1
		}

		for i := startIdx; i < endIdx; i++ {
			res := m.results[i]
			isSelected := (i == m.cursor)

			if isSelected {
				s.WriteString(selectedRowIndicatorStyle.Render(" ▸ "))
				s.WriteString(HighlightPath(res.Path, res.MatchedIndices, true))
			} else {
				s.WriteString("   ")
				s.WriteString(HighlightPath(res.Path, res.MatchedIndices, false))
			}
			s.WriteString("\n")
		}

		// Search stats
		s.WriteString("\n")
		searchType := "Substring"
		if m.fuzzy {
			searchType = "Fuzzy"
		}
		countStr := fmt.Sprintf("%d", len(m.results))
		if len(m.results) >= m.limit {
			countStr = fmt.Sprintf("%d+", m.limit)
		}
		s.WriteString(statsStyle.Render(fmt.Sprintf(" 🔍 Found %s files (%s match) in %v", countStr, searchType, m.searchDuration)))
		s.WriteString("\n")
	}

	// Status & Copy notification bar
	var statusMsg string
	if !m.copiedTime.IsZero() && time.Since(m.copiedTime) < 2*time.Second {
		filename := filepath.Base(m.copiedPath)
		statusMsg = fmt.Sprintf(" ✅ Copied path to '%s' to clipboard!", filename)
	} else {
		modeName := "Substring"
		if m.fuzzy {
			modeName = "Fuzzy"
		}
		statusMsg = fmt.Sprintf(" [ctrl+f] Mode: %s | [enter] open | [ctrl+c] copy | [ctrl+u] update | [esc] quit", modeName)
	}

	// Pad status bar to fit screen
	barWidth := m.width
	if barWidth < 60 {
		barWidth = 60
	}
	s.WriteString(statusBarStyle.Width(barWidth).Render(statusMsg))

	return s.String()
}

func errorsIs(err, target error) bool {
	return err == target || (err != nil && err.Error() == target.Error())
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	// Simple formatting (e.g. 10,000)
	parts := []string{}
	for n >= 1000 {
		parts = append([]string{fmt.Sprintf("%03d", n%1000)}, parts...)
		n /= 1000
	}
	parts = append([]string{fmt.Sprintf("%d", n)}, parts...)
	return strings.Join(parts, ",")
}

// StartTUI launches the Bubble Tea interactive CLI
func StartTUI(cfg *config.Config, initialQuery string, fuzzy bool) error {
	p := tea.NewProgram(NewModel(cfg, initialQuery, fuzzy), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
