// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENCE.md

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application settings persisted to disk.
type Config struct {
	UpdateCheckEnabled bool   `json:"update_check_enabled"`
	AskedAboutUpdates  bool   `json:"asked_about_updates"`
	DismissedVersion   string `json:"dismissed_version"`
	LastCheckDate      string `json:"last_check_date"`
	PendingVersion     string `json:"pending_version"`
	PendingURL         string `json:"pending_url"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude-code-switcher"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file, returning defaults if it doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return &Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, nil
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	return os.WriteFile(path, data, 0644)
}
