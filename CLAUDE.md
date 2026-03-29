# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build executable (hide console window on launch)
go build -o dist/claude-code-switcher.exe -ldflags="-H windowsgui" .

# Run tests
go test ./...

# Run single package tests
go test ./internal/fuzzy
go test ./internal/projects
```

## Architecture

Windows-only native GUI application written in Go. Reads Claude Code project data from `~/.claude/projects/` and opens selected projects in a configurable terminal emulator.

### Package Structure

- `main.go` - Entry point: locks OS thread (`runtime.LockOSThread`), loads projects, runs GUI
- `internal/projects/` - Reads `~/.claude/projects/*/sessions-index.json`, extracts cwd from session `.jsonl` files, or decodes paths from encoded directory names (e.g., `c--work-project` -> `c:\work\project`). Validates path existence on disk. "Last used" comes from session `modified` field (sessions-index.json) or `.jsonl` file modtime (fallback) - never use directory modtime (unreliable, changed by Claude housekeeping).
- `internal/gui/` - Native Win32 GUI via syscall (user32, gdi32, kernel32, comctl32). Owner-drawn listbox with custom item rendering. Subclasses edit control for keyboard navigation (arrows/Enter/Escape)
- `internal/fuzzy/` - Fuzzy string matching with scoring (consecutive bonus, word boundary bonus, start-of-text bonus)
- `internal/terminal/` - Configurable terminal launching. Supports Windows Terminal, WezTerm, cmd.exe, and custom commands. Auto-detect mode tries wt -> wezterm -> cmd. Custom commands support `{dir}` and `{claude}` placeholders
- `internal/config/` - JSON config persistence at `~/.claude-code-switcher/config.json` (update check preferences, pending version state, terminal selection)
- `internal/update/` - Lightweight GitHub Releases API check, semver comparison, notification dedup logic (max once/day, dismissed versions tracked)
- `internal/process/` - Process enumeration via Toolhelp32 (currently unused, kept for future "active project" detection)

### Key Patterns

- All `*_windows.go` files use `//go:build windows` tag
- Win32 API calls via syscall.NewLazyDLL/NewProc pattern
- GUI uses message pump with GetMessageW/TranslateMessage/DispatchMessageW loop
- `runtime.LockOSThread()` in main - required for Win32 GUI, prevents Go from rescheduling the goroutine to a different OS thread
- Project paths encoded as `<drive>--<path-segments-with-dashes>` in Claude's data directory
- Always use `utf16PtrFromString` wrapper for Win32 string args, never `syscall.StringToUTF16Ptr` (deprecated, silently truncates at embedded nulls)
- `DrawTextW` length param: pass `-1` for null-terminated strings, never `len(text)` (byte count, wrong for non-ASCII)
- Paths interpolated into ShellExecute args must be validated (reject `"` chars) to prevent command injection
- `go vet` warns "possible misuse of unsafe.Pointer" on Win32 lParam casts (DRAWITEMSTRUCT, MEASUREITEMSTRUCT) - this is expected and unavoidable
- No external dependencies (no go.sum) - GitHub Actions cache warning about missing go.sum is harmless

### Terminal Launching

Configurable via `terminal` field in config.json and the Settings dialog dropdown. Finds claude's full path since app launchers like Everything/Keypirinha don't inherit the user's PATH. Debug logging writes to `~/claude-switcher-debug.log`.

Terminal modes:
- `""` (auto-detect): tries wt -> wezterm -> cmd in order
- `"wt"`: Windows Terminal only
- `"wezterm"`: WezTerm only
- `"cmd"`: cmd.exe only
- Custom string: run as-is with optional `{dir}` and `{claude}` placeholders (no placeholders = run verbatim)

Claude path search order (checked in this order to avoid PATH lookup delays):
1. `~/.local/bin/claude.exe` - official installer location
2. `%APPDATA%/npm/claude.cmd` - npm global install
3. `~/scoop/shims/claude.exe` - scoop install
4. `C:/ProgramData/chocolatey/bin/claude.exe` - chocolatey install
5. `exec.LookPath("claude")` - last resort PATH lookup (may be slow with network drives)

Each location also checks alternative extensions (.exe, .cmd, or extensionless) where applicable.

Windows Terminal: found via `%LOCALAPPDATA%/Microsoft/WindowsApps/wt.exe`. Syntax: `wt.exe -w 0 nt -d "path" -- "claude.exe"`. The `-w 0` reuses the most recent WT window, `nt` opens a new tab, `--` prevents wt.exe from misinterpreting the command as options. Tab reuse requires the switcher to run at the same elevation level as the existing WT window.

WezTerm: found via PATH lookup first, then `C:\Program Files\WezTerm\wezterm.exe`. Syntax: `wezterm.exe start --new-tab --cwd "path" -- "claude.exe"`. Reuses an existing WezTerm GUI window when possible and starts a new one otherwise. The switcher launches WezTerm via `exec.Command`, hides the helper console window with `SysProcAttr.HideWindow`, and sets `WEZTERM_LOG=error` to suppress the default handoff info log.

cmd.exe fallback: `cmd.exe /k cd /d "path" && "claude.exe"` via ShellExecute.

### Path Resolution

Project path resolution order:
1. `sessions-index.json` `originalPath` field (most reliable)
2. Session `.jsonl` files `cwd` field (for projects without sessions-index.json)
3. Filesystem-walking path decoder (last resort)

Claude's path encoding converts both path separators (`\`) and dots (`.`) to hyphens. For example, `c:\work\root\fanis.dev` becomes `c--work-root-fanis-dev`. The decoder walks the filesystem recursively, trying each hyphen as a path separator, literal hyphen, or dot at each directory level to find the actual path.

### GUI Behavior

- Closes automatically when losing focus (launcher-style behavior)
- Sort button toggles between "By: Recent" and "By: Name" (Tab key also toggles)
- Keyboard shortcuts: arrows to navigate, Enter to open, Escape to close, F1 for Settings, Ctrl+Backspace to delete word
- DPI-aware: font sizes and item heights scale with display DPI
- Settings dialog (gear icon button): update check toggle, terminal selector (dropdown + custom command input), about info. Uses `WM_CTLCOLORSTATIC` for white backgrounds, `IsDialogMessageW` for Tab focus cycling
- One-time onboarding prompt on first launch asking about update notifications (`asked_about_updates` config flag)
- Two-phase update check: background goroutine writes `pending_version`/`pending_url` to config, next launch shows notification via `WM_APP_UPDATE` custom message
- List items: project name + right-aligned timestamp on first line, path on second line
- Owner-drawn listbox requires `InvalidateRect` after `MoveWindow` on resize (items won't repaint correctly otherwise)
- Minimum window size (400x200) enforced via `WM_GETMINMAXINFO` handler
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

## Debug Logging

Set `CLAUDE_SWITCHER_DEBUG=1` environment variable to enable debug logging to `~/claude-switcher-debug.log`.

## Workflow

- Before committing/pushing, always update docs (CLAUDE.md, CHANGELOG.md, README.md) first
- Always show the full diff for review before pushing
- Version lives in three places: `main.go` (appVersion const), `README.md` (badge), `CHANGELOG.md` (new section)
- Git identity rule: ask the user `work` or `personal` before the first commit/push in a repo unless the identity is already explicit. For personal repos, use the user's personal identity and `github.com-personal` SSH host in remotes when setting or changing remotes.
