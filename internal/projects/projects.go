package projects

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Project represents a Claude Code project
type Project struct {
	Name       string    // Last component of the path
	Path       string    // Full path to the project directory
	LastUsed   time.Time // Last modified time
	InUse      bool      // Whether Claude is currently running in this project
	EncodedDir string    // The encoded directory name in .claude/projects/
}

// SessionsIndex represents the sessions-index.json structure
type SessionsIndex struct {
	Version      int            `json:"version"`
	Entries      []SessionEntry `json:"entries"`
	OriginalPath string         `json:"originalPath"`
}

// SessionEntry represents a single session entry
type SessionEntry struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	Summary     string `json:"summary"`
	Modified    string `json:"modified"`
	ProjectPath string `json:"projectPath"`
}

// LoadProjects loads all Claude Code projects from ~/.claude/projects/
func LoadProjects() ([]Project, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		encodedName := entry.Name()
		sessionsFile := filepath.Join(projectsDir, encodedName, "sessions-index.json")

		// Try to get project path and last used time from sessions-index.json
		projectPath, lastUsed, err := loadProjectInfo(sessionsFile)
		if err != nil {
			// Fall back to decoding the path from directory name
			projectPath = decodePath(encodedName)
			if projectPath == "" {
				continue
			}
			// Use directory modification time
			info, err := entry.Info()
			if err == nil {
				lastUsed = info.ModTime()
			}
		}

		project := Project{
			Name:       filepath.Base(projectPath),
			Path:       projectPath,
			EncodedDir: encodedName,
			LastUsed:   lastUsed,
		}

		projects = append(projects, project)
	}

	// Sort by last used (most recent first) by default
	SortByLastUsed(projects)

	return projects, nil
}

// loadProjectInfo reads sessions-index.json and returns the project path and last used time
func loadProjectInfo(filePath string) (string, time.Time, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", time.Time{}, err
	}

	var index SessionsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return "", time.Time{}, err
	}

	// Get project path from originalPath or first entry's projectPath
	projectPath := index.OriginalPath
	if projectPath == "" && len(index.Entries) > 0 {
		projectPath = index.Entries[0].ProjectPath
	}
	if projectPath == "" {
		return "", time.Time{}, os.ErrNotExist
	}

	// Find the most recent modified time
	var latest time.Time
	for _, entry := range index.Entries {
		if entry.Modified == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, entry.Modified)
		if err != nil {
			continue
		}
		if t.After(latest) {
			latest = t
		}
	}

	return projectPath, latest, nil
}

// decodePath converts an encoded path like "c--work-root-project" to a path
// This is a fallback when sessions-index.json is not available
func decodePath(encoded string) string {
	if encoded == "" {
		return ""
	}

	// The encoding uses double dashes to separate the drive from the rest
	// and single dashes within sections for path separators
	// Example: "c--work-root-project" -> "c:\work\root\project"
	// Example: "C--Users-micro" -> "C:\Users\micro"

	parts := strings.Split(encoded, "--")
	if len(parts) < 1 {
		return ""
	}

	// First part is the drive letter
	result := parts[0] + ":"

	for i := 1; i < len(parts); i++ {
		part := parts[i]
		// Single dashes within a section are path separators
		subParts := strings.Split(part, "-")
		for _, subPart := range subParts {
			if subPart != "" {
				result += string(filepath.Separator) + subPart
			}
		}
	}

	return result
}

// SortByLastUsed sorts projects by last used time (most recent first)
func SortByLastUsed(projects []Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUsed.After(projects[j].LastUsed)
	})
}

// SortByName sorts projects alphabetically by name
func SortByName(projects []Project) {
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
}
