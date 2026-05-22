# wheretf (WTF — Where's The File?)

**WTF** is a blazing-fast, cross-platform interactive terminal file finder and CLI searcher written in pure Go.

This NPM package is a lightweight global wrapper that automatically downloads and caches the optimized, pre-compiled native WTF binary for your specific operating system (Windows, macOS, or Linux) and CPU architecture during installation, then exposes it to your terminal with zero overhead.

---

## ⚡ Installation

Install the package globally via npm:

```bash
npm install -g wheretf
```

---

## 🚀 Usage

Once installed, simply type **`wtf`** in your terminal:

```bash
wtf
```

You can also start a search query directly:

```bash
wtf main.go
```

---

## 🔍 Features

*   **Insanely Fast:** Scans 160,000+ files per second using concurrent Go routines.
*   **Zero Dependencies:** Runs on a pre-compiled native binary built for your OS/CPU.
*   **Fuzzy Filtering:** Full fuzzy matching as you type.
*   **Beautiful Terminal UI:** Styled with high-fidelity Charm Bubble Tea interactive layouts.

For full documentation, source code, and release notes, visit the [Main GitHub Repository](https://github.com/hariharen9/wtf).
