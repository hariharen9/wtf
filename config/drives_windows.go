//go:build windows
package config

import (
	"syscall"
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives = modkernel32.NewProc("GetLogicalDrives")
)

// getWindowsDrives discovers and returns all active logical drives on Windows.
func getWindowsDrives() []string {
	var drives []string
	
	// Call GetLogicalDrives via kernel32.dll
	r1, _, _ := procGetLogicalDrives.Call()
	if r1 == 0 {
		return []string{"C:\\"} // Safe fallback
	}
	
	bitmask := uint32(r1)
	
	for i := 0; i < 26; i++ {
		if (bitmask & (1 << uint(i))) != 0 {
			// Skip A:\ and B:\ (floppy drives) to avoid slowing down or prompting dialogs
			if i < 2 {
				continue
			}
			driveLetter := string(rune('A'+i)) + ":\\"
			drives = append(drives, driveLetter)
		}
	}
	
	if len(drives) == 0 {
		return []string{"C:\\"}
	}
	return drives
}
