# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
