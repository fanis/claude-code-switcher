package main

import (
	"os"

	"github.com/fhatzidakis/claude-code-switcher/internal/gui"
	"github.com/fhatzidakis/claude-code-switcher/internal/projects"
)

func main() {
	// Load projects from Claude Code data
	projectList, err := projects.LoadProjects()
	if err != nil {
		os.Exit(1)
	}

	// Run the GUI
	gui.Run(projectList)
}
