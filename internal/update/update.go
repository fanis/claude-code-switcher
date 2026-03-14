// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENCE.md

package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fanis/claude-code-switcher/internal/config"
)

type releaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckLatest queries the GitHub API for the latest release.
// Returns the version string, release URL, and any error.
func CheckLatest() (string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/fanis/claude-code-switcher/releases/latest")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	return release.TagName, release.HTMLURL, nil
}

// IsNewer returns true if latest is a higher semver than current.
// Both must be in X.Y.Z format (no "v" prefix).
func IsNewer(current, latest string) bool {
	curParts := parseVersion(current)
	latParts := parseVersion(latest)
	if curParts == nil || latParts == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if latParts[i] > curParts[i] {
			return true
		}
		if latParts[i] < curParts[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) []int {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

// ShouldNotify returns true if the user should be shown an update notification.
func ShouldNotify(cfg *config.Config, latestVersion string) bool {
	if !cfg.UpdateCheckEnabled {
		return false
	}
	if latestVersion == cfg.DismissedVersion {
		return false
	}
	today := time.Now().Format("2006-01-02")
	if cfg.LastCheckDate == today {
		return false
	}
	return true
}
