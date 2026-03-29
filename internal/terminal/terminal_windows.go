// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENCE.md

//go:build windows

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func utf16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// parentHwnd is the owner window for message boxes, set via SetParentHwnd
var parentHwnd uintptr

// SetParentHwnd sets the parent window handle for dialogs shown by this package
func SetParentHwnd(hwnd uintptr) {
	parentHwnd = hwnd
}

// OpenProject opens a terminal with the given directory and executes claude.
// The terminalSetting controls which terminal to use:
//   - "" (empty): auto-detect (wt -> wezterm -> cmd)
//   - "wt": Windows Terminal
//   - "wezterm": WezTerm
//   - "cmd": cmd.exe
//   - anything else: custom command with optional {dir} and {claude} placeholders
func OpenProject(projectPath, terminalSetting string) error {
	logDebug("OpenProject called for: %s (terminal=%q)", projectPath, terminalSetting)

	// Reject paths containing double quotes to prevent command injection
	if strings.Contains(projectPath, `"`) {
		return fmt.Errorf("project path contains invalid character: %s", projectPath)
	}

	switch terminalSetting {
	case "", "wt", "wezterm", "cmd":
		return openWithPreset(projectPath, terminalSetting)
	default:
		return openWithCustom(projectPath, terminalSetting)
	}
}

// openWithPreset handles built-in terminal presets and auto-detection.
func openWithPreset(projectPath, preset string) error {
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

	if preset == "wt" {
		wtPath := findWindowsTerminal()
		if wtPath == "" {
			return fmt.Errorf("Windows Terminal not found")
		}
		return launchWithWT(wtPath, projectPath, claudePath)
	}

	if preset == "wezterm" {
		weztermPath := findWezTerm()
		if weztermPath == "" {
			return fmt.Errorf("WezTerm not found")
		}
		return launchWithWezTerm(weztermPath, projectPath, claudePath)
	}

	if preset == "cmd" {
		return launchWithCmd(projectPath, claudePath, false)
	}

	// Auto-detect: wt -> wezterm -> cmd
	wtPath := findWindowsTerminal()
	if wtPath != "" {
		logDebug("Auto: found Windows Terminal: %s", wtPath)
		if err := launchWithWT(wtPath, projectPath, claudePath); err == nil {
			return nil
		}
		logDebug("Auto: Windows Terminal failed, trying next")
	}

	weztermPath := findWezTerm()
	if weztermPath != "" {
		logDebug("Auto: found WezTerm: %s", weztermPath)
		if err := launchWithWezTerm(weztermPath, projectPath, claudePath); err == nil {
			return nil
		}
		logDebug("Auto: WezTerm failed, trying next")
	}

	logDebug("Auto: falling back to cmd.exe")
	return launchWithCmd(projectPath, claudePath, wtPath != "" || weztermPath != "")
}

// openWithCustom launches a custom terminal command with placeholder substitution.
// Supported placeholders: {dir} for project path, {claude} for claude executable path.
// If no placeholders are present, the command is run as-is.
func openWithCustom(projectPath, command string) error {
	// Only find claude if the command uses {claude}
	expandedCmd := command
	if strings.Contains(command, "{claude}") {
		claudePath := findClaude()
		if claudePath == "" {
			showErrorDialog("Claude Code Not Found",
				"Claude Code executable was not found.\n\n"+
					"Please install Claude Code and try again.")
			return fmt.Errorf("claude executable not found")
		}
		expandedCmd = strings.ReplaceAll(expandedCmd, "{claude}", claudePath)
	}
	expandedCmd = strings.ReplaceAll(expandedCmd, "{dir}", projectPath)

	logDebug("Custom terminal: %s", expandedCmd)

	// Split into executable and arguments
	// Handles quoted paths: "C:\Program Files\term.exe" -arg1 -arg2
	parts := splitCommand(expandedCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty terminal command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if isWezTermStart(parts) {
		cmd.Env = append(os.Environ(), "WEZTERM_LOG=error")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	logDebug("exec.Command: %s %v", parts[0], parts[1:])

	return cmd.Start()
}

func isWezTermStart(parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	base := strings.ToLower(filepath.Base(parts[0]))
	return (base == "wezterm" || base == "wezterm.exe") && strings.EqualFold(parts[1], "start")
}

// splitCommand splits a command string into executable and arguments,
// respecting double-quoted segments.
func splitCommand(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ' ' && !inQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(ch)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// launchWithWT launches Windows Terminal directly using exec.Command
func launchWithWT(wtPath, projectPath, claudePath string) error {
	// -w 0: reuse the most recent WT window instead of opening a new one
	// nt: explicitly open a new tab
	// --: separator to prevent wt from misinterpreting the command as options
	cmd := exec.Command(wtPath, "-w", "0", "nt", "-d", projectPath, "--", claudePath)
	logDebug("exec.Command: %s %v", wtPath, cmd.Args[1:])

	err := cmd.Start()
	if err != nil {
		logDebug("exec.Command failed: %v", err)
		return err
	}
	return nil
}

// launchWithWezTerm launches WezTerm, opening a new tab in an existing GUI
// window or starting a new GUI window if none exists.
func launchWithWezTerm(weztermPath, projectPath, claudePath string) error {
	cmd := exec.Command(weztermPath, "start", "--new-tab", "--cwd", projectPath, "--", claudePath)
	cmd.Env = append(os.Environ(), "WEZTERM_LOG=error")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	logDebug("exec.Command: %s %v", weztermPath, cmd.Args[1:])

	err := cmd.Start()
	if err != nil {
		logDebug("exec.Command failed: %v", err)
		return err
	}
	return nil
}

// launchWithCmd launches plain cmd.exe as fallback
func launchWithCmd(projectPath, claudePath string, otherWasFound bool) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	cmdPath := `C:\Windows\System32\cmd.exe`
	args := `/k cd /d "` + projectPath + `" && "` + claudePath + `"`

	logDebug("ShellExecute (cmd fallback): %s %s", cmdPath, args)

	ret, _, err := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("open"))),
		uintptr(unsafe.Pointer(utf16PtrFromString(cmdPath))),
		uintptr(unsafe.Pointer(utf16PtrFromString(args))),
		0,
		1, // SW_SHOWNORMAL
	)

	logDebug("ShellExecute returned: %d, err: %v", ret, err)

	if ret <= 32 {
		otherStatus := "not found"
		if otherWasFound {
			otherStatus = "failed"
		}
		return fmt.Errorf("could not open terminal.\n\nTried:\n- Other terminals (%s)\n- cmd.exe (failed with code %d)\n\nCheck that the project path exists:\n%s", otherStatus, ret, projectPath)
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

// findWezTerm returns full path to wezterm.exe if found
func findWezTerm() string {
	// Prefer PATH so users can override the installed copy.
	if path, err := exec.LookPath("wezterm"); err == nil {
		return path
	}

	// Common install location
	path := `C:\Program Files\WezTerm\wezterm.exe`
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// findClaude returns full path to claude executable, or empty string if not found
func findClaude() string {
	// Check common installation locations first (fast, predictable paths)
	// This avoids exec.LookPath which can hang if PATH contains network drives
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logDebug("findClaude: failed to get home dir: %v", err)
		return ""
	}

	appData := os.Getenv("APPDATA")

	var locations []string

	// Official installer location
	locations = append(locations,
		filepath.Join(homeDir, ".local", "bin", "claude.exe"),
		filepath.Join(homeDir, ".local", "bin", "claude"),
	)

	// npm global install (only if APPDATA is set)
	if appData != "" {
		locations = append(locations,
			filepath.Join(appData, "npm", "claude.cmd"),
			filepath.Join(appData, "npm", "claude"),
		)
	}

	// scoop install
	locations = append(locations,
		filepath.Join(homeDir, "scoop", "shims", "claude.exe"),
		filepath.Join(homeDir, "scoop", "shims", "claude.cmd"),
	)

	// chocolatey install
	locations = append(locations,
		`C:\ProgramData\chocolatey\bin\claude.exe`,
		`C:\ProgramData\chocolatey\bin\claude.cmd`,
	)

	for _, loc := range locations {
		logDebug("findClaude: checking %s", loc)
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Last resort: try PATH lookup (may be slow if PATH has network drives)
	logDebug("findClaude: trying exec.LookPath")
	if path, err := exec.LookPath("claude"); err == nil {
		return path
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
		parentHwnd,
		uintptr(unsafe.Pointer(utf16PtrFromString(message))),
		uintptr(unsafe.Pointer(utf16PtrFromString(title))),
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
		parentHwnd,
		uintptr(unsafe.Pointer(utf16PtrFromString(message))),
		uintptr(unsafe.Pointer(utf16PtrFromString(title))),
		MB_OK|MB_ICONINFORMATION,
	)
}

func logDebug(format string, args ...interface{}) {
	if os.Getenv("CLAUDE_SWITCHER_DEBUG") == "" {
		return
	}

	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, "claude-switcher-debug.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	msg := fmt.Sprintf(format, args...)
	f.WriteString(fmt.Sprintf("%s\n", msg))
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
