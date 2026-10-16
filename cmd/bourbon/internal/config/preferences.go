// Package config provides project manifest and remembered developer preferences.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Preferences represents locally remembered developer choices on the machine.
type Preferences struct {
	Profile      string   `json:"profile"`
	Structure    string   `json:"structure"`
	Database     string   `json:"database"`
	Cache        string   `json:"cache"`
	Auth         string   `json:"auth"`
	Capabilities []string `json:"capabilities"`
}

// DefaultPreferences returns the sensible out-of-the-box defaults.
func DefaultPreferences() *Preferences {
	return &Preferences{
		Profile:      "fullstack",
		Structure:    "feature",
		Database:     "postgres",
		Cache:        "redis",
		Auth:         "session",
		Capabilities: []string{"jobs", "scheduling", "mail", "storage", "openapi", "observability"},
	}
}

// GetPreferencesPath returns the path to the defaults.json configuration file.
func GetPreferencesPath() string {
	if custom := os.Getenv("BOURBON_CONFIG_DIR"); custom != "" {
		return filepath.Join(custom, "defaults.json")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".bourbon", "defaults.json")
}

// LoadPreferences reads remembered developer choices from disk or returns default preferences.
func LoadPreferences() (*Preferences, error) {
	path := GetPreferencesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultPreferences(), nil
		}
		return nil, err
	}

	prefs := DefaultPreferences()
	if err := json.Unmarshal(data, prefs); err != nil {
		return DefaultPreferences(), nil
	}

	if prefs.Profile == "" {
		prefs.Profile = "fullstack"
	}
	if prefs.Structure == "" {
		prefs.Structure = "feature"
	}
	if prefs.Database == "" {
		prefs.Database = "postgres"
	}
	if prefs.Cache == "" {
		prefs.Cache = "redis"
	}
	if prefs.Auth == "" {
		prefs.Auth = "session"
	}

	return prefs, nil
}

// SavePreferences persists developer choices to the local machine.
func SavePreferences(prefs *Preferences) error {
	if prefs == nil {
		return nil
	}

	path := GetPreferencesPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ResetPreferences deletes the locally remembered developer choices.
func ResetPreferences() error {
	path := GetPreferencesPath()
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
