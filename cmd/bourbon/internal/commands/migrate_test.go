package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateMissingRunner(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_migrate_empty_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"migrate"})
	if code == 0 {
		t.Errorf("expected migrate without runner to fail, got exit code 0")
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "no migration runner found") {
		t.Errorf("expected error message about missing runner, got: %s", combined)
	}
}

func TestMigrateCreate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_migrate_create_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"migrate", "create", "create_invoices"})
	if code != 0 {
		t.Fatalf("migrate create failed: %s", stderr.String())
	}

	migDir := filepath.Join(tempDir, "db", "migrations")
	files, err := os.ReadDir(migDir)
	if err != nil {
		t.Fatalf("failed reading migrations dir: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("expected migration file to be created, but directory is empty")
	}

	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_create_invoices.sql") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected migration file ending in _create_invoices.sql, found: %v", files)
	}
}

func TestMigrateWithRunner(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_migrate_runner_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Create a minimal go.mod and cmd/migrate/main.go
	_ = os.WriteFile("go.mod", []byte("module testmig\n\ngo 1.23\n"), 0644)
	migCmdDir := filepath.Join("cmd", "migrate")
	_ = os.MkdirAll(migCmdDir, 0755)

	runnerCode := `package main

import (
	"fmt"
	"os"
)

func main() {
	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}
	fmt.Printf("runner_executed:%s\n", action)
}
`
	_ = os.WriteFile(filepath.Join(migCmdDir, "main.go"), []byte(runnerCode), 0644)

	tests := []struct {
		args     []string
		expected string
	}{
		{args: []string{"migrate"}, expected: "runner_executed:up"},
		{args: []string{"migrate", "up"}, expected: "runner_executed:up"},
		{args: []string{"migrate", "rollback"}, expected: "runner_executed:rollback"},
		{args: []string{"migrate", "status"}, expected: "runner_executed:status"},
	}

	for _, tt := range tests {
		var stdout, stderr bytes.Buffer
		runner := NewRunner(&stdout, &stderr)

		code := runner.Run(tt.args)
		if code != 0 {
			t.Errorf("command %v failed with code %d: %s", tt.args, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), tt.expected) {
			t.Errorf("for args %v, expected output %q, got %q", tt.args, tt.expected, stdout.String())
		}
	}
}
