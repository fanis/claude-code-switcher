// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENCE.md

package fuzzy

import (
	"sort"
	"strings"
	"unicode"
)

// Match performs fuzzy matching of pattern against text
// Returns true if pattern fuzzy-matches text, along with a score (higher is better)
func Match(pattern, text string) (bool, int) {
	if pattern == "" {
		return true, 0
	}

	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	patternIdx := 0
	score := 0
	lastMatchIdx := -1
	consecutiveBonus := 0

	for i, char := range text {
		if patternIdx < len(pattern) && char == rune(pattern[patternIdx]) {
			// First character must match at a word boundary
			if patternIdx == 0 && i > 0 && unicode.IsLetter(rune(text[i-1])) {
				continue
			}

			patternIdx++
			score += 10 // Base score for match

			// Penalty for gaps between matches
			if lastMatchIdx >= 0 {
				gap := i - lastMatchIdx - 1
				if gap > 0 {
					score -= gap * 3
				}
			}

			// Bonus for consecutive matches
			if lastMatchIdx == i-1 {
				consecutiveBonus++
				score += consecutiveBonus * 5
			} else {
				consecutiveBonus = 0
			}

			// Bonus for matching at start of word
			if i == 0 || !unicode.IsLetter(rune(text[i-1])) {
				score += 15
			}

			// Bonus for matching at start of text
			if i == 0 {
				score += 20
			}

			lastMatchIdx = i
		}
	}

	// All pattern characters must be found
	if patternIdx < len(pattern) {
		return false, 0
	}

	// Reject matches where gap penalties outweigh match quality
	if score <= 0 {
		return false, 0
	}

	return true, score
}

// FilterAndScore filters a list of strings by fuzzy matching and returns matched items with scores
func FilterAndScore(pattern string, items []string) []ScoredItem {
	var results []ScoredItem

	for i, item := range items {
		if matched, score := Match(pattern, item); matched {
			results = append(results, ScoredItem{
				Index: i,
				Text:  item,
				Score: score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// ScoredItem represents an item with its fuzzy match score
type ScoredItem struct {
	Index int
	Text  string
	Score int
}
