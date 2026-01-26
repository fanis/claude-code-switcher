package projects

import (
	"encoding/json"
	"fmt"
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

// ErrNoProjects indicates the .claude/projects directory doesn't exist
var ErrNoProjects = fmt.Errorf("no Claude Code projects found")

// LoadProjects loads all Claude Code projects from ~/.claude/projects/
func LoadProjects() ([]Project, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoProjects
		}
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ErrNoProjects
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
// Since we can't distinguish path separators from literal hyphens in folder names,
// we try multiple interpretations and verify against the filesystem
func decodePath(encoded string) string {
	if encoded == "" {
		return ""
	}

	// The encoding uses double dashes to separate the drive from the rest
	// Single dashes could be path separators OR literal hyphens in folder names
	// Example: "c--work-root-project" -> "c:\work\root\project"
	// Example: "c--install-headlines-neutralizer" could be:
	//   "c:\install\headlines\neutralizer" OR "c:\install\headlines-neutralizer"

	parts := strings.Split(encoded, "--")
	if len(parts) < 1 {
		return ""
	}

	// Drive only case (e.g., "c")
	if len(parts) == 1 {
		return parts[0] + ":"
	}

	// First part is the drive letter
	drive := parts[0] + ":"

	// Get the path portion (everything after --)
	pathPart := strings.Join(parts[1:], "--") // Rejoin in case there were multiple --

	// Try to find a valid path by testing different interpretations
	path := findValidPath(drive, pathPart)
	if path != "" {
		return path
	}

	// Fallback: simple dash-to-separator conversion (original behavior)
	result := drive
	segments := strings.Split(pathPart, "-")
	for _, seg := range segments {
		if seg != "" {
			result += string(filepath.Separator) + seg
		}
	}
	return result
}

// findValidPath tries different interpretations of the encoded path
// and returns the first one that exists on disk
func findValidPath(drive, pathPart string) string {
	segments := strings.Split(pathPart, "-")
	if len(segments) == 0 {
		return ""
	}

	// Try all possible groupings of segments
	// Start with most likely: each segment is a folder (original behavior)
	// Then try: last N segments are one folder name with hyphens
	// Then try: all segments are one folder name

	// Generate candidate paths by trying different groupings
	candidates := generatePathCandidates(drive, segments)

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// generatePathCandidates generates possible path interpretations
// prioritizing more likely folder structures
func generatePathCandidates(drive string, segments []string) []string {
	var candidates []string
	n := len(segments)

	if n == 0 {
		return candidates
	}

	// Try: all segments as separate folders (most common)
	path := drive
	for _, seg := range segments {
		path += string(filepath.Separator) + seg
	}
	candidates = append(candidates, path)

	// Try: last 2+ segments joined with hyphens (e.g., "headlines-neutralizer")
	for joinFrom := n - 2; joinFrom >= 0; joinFrom-- {
		path := drive
		for i := 0; i < joinFrom; i++ {
			path += string(filepath.Separator) + segments[i]
		}
		joined := strings.Join(segments[joinFrom:], "-")
		path += string(filepath.Separator) + joined
		candidates = append(candidates, path)
	}

	return candidates
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
