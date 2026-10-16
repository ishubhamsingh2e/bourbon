package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCommands(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_config_cmd_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "cfg")
	os.Setenv("BOURBON_CONFIG_DIR", configDir)
	defer os.Unsetenv("BOURBON_CONFIG_DIR")

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	// Test config show
	code := runner.Run([]string{"config", "show"})
	if code != 0 {
		t.Fatalf("config show failed: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Profile:") {
		t.Errorf("expected stdout to contain 'Profile:', got: %s", stdout.String())
	}

	// Test config reset
	stdout.Reset()
	code = runner.Run([]string{"config", "reset"})
	if code != 0 {
		t.Fatalf("config reset failed: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Successfully reset") {
		t.Errorf("expected stdout to contain 'Successfully reset', got: %s", stdout.String())
	}
}
