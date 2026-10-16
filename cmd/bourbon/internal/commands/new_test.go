package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunNewCommandWithDefaults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_cli_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "bourbon_config")
	os.Setenv("BOURBON_CONFIG_DIR", configDir)
	defer os.Unsetenv("BOURBON_CONFIG_DIR")

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{
		"new", "demo-app",
		"--defaults",
		"--profile=api",
		"--structure=feature",
		"--db=sqlite",
		"--cache=memory",
		"--auth=jwt",
	})
	if code != 0 {
		t.Fatalf("runNew failed with code %d, stderr: %s", code, stderr.String())
	}

	// Verify generated modular config files
	expectedFiles := []string{
		"demo-app/.env",
		"demo-app/.env.example",
		"demo-app/bourbon.yaml",
		"demo-app/go.mod",
		"demo-app/config/settings.go",
		"demo-app/internal/platform/container.go",
		"demo-app/internal/modules/system/routes.go",
		"demo-app/internal/modules/users/controller.go",
		"demo-app/internal/modules/users/service.go",
		"demo-app/internal/modules/users/repository.go",
		"demo-app/internal/modules/users/model.go",
		"demo-app/internal/modules/users/routes.go",
		"demo-app/main.go",
		"demo-app/main_test.go",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("expected file missing: %s (%v)", file, err)
		}
	}
}

func TestRunNewCommandLayeredStructure(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_layered_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "bourbon_config")
	os.Setenv("BOURBON_CONFIG_DIR", configDir)
	defer os.Unsetenv("BOURBON_CONFIG_DIR")

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{
		"new", "layered-app",
		"--defaults",
		"--profile=fullstack",
		"--structure=layered",
		"--no-remember",
	})
	if code != 0 {
		t.Fatalf("runNew layered failed with code %d, stderr: %s", code, stderr.String())
	}

	if _, err := os.Stat("layered-app/controllers/users_controller.go"); err != nil {
		t.Errorf("expected layered-app/controllers/users_controller.go to exist: %v", err)
	}
	if _, err := os.Stat("layered-app/config/settings.go"); err != nil {
		t.Errorf("expected layered-app/config/settings.go to exist: %v", err)
	}
	if _, err := os.Stat("layered-app/main.go"); err != nil {
		t.Errorf("expected layered-app/main.go to exist: %v", err)
	}
}
