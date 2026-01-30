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

- `main.go` - Entry point: locks OS thread (`runtime.LockOSThread`), loads projects, runs GUI
- `internal/projects/` - Reads `~/.claude/projects/*/sessions-index.json`, extracts cwd from session `.jsonl` files, or decodes paths from encoded directory names (e.g., `c--work-project` -> `c:\work\project`). Validates path existence on disk.
- `internal/gui/` - Native Win32 GUI via syscall (user32, gdi32, kernel32). Owner-drawn listbox with custom item rendering. Subclasses edit control for keyboard navigation (arrows/Enter/Escape)
- `internal/fuzzy/` - Fuzzy string matching with scoring (consecutive bonus, word boundary bonus, start-of-text bonus)
- `internal/terminal/` - Launches Windows Terminal via ShellExecute API. Falls back to cmd.exe if wt.exe not found
- `internal/process/` - Process enumeration via Toolhelp32 (currently unused)

### Key Patterns

- All `*_windows.go` files use `//go:build windows` tag
- Win32 API calls via syscall.NewLazyDLL/NewProc pattern
- GUI uses message pump with GetMessageW/TranslateMessage/DispatchMessageW loop
- `runtime.LockOSThread()` in main - required for Win32 GUI, prevents Go from rescheduling the goroutine to a different OS thread
- Project paths encoded as `<drive>--<path-segments-with-dashes>` in Claude's data directory

### Terminal Launching

Uses ShellExecute API (not exec.Command) to launch terminals. Finds claude's full path since app launchers like Everything/Keypirinha don't inherit the user's PATH.

Claude path search order (checked in this order to avoid PATH lookup delays):
1. `~/.local/bin/claude.exe` - official installer location
3. `%APPDATA%/npm/claude.cmd` - npm global install
4. `~/scoop/shims/claude.exe` - scoop install
5. `C:/ProgramData/chocolatey/bin/claude.exe` - chocolatey install

Windows Terminal syntax: `wt.exe -d "path" -- "claude.exe"`. The `--` separator is required to prevent wt.exe from misinterpreting the command as options.

Fallback: `cmd.exe /k cd /d "path" && "claude.exe"` when wt.exe not found. Shows info dialog suggesting Windows Terminal installation.

### Path Resolution

Project path resolution order:
1. `sessions-index.json` `originalPath` field (most reliable)
2. Session `.jsonl` files `cwd` field (for projects without sessions-index.json)
3. Filesystem-walking path decoder (last resort)

Claude's path encoding converts both path separators (`\`) and dots (`.`) to hyphens. For example, `c:\work\root\fanis.dev` becomes `c--work-root-fanis-dev`. The decoder walks the filesystem recursively, trying each hyphen as a path separator, literal hyphen, or dot at each directory level to find the actual path.

### GUI Behavior

- Closes automatically when losing focus (launcher-style behavior)
- Sort button toggles between "By: Recent" and "By: Name" (Tab key also toggles)
- Keyboard shortcuts: arrows to navigate, Enter to open, Escape to close, F1 for About, Ctrl+Backspace to delete word
- DPI-aware: font sizes and item heights scale with display DPI
- Custom modal About dialog with version and clickable GitHub link
- Projects with missing directories shown as "[NOT FOUND]" with gray styling
- Shows "Opening [project]..." in title bar with disabled UI while launching
- Error dialogs are owned by the main window (proper z-order)
- Shows error dialog if Claude Code executable not found
- Shows error dialog if no projects exist (`.claude/projects/` missing or empty)

## Deployment

See [DEPLOY.md](DEPLOY.md) for full procedure. Summary:

1. Update `CHANGELOG.md` with new version and changes
2. Update version in `README.md`
3. Commit changes
4. Commit "Release X.Y.Z"
5. Tag with `X.Y.Z` (no `v` prefix - required for GitHub Actions)
6. Push commits and tag

Tag format must be `X.Y.Z` (e.g., `0.1.1`) to trigger the release workflow.

## Workflow

- Before committing/pushing, always update docs (CLAUDE.md, CHANGELOG.md, README.md) first
- Always show the full diff for review before pushing
