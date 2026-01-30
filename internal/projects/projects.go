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
	PathExists bool      // Whether the project directory exists on disk
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
		projectDir := filepath.Join(projectsDir, encodedName)
		sessionsFile := filepath.Join(projectDir, "sessions-index.json")

		// Try to get project path and last used time from sessions-index.json
		projectPath, lastUsed, err := loadProjectInfo(sessionsFile)
		if err != nil {
			// Try reading cwd from a session .jsonl file
			projectPath = extractCwdFromSessions(projectDir)
			if projectPath == "" {
				// Last resort: decode the path from directory name
				projectPath = decodePath(encodedName)
			}
			if projectPath == "" {
				continue
			}
			// Use directory modification time
			info, err := entry.Info()
			if err == nil {
				lastUsed = info.ModTime()
			}
		}

		_, statErr := os.Stat(projectPath)
		project := Project{
			Name:       filepath.Base(projectPath),
			Path:       projectPath,
			EncodedDir: encodedName,
			LastUsed:   lastUsed,
			PathExists: statErr == nil,
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

// sessionMessage represents the minimal structure of a session .jsonl entry
// that contains cwd information
type sessionMessage struct {
	Cwd string `json:"cwd"`
}

// extractCwdFromSessions reads the first .jsonl file in the project directory
// and extracts the cwd field, which contains the actual project path.
// This is used when sessions-index.json is not available.
func extractCwdFromSessions(projectDir string) string {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		f, err := os.Open(filepath.Join(projectDir, entry.Name()))
		if err != nil {
			continue
		}

		// Read enough to find the cwd field in the first few lines
		buf := make([]byte, 4096)
		n, _ := f.Read(buf)
		f.Close()

		if n == 0 {
			continue
		}

		// Parse each line as JSON looking for cwd
		for _, line := range strings.Split(string(buf[:n]), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var msg sessionMessage
			if err := json.Unmarshal([]byte(line), &msg); err == nil && msg.Cwd != "" {
				return msg.Cwd
			}
		}
	}

	return ""
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
// and returns the first one that exists on disk.
//
// Claude's path encoding converts both path separators (\) and dots (.)
// to hyphens. For example, "c:\work\root\fanis.dev" becomes
// "c--work-root-fanis-dev". This function walks the filesystem recursively,
// trying each hyphen as a path separator, literal hyphen, or dot to find
// the actual path.
func findValidPath(drive, pathPart string) string {
	segments := strings.Split(pathPart, "-")
	if len(segments) == 0 {
		return ""
	}

	return resolveSegments(drive, segments)
}

// resolveSegments recursively resolves path segments by trying different
// joiners (path separator, hyphen, dot) at each level, validating against
// the filesystem at each step.
func resolveSegments(basePath string, segments []string) string {
	if len(segments) == 0 {
		// All segments consumed - check if this path exists
		if _, err := os.Stat(basePath); err == nil {
			return basePath
		}
		return ""
	}

	// Try consuming 1..N segments as a single directory/file name.
	// For each group size, try joining with different separators.
	for count := 1; count <= len(segments); count++ {
		group := segments[:count]
		rest := segments[count:]

		// Try different joiners between segments in this group:
		// - hyphen (literal hyphen in folder name)
		// - dot (e.g., "fanis.dev")
		// For single segments, no joiner needed.
		var names []string
		if count == 1 {
			names = []string{group[0]}
		} else {
			hyphenName := strings.Join(group, "-")
			dotName := strings.Join(group, ".")
			names = []string{hyphenName, dotName}
		}

		for _, name := range names {
			candidate := basePath + string(filepath.Separator) + name
			if _, err := os.Stat(candidate); err == nil {
				if len(rest) == 0 {
					return candidate
				}
				// This level exists, recurse for remaining segments
				result := resolveSegments(candidate, rest)
				if result != "" {
					return result
				}
			}
		}
	}

	return ""
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
