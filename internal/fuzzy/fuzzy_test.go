package fuzzy

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		text        string
		wantMatch   bool
		wantMinScore int
	}{
		{
			name:        "empty pattern matches anything",
			pattern:     "",
			text:        "anything",
			wantMatch:   true,
			wantMinScore: 0,
		},
		{
			name:        "exact match",
			pattern:     "test",
			text:        "test",
			wantMatch:   true,
			wantMinScore: 40,
		},
		{
			name:        "prefix match",
			pattern:     "tes",
			text:        "testing",
			wantMatch:   true,
			wantMinScore: 30,
		},
		{
			name:        "fuzzy match with gaps",
			pattern:     "tst",
			text:        "testing",
			wantMatch:   true,
			wantMinScore: 20,
		},
		{
			name:        "case insensitive",
			pattern:     "TEST",
			text:        "testing",
			wantMatch:   true,
			wantMinScore: 20,
		},
		{
			name:        "no match",
			pattern:     "xyz",
			text:        "testing",
			wantMatch:   false,
			wantMinScore: 0,
		},
		{
			name:        "partial pattern not found",
			pattern:     "testx",
			text:        "testing",
			wantMatch:   false,
			wantMinScore: 0,
		},
		{
			name:        "word boundary bonus",
			pattern:     "cc",
			text:        "claude-code",
			wantMatch:   true,
			wantMinScore: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotScore := Match(tt.pattern, tt.text)
			if gotMatch != tt.wantMatch {
				t.Errorf("Match() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
			if gotMatch && gotScore < tt.wantMinScore {
				t.Errorf("Match() gotScore = %v, want at least %v", gotScore, tt.wantMinScore)
			}
		})
	}
}

func TestFilterAndScore(t *testing.T) {
	items := []string{
		"claude-code-switcher",
		"trading-newsletter",
		"headlines-neutralizer",
		"test-project",
	}

	tests := []struct {
		name       string
		pattern    string
		wantCount  int
		wantFirst  string
	}{
		{
			name:      "empty pattern returns all",
			pattern:   "",
			wantCount: 4,
		},
		{
			name:      "filter with strong match first",
			pattern:   "test",
			wantCount: 2, // test-project and headlines-neutralizer both match
			wantFirst: "test-project",
		},
		{
			name:      "fuzzy filter",
			pattern:   "ccs",
			wantCount: 1,
			wantFirst: "claude-code-switcher",
		},
		{
			name:      "multiple matches sorted by score",
			pattern:   "t",
			wantCount: 4, // all have 't'
		},
		{
			name:      "no matches",
			pattern:   "xyz",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := FilterAndScore(tt.pattern, items)
			if len(results) != tt.wantCount {
				t.Errorf("FilterAndScore() returned %d results, want %d", len(results), tt.wantCount)
			}
			if tt.wantFirst != "" && len(results) > 0 && results[0].Text != tt.wantFirst {
				t.Errorf("FilterAndScore() first result = %v, want %v", results[0].Text, tt.wantFirst)
			}
		})
	}
}
