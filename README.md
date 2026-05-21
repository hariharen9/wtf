# ⚡ WTF — Where's The File?

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Platform Support](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-8A2BE2?style=for-the-badge)](https://github.com)
[![Built with Bubble Tea](https://img.shields.io/badge/Built%20With-Bubble%20Tea-F43F5E?style=for-the-badge)](https://github.com/charmbracelet/bubbletea)

> The insanely fast, cross-platform CLI alternative to Windows' *Everything*. Locate, copy, or open any file on your drive in milliseconds.

---

## 🤔 Why WTF Exists

As developers, we are constantly hunting for files. But our current toolsets force us to make painful compromises:
*   **Traditional search tools** (`find`, `fd`) perform a live directory crawl every time you query. Searching an entire drive takes seconds or minutes, stalling your momentum.
*   **Indexed search tools** (`locate`, `plocate`) rely on background databases that update only once a day. If you created a file 5 minutes ago, they won't find it.
*   **Platform-specific giants** (*Everything* on Windows) are incredible, but lock you into a single operating system and require a mouse and GUI to navigate.

**WTF bridges the gap.** It is a single, zero-dependency CLI utility that gives you the instantaneous indexing power of *Everything*, the smart fuzzy filtering of `fzf`, and a gorgeous interactive interface—working uniformly across **Windows, macOS, and Linux**.

---

## 🔥 How It Beats the Rest

### ⚡ Insanely Fast Indexing
WTF utilizes a highly concurrent multi-threaded filesystem walker. It bypasses noise (like `node_modules`, `.git`, and system temp folders) to traverse **160,000+ files per second**. Rebuilding your index takes less than a blink.

### 🧠 Instant Smart Search
By loading a flat, lightweight index straight into memory and splitting the lookup across your CPU cores, WTF performs queries in **under 2 milliseconds**. 
*   **Contiguous Substring Matching (Default):** Exact sequence matching for maximum precision. Typing `hari` instantly locates `hariharen` and filters out noisy matches.
*   **Fuzzy Matching (Optional):** Typo-tolerant, discovery search that understands directory boundaries when you only remember parts of a filename. Easily toggleable inside the TUI!
*   **Smart Case:** Automatically switches to case-sensitive matching the moment you type a capital letter.

### 🎨 Stunning Terminal Aesthetics
Built using the state-of-the-art **Charmbracelet (Bubble Tea & Lipgloss)** terminal ecosystem, WTF features a premium color scheme out of the box:
*   Muted grey directory structures keep layouts clean.
*   Bright sky-blue filenames jump out at you.
*   Emerald green highlights instantly show you *exactly* which characters matched your query.

### 🔌 Seamless Shell Integration
WTF is fully pipeline-ready. Run it interactively as a terminal UI, or use it in your scripts. It automatically detects if its output is being piped to another tool and strips out ANSI styling to supply raw, clean paths.

---

## 🕹️ How to Use It

Launch the gorgeous interactive search TUI by running:
```bash
wtf
```
*Start typing your query immediately. Use `↑`/`↓` keys to navigate.*

### Interactive Shortcuts
*   `Enter` — Instantly **opens** the selected file or folder in your system's default editor or viewer.
*   `Ctrl + C` — **Copies** the absolute path of the selected file to your clipboard.
*   `Ctrl + F` — **Toggles** between contiguous Substring matching (Default) and Fuzzy search modes dynamically.
*   `Ctrl + U` — **Re-indexes** your filesystem in the background without leaving the app.
*   `Esc` / `Ctrl + Q` — Quit.

---

## 💻 CLI Commands (For the Power Users)

Run tasks or fetch files directly from your command line without launching the full TUI:

*   **Update the index:**
    ```bash
    wtf update
    ```
*   **Direct search (prints to stdout):**
    ```bash
    wtf search app.js
    ```
*   **Instantly open a file by query:**
    ```bash
    wtf -o main.go
    ```
*   **Instantly copy a path by query:**
    ```bash
    wtf -c config
    ```

---

## 📦 Installation

WTF can be installed instantly with zero external dependencies using our high-impact custom installers:

### 🚀 Recommended (macOS & Linux)
Install the optimized native binary directly into `~/.wtf/bin` via curl:
```bash
curl -fsSL https://raw.githubusercontent.com/hariharen9/wtf/main/install.sh | sh
```

### ⚡ Recommended (Windows)
Install the native binary directly and automatically configure your system `PATH` via PowerShell:
```powershell
irm https://raw.githubusercontent.com/hariharen9/wtf/main/install.ps1 | iex
```

---

### 📦 Package Managers

If you prefer using your ecosystem's native packaging tools, WTF is also fully distributed across all major registries:

#### 1. Homebrew (macOS & Linux)
To install using Homebrew, tap the repository and install:
```bash
brew tap hariharen9/wtf
brew install wtf
```

#### 2. npm (Node.js Ecosystem)
To install the global CLI wrapper via npm:
```bash
npm install -g wtf-cli
```

#### 3. pip (Python Ecosystem)
To install the global CLI wrapper via pip:
```bash
pip install wtf-cli
```

#### 4. Native Go (Compile from Source)
If you have Go installed, you can compile and install it directly to your `$GOPATH/bin` with:
```bash
go install github.com/hariharen9/wtf@latest
```
*Make sure your local `$GOPATH/bin` (typically `~/go/bin` or `%USERPROFILE%\go\bin`) is in your system's `PATH`!*

---

## 🚀 Getting Started

Once installed, generate your initial search database index by running:
```bash
wtf update
```
And that's it! You are ready to locate any file on your computer in sub-milliseconds.

---

<p align="center">
  Made with ⚡ by the WTF team.
</p>
