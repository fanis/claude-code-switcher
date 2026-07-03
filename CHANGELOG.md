# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Crash when pressing Enter with an empty (fully filtered) project list: the "no selection" result of `LB_GETCURSEL` was compared as a 32-bit value and never matched on 64-bit Windows
- Opening the Settings dialog repeatedly leaked a Win32 callback per open (Go callbacks are limited and never released), eventually crashing the app
- Fuzzy search and Ctrl+Backspace word deletion now handle non-ASCII characters correctly (rune-aware matching; UTF-16-aware cursor math)
- Data race between the background update check and the UI thread when reading/saving the config
- Bold/small fonts and the dialog background brush are now released on window destruction (previously leaked)
- Down arrow now selects the first item when nothing is selected instead of doing nothing
- Settings button is now disabled along with the other controls while a project is opening
- Config file is written atomically (temp file + rename) so a crash mid-save can't corrupt it

### Changed
- Project directories are scanned concurrently at startup, so one slow path (e.g. a dead network drive) no longer stalls launch
- List drawing no longer re-resolves GDI/user32 functions on every painted item
- Search strings are cached instead of being rebuilt on every keystroke
- Sorting (projects and search results) is now stable, keeping equal entries in a deterministic order
- Removed unused code: `update.ShouldNotify`, `terminal.showInfoDialog`, `terminal.FocusWindow`
- `decodePath` unit tests now pass on non-Windows hosts too
- Renamed `LICENCE.md` to `LICENSE.md` so packaging tools detect it automatically
- Chocolatey package no longer bundles the binary: it is downloaded at install time from the official GitHub release with a SHA256 checksum, and a `VERIFICATION.txt` is included describing how to verify it

## [0.3.1] - 2026-03-29

### Added
- Configurable terminal emulator: Windows Terminal, WezTerm, cmd.exe, or custom command
- Terminal selector dropdown in Settings dialog
- Custom command support with `{dir}` and `{claude}` placeholders
- Tab key cycles focus between controls in Settings dialog
- Welcome message on first-launch onboarding prompt

### Changed
- Auto-detect mode now tries Windows Terminal, then WezTerm, then cmd.exe (was: wt then cmd)
- Built-in WezTerm launching now prefers `wezterm.exe` from PATH, opens tabs via `start --new-tab`, hides the helper console window, and suppresses the default handoff info log
- Built-in WezTerm and custom `wezterm start ...` commands now share the same silent launch behavior

## [0.3.0] - 2026-03-14

### Added
- Settings dialog (gear icon button, F1) replaces About dialog
- Optional update notifications via GitHub Releases (opt-in, checked once per day)
- Two-phase update check: background check on current launch, notification on next launch
- First-launch onboarding prompt asking about update notifications
- Config file at `~/.claude-code-switcher/config.json`

### Changed
- Windows Terminal opens new tab in existing window instead of a new window (`-w 0 nt`)
- Use `exec.Command` instead of `ShellExecute` for Windows Terminal launch
- Remove info dialog when Windows Terminal is not installed (silent fallback to cmd.exe)

## [0.2.3] - 2026-03-11

### Fixed
- Fix "last used" showing stale timestamps when sessions-index.json is outdated (now also checks .jsonl modtimes)
- Improve fuzzy search: add gap penalty, require first character at word boundary, reject low-quality scattered matches

## [0.2.2] - 2026-03-10

### Fixed
- Fix "last used" timestamps using directory modtime instead of session file modtime (showed wrong times)
- Fix listbox rendering artifacts when resizing window (missing InvalidateRect)

### Changed
- Move timestamp display to first line (right-aligned) for better visibility
- Enforce minimum window size of 400x200 pixels
- Build output now targets dist/ directory

## [0.2.1] - 2026-03-06

### Fixed
- Fix command injection vulnerability: reject project paths containing double quotes
- Fix text rendering bug for non-ASCII project names (drawText byte count vs character count)
- Replace deprecated `syscall.StringToUTF16Ptr` with safe `UTF16PtrFromString` wrapper

### Changed
- Debug logging now gated behind `CLAUDE_SWITCHER_DEBUG` env var (was always on)
- Replace O(n^2) bubble sort with `sort.Slice` in fuzzy matching
- Remove non-functional "in use" detection (process package still unused)
- Remove dead code: unused variables, structs, and constants

## [0.2.0] - 2026-01-30

### Fixed
- Fix crash on first interaction after build (add runtime.LockOSThread for Win32 GUI)
- Fix error dialogs appearing behind main window (proper owner HWND)
- Fix duplicate error dialog when opening a missing project
- Fix path decoding for folder names containing dots (e.g., fanis.dev)

### Added
- Visual "[NOT FOUND]" marker with gray styling for projects whose directories no longer exist
- "Opening..." indicator in title bar while launching a terminal
- Version displayed in About dialog
- Extract project paths from session .jsonl files when sessions-index.json is missing
- Recursive filesystem-walking path decoder that handles dots and hyphens

### Changed
- Remove always-on-top (WS_EX_TOPMOST) for standard launcher z-order behavior

## [0.1.1] - 2026-01-27

### Fixed
- Fix hang on first click when launched from app launchers (e.g., Everything, Keypirinha)
  - Check known installation paths before PATH lookup to avoid network drive delays

## [0.1.0] - 2026-01-26

### Added
- Initial release
- Native Win32 GUI for minimal startup time
- Fuzzy search to filter projects by name or path
- Sort by recent use (default) or alphabetically by name (Tab to toggle)
- Opens selected project in Windows Terminal (falls back to cmd.exe if not installed)
- Launcher-style behavior: closes automatically when losing focus
- DPI-aware: scales properly on high-DPI displays
- Keyboard shortcuts:
  - Tab: Toggle sort mode
  - Arrow keys: Navigate list
  - Enter: Open project
  - Escape: Close
  - Ctrl+Backspace: Delete word in search
  - F1: About dialog
- Smart path decoding handles folder names with hyphens
- Graceful error handling:
  - Shows error if Claude Code not found (checks multiple install locations)
  - Shows error if no projects exist
  - Shows error if project directory was moved/deleted
  - Shows info dialog suggesting Windows Terminal installation when using cmd.exe fallback
