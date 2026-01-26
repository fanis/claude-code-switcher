package main

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/gui"
	"github.com/fanis/claude-code-switcher/internal/projects"
)

func main() {
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

	// Run the GUI
	gui.Run(projectList)
}

func showError(title, message string) {
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
