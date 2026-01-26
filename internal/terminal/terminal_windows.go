//go:build windows

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// OpenInWindowsTerminal opens Windows Terminal (or cmd.exe fallback) with the given directory
// and executes 'claude' command
func OpenInWindowsTerminal(projectPath string) error {
	// Check if claude is installed
	claudePath := findClaude()
	if claudePath == "" {
		showErrorDialog("Claude Code Not Found",
			"Claude Code executable was not found.\n\n"+
				"Please install Claude Code and try again.\n\n"+
				"Checked locations:\n"+
				"- ~/.local/bin/\n"+
				"- %APPDATA%/npm/\n"+
				"- ~/scoop/shims/\n"+
				"- C:/ProgramData/chocolatey/bin/")
		return fmt.Errorf("claude executable not found")
	}
	logDebug("Found claude at: %s", claudePath)

	// Check if project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		showErrorDialog("Project Not Found",
			"The project directory no longer exists:\n\n"+projectPath+"\n\n"+
				"It may have been moved or deleted.")
		return fmt.Errorf("project path not found: %s", projectPath)
	}

	// Try Windows Terminal first, fall back to cmd.exe
	wtPath := findWindowsTerminal()
	wtFound := wtPath != ""

	if wtFound {
		logDebug("Found Windows Terminal: %s", wtPath)
		err := launchWithWT(wtPath, projectPath, claudePath)
		if err == nil {
			return nil
		}
		logDebug("Windows Terminal failed: %v, falling back to cmd.exe", err)
	} else {
		logDebug("Windows Terminal not found, using cmd.exe")
		showInfoDialog("Windows Terminal Not Found",
			"Windows Terminal is not installed. Using cmd.exe instead.\n\n"+
				"For a better experience, install Windows Terminal from the Microsoft Store.")
	}

	return launchWithCmd(projectPath, claudePath, wtFound)
}

// launchWithWT launches Windows Terminal directly
func launchWithWT(wtPath, projectPath, claudePath string) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	// Use -- separator to clearly separate wt options from the command
	args := `-d "` + projectPath + `" -- "` + claudePath + `"`

	logDebug("ShellExecute (wt direct): %s %s", wtPath, args)

	ret, _, err := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(wtPath))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(args))),
		0,
		1, // SW_SHOWNORMAL
	)

	logDebug("ShellExecute returned: %d, err: %v", ret, err)

	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed with code %d", ret)
	}
	return nil
}

// launchWithCmd launches plain cmd.exe as fallback
func launchWithCmd(projectPath, claudePath string, wtWasFound bool) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	cmdPath := `C:\Windows\System32\cmd.exe`
	args := `/k cd /d "` + projectPath + `" && "` + claudePath + `"`

	logDebug("ShellExecute (cmd fallback): %s %s", cmdPath, args)

	ret, _, err := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(cmdPath))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(args))),
		0,
		1, // SW_SHOWNORMAL
	)

	logDebug("ShellExecute returned: %d, err: %v", ret, err)

	if ret <= 32 {
		wtStatus := "not found"
		if wtWasFound {
			wtStatus = "failed"
		}
		return fmt.Errorf("could not open terminal.\n\nTried:\n- Windows Terminal (%s)\n- cmd.exe (failed with code %d)\n\nCheck that the project path exists:\n%s", wtStatus, ret, projectPath)
	}
	return nil
}

// findWindowsTerminal returns full path to wt.exe if found
func findWindowsTerminal() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return ""
	}
	wtPath := filepath.Join(localAppData, "Microsoft", "WindowsApps", "wt.exe")
	if _, err := os.Stat(wtPath); err == nil {
		return wtPath
	}
	return ""
}

// findClaude returns full path to claude executable, or empty string if not found
func findClaude() string {
	// Try to find via PATH first (works when launched from proper shell)
	if path, err := exec.LookPath("claude"); err == nil {
		return path
	}

	// Check common installation locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	locations := []string{
		// Official installer location
		filepath.Join(homeDir, ".local", "bin", "claude.exe"),
		filepath.Join(homeDir, ".local", "bin", "claude"),
		// npm global install
		filepath.Join(os.Getenv("APPDATA"), "npm", "claude.cmd"),
		filepath.Join(os.Getenv("APPDATA"), "npm", "claude"),
		// scoop install
		filepath.Join(homeDir, "scoop", "shims", "claude.exe"),
		filepath.Join(homeDir, "scoop", "shims", "claude.cmd"),
		// chocolatey install
		`C:\ProgramData\chocolatey\bin\claude.exe`,
		`C:\ProgramData\chocolatey\bin\claude.cmd`,
	}

	for _, loc := range locations {
		if loc == "" {
			continue
		}
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// showErrorDialog shows an error message box
func showErrorDialog(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	const MB_OK = 0x00000000
	const MB_ICONERROR = 0x00000010

	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		MB_OK|MB_ICONERROR,
	)
}

// showInfoDialog shows an info message box
func showInfoDialog(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	const MB_OK = 0x00000000
	const MB_ICONINFORMATION = 0x00000040

	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		MB_OK|MB_ICONINFORMATION,
	)
}

func logDebug(format string, args ...interface{}) {
	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, "claude-switcher-debug.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
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
