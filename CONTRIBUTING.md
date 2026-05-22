# Contributing to WTF ⚡

First off, thank you for taking the time to contribute to **WTF** ("Where's The File")! It is people like you who make WTF a blazing-fast, premium utility for developers worldwide.

Please take a moment to read this guide to understand our repository architecture, design patterns, development cycle, and how to get your contributions successfully merged.

---

## 🛠️ Codebase Philosophy & Architecture

WTF is built on three core pillars:
1. **Insane Performance:** Code must be optimized for execution speed and zero heap allocations during lookups.
2. **Zero Dependencies (CGO-free):** It must remain 100% pure Go. Cross-compilation must work seamlessly from any host without external C compilers.
3. **Stunning Terminal UX:** The interactive search screen (Bubble Tea) must feel immediate, fluid, and visually delightful.

### File Structure
- [main.go](file:///e:/Projects/wtf/main.go): The entrypoint routing CLI flags, commands, and launching the TUI.
- [config/](file:///e:/Projects/wtf/config/): Configuration load/save pipelines, ignore rules, and platform-specific drive discovery (e.g. dynamic CGO-free win32 DLL syscalls).
- [indexer/](file:///e:/Projects/wtf/indexer/): The concurrent directory traversal pool (`walker.go`) and chunked parallel search/atomic index writer (`index.go`).
- [search/](file:///e:/Projects/wtf/search/): Matching algorithms (contiguous exact substring and fzf-inspired sequence-scored fuzzy matching).
- [tui/](file:///e:/Projects/wtf/tui/): The gorgeous interactive terminal UI built using Charmbracelet's Bubble Tea (MVU) and Lipgloss frameworks.

---

## 🚀 Setting Up Your Workspace

### Prerequisites
- **Go 1.22+** (Go 1.26 recommended) installed on your machine.
- A terminal of your choice supporting true-color ANSI formatting.

### Dev Installation
1. Clone your fork of WTF:
   ```bash
   git clone https://github.com/YOUR_USERNAME/wtf.git
   cd wtf
   ```
2. Verify you can build and run it locally:
   ```bash
   go run main.go
   ```

---

## ⚙️ Development Lifecycle

### 1. Running Unit Tests
All search scoring, substring indexers, and path highlights have complete unit tests under `search/`. Run them before submitting changes:
```bash
go test -v ./...
```

### 2. Formatting Code
Keep the Go code idiomatic and clean. Ensure it is formatted before commits:
```bash
go fmt ./...
```

### 3. Compiling Statically Linked Binaries
We strip symbols and DWARF debug info to shrink the final binary size by over 40%:
```bash
go build -ldflags "-s -w" -o wtf main.go
```

### 4. Cross-Compilation Testing
You can compile WTF for other systems from your current OS:
```bash
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o wtf-darwin-arm64 main.go

# Windows (x64)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o wtf-windows-amd64.exe main.go

# Linux (AMD64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o wtf-linux-amd64 main.go
```

---

## 🎨 Coding Conventions & Best Practices

1. **Avoid CGO at all costs:** If platform-specific hooks are required (like accessing Windows logical drive letters), use Go's native `syscall` or `golang.org/x/sys` packages rather than binding C modules.
2. **Allocation-Free Matching:** The byte slices inside `indexer/index.go` must be scanned directly on boundary index indicators. Only convert matching path substrings into Go string variables to avoid garbage collection overhead.
3. **Keep Bubble Tea Viewport Responsive:** Ensure that dynamic line calculations in `tui.go` always handle terminal size adjustments via `tea.WindowSizeMsg` and clamp viewport index ranges safely to avoid panic indexes.
4. **ANSI Safety in Pipelines:** When writing outputs directly to stdout, always check `os.Stdout.Stat()` to determine if the terminal output device is character-based (`os.ModeCharDevice`). If the tool is piped or redirected, color escape sequences must be stripped out entirely.

---

## 📥 Submitting Your Pull Request (PR)

1. Create a descriptive feature branch for your edits:
   ```bash
   git checkout -b feature/awesome-speedup
   ```
2. Write clean commits following standard semantic commit messages (e.g. `feat: improve fuzzy match gap penalties`, `fix: clamp tui scroll offset on viewport shrink`).
3. Push to your branch and open a Pull Request against WTF's main branch.
4. Ensure all CI/CD workflows on GitHub Actions pass.

Thank you for helping us keep WTF blazing-fast and accessible to everyone! ⚡
