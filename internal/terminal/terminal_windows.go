//go:build windows

package terminal

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// OpenInWindowsTerminal opens a new tab in Windows Terminal with the given directory
// and executes 'claude' command
func OpenInWindowsTerminal(projectPath string) error {
	// Find wt.exe in common locations
	wtPath := findWindowsTerminal()
	if wtPath == "" {
		wtPath = "wt.exe" // Fall back to PATH
	}

	// Windows Terminal syntax: wt.exe -w 0 nt -d "path" -- cmd.exe /k claude
	// -w 0 = use existing window (or create if none)
	// nt = new-tab (short form)
	// -d = starting directory
	// -- = separator before command
	cmd := exec.Command(wtPath,
		"-w", "0",
		"nt",
		"-d", projectPath,
		"--",
		"cmd.exe", "/k", "claude",
	)

	// Hide the console window for this launcher process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	return cmd.Start()
}

// findWindowsTerminal looks for wt.exe in common locations
func findWindowsTerminal() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return ""
	}

	// Check Windows Terminal from Microsoft Store
	wtPaths := []string{
		filepath.Join(localAppData, "Microsoft", "WindowsApps", "wt.exe"),
	}

	for _, p := range wtPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
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
