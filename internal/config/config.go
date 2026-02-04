package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the user configuration for Burrow.
type Config struct {
	DisabledCategories []string `json:"disabled_categories"`
	ExcludedPaths      []string `json:"excluded_paths"`
	SizeThresholdMB    int64    `json:"size_threshold_mb"`
	EnableAuth         bool     `json:"enable_auth"`
}

// Load loads the configuration from ~/.config/burrow/config.json.
func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "burrow")
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil // Return default empty config
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
