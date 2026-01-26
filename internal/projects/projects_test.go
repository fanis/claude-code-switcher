package projects

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDecodePath(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected string
	}{
		{
			name:     "simple path",
			encoded:  "c--work",
			expected: "c:\\work",
		},
		{
			name:     "nested path with single dashes",
			encoded:  "c--work-root-project",
			expected: "c:\\work\\root\\project",
		},
		{
			name:     "user directory",
			encoded:  "C--Users-micro",
			expected: "C:\\Users\\micro",
		},
		{
			name:     "empty string",
			encoded:  "",
			expected: "",
		},
		{
			name:     "drive only",
			encoded:  "c",
			expected: "c:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodePath(tt.encoded)
			if result != tt.expected {
				t.Errorf("decodePath(%q) = %q, want %q", tt.encoded, result, tt.expected)
			}
		})
	}
}

func TestLoadProjectInfo(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "claude-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test sessions-index.json
	sessionsIndex := SessionsIndex{
		Version: 1,
		Entries: []SessionEntry{
			{
				SessionID:   "test-session-1",
				Modified:    "2026-01-26T10:00:00Z",
				ProjectPath: "c:\\work\\test-project",
			},
			{
				SessionID:   "test-session-2",
				Modified:    "2026-01-25T10:00:00Z",
				ProjectPath: "c:\\work\\test-project",
			},
		},
		OriginalPath: "c:\\work\\test-project",
	}

	data, err := json.Marshal(sessionsIndex)
	if err != nil {
		t.Fatalf("Failed to marshal sessions index: %v", err)
	}

	sessionsFile := filepath.Join(tmpDir, "sessions-index.json")
	if err := os.WriteFile(sessionsFile, data, 0644); err != nil {
		t.Fatalf("Failed to write sessions file: %v", err)
	}

	// Test loading project info
	path, lastUsed, err := loadProjectInfo(sessionsFile)
	if err != nil {
		t.Fatalf("loadProjectInfo() error = %v", err)
	}

	if path != "c:\\work\\test-project" {
		t.Errorf("loadProjectInfo() path = %v, want %v", path, "c:\\work\\test-project")
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-01-26T10:00:00Z")
	if !lastUsed.Equal(expectedTime) {
		t.Errorf("loadProjectInfo() lastUsed = %v, want %v", lastUsed, expectedTime)
	}
}

func TestSortByLastUsed(t *testing.T) {
	now := time.Now()
	projects := []Project{
		{Name: "old", LastUsed: now.Add(-24 * time.Hour)},
		{Name: "newest", LastUsed: now},
		{Name: "middle", LastUsed: now.Add(-1 * time.Hour)},
	}

	SortByLastUsed(projects)

	if projects[0].Name != "newest" {
		t.Errorf("First project should be 'newest', got %v", projects[0].Name)
	}
	if projects[1].Name != "middle" {
		t.Errorf("Second project should be 'middle', got %v", projects[1].Name)
	}
	if projects[2].Name != "old" {
		t.Errorf("Third project should be 'old', got %v", projects[2].Name)
	}
}

func TestSortByName(t *testing.T) {
	projects := []Project{
		{Name: "Zebra"},
		{Name: "apple"},
		{Name: "Banana"},
	}

	SortByName(projects)

	if projects[0].Name != "apple" {
		t.Errorf("First project should be 'apple', got %v", projects[0].Name)
	}
	if projects[1].Name != "Banana" {
		t.Errorf("Second project should be 'Banana', got %v", projects[1].Name)
	}
	if projects[2].Name != "Zebra" {
		t.Errorf("Third project should be 'Zebra', got %v", projects[2].Name)
	}
}
