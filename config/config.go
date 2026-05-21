package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Roots          []string `json:"roots"`
	IgnorePatterns []string `json:"ignore_patterns"`
	MaxDepth       int      `json:"max_depth"`
	FollowSymlinks bool     `json:"follow_symlinks"`
}

// DefaultConfig returns the default configuration based on the OS.
func DefaultConfig() *Config {
	var roots []string
	var ignores []string

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	if runtime.GOOS == "windows" {
		roots = getWindowsDrives()
		ignores = []string{
			`node_modules`,
			`.git`,
			`.idea`,
			`.vscode`,
			`.cache`,
			`AppData\Local\Temp`,
			`AppData\Local\Microsoft`,
			`AppData\Local\Packages`,
			`Program Files\WindowsApps`,
			`ProgramData\Microsoft`,
			`System Volume Information`,
			`$Recycle.Bin`,
			`$WinREAgent`,
			`Windows`,
			`Microsoft`,
		}
	} else if runtime.GOOS == "darwin" {
		roots = []string{home}
		ignores = []string{
			"node_modules",
			".git",
			".idea",
			".vscode",
			"Library/Caches",
			"Library/Logs",
			"Library/Application Support",
			"Library/Containers",
			"Library/Metadata",
		}
	} else { // Linux & others
		roots = []string{home}
		ignores = []string{
			"node_modules",
			".git",
			".idea",
			".vscode",
			".cache",
			".local/share/Trash",
			"tmp",
		}
	}

	return &Config{
		Roots:          roots,
		IgnorePatterns: ignores,
		MaxDepth:       50,
		FollowSymlinks: false,
	}
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "wtf", "config.json"), nil
}

// LoadConfig loads the configuration from disk, creating a default one if it doesn't exist.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		cfg := DefaultConfig()
		if err := SaveConfig(cfg); err != nil {
			return cfg, fmt.Errorf("failed to save default config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to disk.
func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ShouldIgnore checks if a path contains any of the ignored patterns.
func (cfg *Config) ShouldIgnore(path string) bool {
	// Normalize path separators to forward slashes for cross-platform pattern matching
	normalized := strings.ReplaceAll(path, "\\", "/")
	
	for _, pattern := range cfg.IgnorePatterns {
		// Normalize pattern separators
		normPattern := strings.ReplaceAll(pattern, "\\", "/")
		
		// 1. Exact match of a segment (e.g. "/.git/" or "/node_modules/")
		if strings.Contains("/"+normalized+"/", "/"+normPattern+"/") {
			return true
		}
		
		// 2. Simple prefix or suffix matching
		if strings.HasSuffix(normalized, normPattern) || strings.HasPrefix(normalized, normPattern) {
			return true
		}
	}
	return false
}
