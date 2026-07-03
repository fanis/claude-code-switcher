// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

package main

import (
	"errors"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/config"
	"github.com/fanis/claude-code-switcher/internal/gui"
	"github.com/fanis/claude-code-switcher/internal/projects"
)

// utf16PtrFromString returns a *uint16 rather than uintptr so the buffer
// stays referenced until the syscall argument conversion happens at the
// call site (a stored uintptr would not keep it alive).
func utf16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

const appVersion = "0.3.1"

func main() {
	// Win32 GUI operations must all happen on the same OS thread.
	// Without this, Go may reschedule the goroutine to a different thread
	// between window creation and message processing, crashing on first interaction.
	runtime.LockOSThread()
	// Load projects from Claude Code data
	projectList, err := projects.LoadProjects()
	if err != nil {
		if errors.Is(err, projects.ErrNoProjects) {
			showError("No Projects Found",
				"No Claude Code projects were found.\n\n"+
					"Please run Claude Code in a project directory first,\n"+
					"then try again.")
		} else {
			showError("Error Loading Projects", err.Error())
		}
		return
	}

	// Load config (non-fatal if missing)
	cfg, _ := config.Load()

	// Run the GUI
	gui.Run(projectList, appVersion, cfg)
}

func showError(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	const MB_OK = 0x00000000
	const MB_ICONERROR = 0x00000010

	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString(message))),
		uintptr(unsafe.Pointer(utf16PtrFromString(title))),
		MB_OK|MB_ICONERROR,
	)
}
