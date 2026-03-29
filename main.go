// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENCE.md

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

func utf16PtrFromString(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
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
		utf16PtrFromString(message),
		utf16PtrFromString(title),
		MB_OK|MB_ICONERROR,
	)
}
