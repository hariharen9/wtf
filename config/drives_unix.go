//go:build !windows
package config

// getWindowsDrives returns nil on Unix platforms.
func getWindowsDrives() []string {
	return nil
}
