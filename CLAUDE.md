# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build executable (hide console window on launch)
go build -o claude-code-switcher.exe -ldflags="-H windowsgui" .

# Run tests
go test ./...

# Run single package tests
go test ./internal/fuzzy
go test ./internal/projects
```

## Architecture

Windows-only native GUI application written in Go. Reads Claude Code project data from `~/.claude/projects/` and opens selected projects in Windows Terminal via `wt.exe`.

### Package Structure

- `main.go` - Entry point: loads projects, runs GUI
- `internal/projects/` - Reads `~/.claude/projects/*/sessions-index.json`, decodes paths from encoded directory names (e.g., `c--work-project` -> `c:\work\project`)
- `internal/gui/` - Native Win32 GUI via syscall (user32, gdi32, kernel32). Owner-drawn listbox with custom item rendering. Subclasses edit control for keyboard navigation (arrows/Enter/Escape)
- `internal/fuzzy/` - Fuzzy string matching with scoring (consecutive bonus, word boundary bonus, start-of-text bonus)
- `internal/terminal/` - Launches Windows Terminal via ShellExecute API. Falls back to cmd.exe if wt.exe not found
- `internal/process/` - Process enumeration via Toolhelp32 (currently unused)

### Key Patterns

- All `*_windows.go` files use `//go:build windows` tag
- Win32 API calls via syscall.NewLazyDLL/NewProc pattern
- GUI uses message pump with GetMessageW/TranslateMessage/DispatchMessageW loop
- Project paths encoded as `<drive>--<path-segments-with-dashes>` in Claude's data directory

### Terminal Launching

Uses ShellExecute API (not exec.Command) to launch terminals. Finds claude's full path since app launchers like Everything/Keypirinha don't inherit the user's PATH.

Claude path search order:
1. `exec.LookPath("claude")` - works when launched from shell with proper PATH
2. `~/.local/bin/claude.exe` - official installer location
3. `%APPDATA%/npm/claude.cmd` - npm global install
4. `~/scoop/shims/claude.exe` - scoop install
5. `C:/ProgramData/chocolatey/bin/claude.exe` - chocolatey install

Windows Terminal syntax: `wt.exe -d "path" -- "claude.exe"`. The `--` separator is required to prevent wt.exe from misinterpreting the command as options.

Fallback: `cmd.exe /k cd /d "path" && "claude.exe"` when wt.exe not found. Shows info dialog suggesting Windows Terminal installation.

### Path Decoding

Project paths are encoded in directory names like `c--install-headlines-neutralizer`. The decoder:
1. Tries simple dash-to-separator conversion
2. If path doesn't exist, tries alternative interpretations (e.g., `headlines-neutralizer` as single folder)
3. Validates against filesystem to find correct path

### GUI Behavior

- Closes automatically when losing focus (launcher-style behavior)
- Sort button toggles between "By: Recent" and "By: Name" (Tab key also toggles)
- Keyboard shortcuts: arrows to navigate, Enter to open, Escape to close, F1 for About, Ctrl+Backspace to delete word
- DPI-aware: font sizes and item heights scale with display DPI
- Custom modal About dialog with clickable GitHub link
- Shows error dialog if Claude Code executable not found
- Shows error dialog if no projects exist (`.claude/projects/` missing or empty)
- Shows error dialog if project directory was moved/deleted
