package commands

import (
	"bytes"
	"os"
	"testing"
)

func TestDevProxyBuffering(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_dev_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	manifest := `version: "1"
project:
  name: "dev-app"
  profile: "fullstack"
  structure: "feature"
  go_module: "dev-app"

paths:
  source_dir: "apps"
  migrations_dir: "db/migrations"
`
	_ = os.WriteFile("bourbon.yaml", []byte(manifest), 0644)

	proxy := NewDevProxy(8080)
	proxy.SetReady(true)

	// Run dev --dry-run
	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"dev", "--dry-run"})
	if code != 0 {
		t.Errorf("dev --dry-run failed with code %d, stderr: %s", code, stderr.String())
	}
}
