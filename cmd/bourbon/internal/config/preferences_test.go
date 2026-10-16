package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferencesSaveLoadReset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_prefs_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldEnv := os.Getenv("BOURBON_CONFIG_DIR")
	os.Setenv("BOURBON_CONFIG_DIR", tempDir)
	defer os.Setenv("BOURBON_CONFIG_DIR", oldEnv)

	// Test default when none exists
	prefs, err := LoadPreferences()
	if err != nil {
		t.Fatalf("expected nil error on missing preferences, got: %v", err)
	}
	if prefs.Profile != "fullstack" || prefs.Structure != "feature" {
		t.Errorf("unexpected default prefs: %+v", prefs)
	}

	// Test save
	prefs.Profile = "api"
	prefs.Structure = "layered"
	prefs.Database = "sqlite"
	prefs.Cache = "memory"
	prefs.Auth = "jwt"
	prefs.Capabilities = []string{"openapi", "observability"}

	if err := SavePreferences(prefs); err != nil {
		t.Fatalf("failed to save preferences: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "defaults.json")); err != nil {
		t.Fatalf("defaults.json was not created: %v", err)
	}

	// Test load saved
	loaded, err := LoadPreferences()
	if err != nil {
		t.Fatalf("failed to load saved preferences: %v", err)
	}
	if loaded.Profile != "api" || loaded.Structure != "layered" || loaded.Database != "sqlite" {
		t.Errorf("loaded preferences mismatch: %+v", loaded)
	}

	// Test reset
	if err := ResetPreferences(); err != nil {
		t.Fatalf("failed to reset preferences: %v", err)
	}

	resetLoaded, err := LoadPreferences()
	if err != nil {
		t.Fatalf("failed to load after reset: %v", err)
	}
	if resetLoaded.Profile != "fullstack" {
		t.Errorf("expected default profile after reset, got: %s", resetLoaded.Profile)
	}
}
