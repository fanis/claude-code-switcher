//go:build windows

package terminal

import (
	"os/exec"
	"syscall"
)

// OpenInWindowsTerminal opens a new tab in Windows Terminal with the given directory
// and executes 'claude' command
func OpenInWindowsTerminal(projectPath string) error {
	// wt.exe -w 0 new-tab -d "path" cmd /k claude
	// -w 0 means use the most recently used window (or create new if none exists)
	cmd := exec.Command("wt.exe",
		"-w", "0",
		"new-tab",
		"-d", projectPath,
		"cmd", "/k", "claude",
	)

	// Hide the console window for this launcher process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	return cmd.Start()
}

// FocusWindow brings the window with the given HWND to the foreground
func FocusWindow(hwnd uintptr) error {
	if hwnd == 0 {
		return nil
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	setForegroundWindow := user32.NewProc("SetForegroundWindow")
	showWindow := user32.NewProc("ShowWindow")

	const SW_RESTORE = 9

	// Restore if minimized
	showWindow.Call(hwnd, SW_RESTORE)

	// Bring to foreground
	setForegroundWindow.Call(hwnd)

	return nil
}
