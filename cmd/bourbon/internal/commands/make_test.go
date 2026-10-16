package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMakeOutsideProjectFails(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_empty_dir_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"make", "app", "orders"})
	if code == 0 {
		t.Errorf("expected make outside project to fail, got 0")
	}
}

func TestMakeAppGeneratorFeature(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_make_app_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	manifest := `version: "1"
project:
  name: "feature-app"
  profile: "fullstack"
  structure: "feature"
  go_module: "feature-app"

paths:
  source_dir: "apps"
  migrations_dir: "db/migrations"
`
	_ = os.WriteFile("bourbon.yaml", []byte(manifest), 0644)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"make", "app", "orders"})
	if code != 0 {
		t.Fatalf("make app orders failed: %s", stderr.String())
	}

	expectedFiles := []string{
		"apps/orders/controller.go",
		"apps/orders/service.go",
		"apps/orders/repository.go",
		"apps/orders/model.go",
		"apps/orders/routes.go",
		"apps/orders/orders_test.go",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s missing: %v", f, err)
		}
	}
}

func TestMakeGeneratorsFeatureLayout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_make_feature_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	manifest := `version: "1"
project:
  name: "feature-app"
  profile: "fullstack"
  structure: "feature"
  go_module: "feature-app"

paths:
  source_dir: "apps"
  migrations_dir: "db/migrations"
`
	_ = os.WriteFile("bourbon.yaml", []byte(manifest), 0644)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	generators := []struct {
		subcmd   string
		name     string
		expected string
	}{
		{"controller", "billing", "apps/billing/controller.go"},
		{"service", "billing", "apps/billing/service.go"},
		{"repository", "billing", "apps/billing/repository.go"},
		{"model", "billing", "apps/billing/model.go"},
		{"test", "billing", "apps/billing/billing_test.go"},
		{"middleware", "auth_guard", "middleware/auth_guard.go"},
		{"job", "send_report", "jobs/send_report_job.go"},
	}

	for _, g := range generators {
		stdout.Reset()
		stderr.Reset()

		code := runner.Run([]string{"make", g.subcmd, g.name})
		if code != 0 {
			t.Errorf("make %s failed with code %d, stderr: %s", g.subcmd, code, stderr.String())
			continue
		}

		if _, err := os.Stat(g.expected); err != nil {
			t.Errorf("expected file %s does not exist: %v", g.expected, err)
		}
	}

	// Test migration generator
	stdout.Reset()
	code := runner.Run([]string{"make", "migration", "create_orders_table"})
	if code != 0 {
		t.Errorf("make migration failed: %s", stderr.String())
	}
	files, _ := filepath.Glob("db/migrations/*_create_orders_table.sql")
	if len(files) != 1 {
		t.Errorf("expected 1 migration file, found %d", len(files))
	}

	// Test make migration without explicit name
	stdout.Reset()
	stderr.Reset()
	code = runner.Run([]string{"make", "migration"})
	if code != 0 {
		t.Fatalf("make migration without name failed: %s", stderr.String())
	}
	defaultFiles, _ := filepath.Glob("db/migrations/*_migration.sql")
	if len(defaultFiles) != 1 {
		t.Errorf("expected 1 default migration file, found %d", len(defaultFiles))
	}
}

func TestMakeGeneratorsLayeredLayout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_make_layered_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	manifest := `version: "1"
project:
  name: "layered-app"
  profile: "fullstack"
  structure: "layered"
  go_module: "layered-app"

paths:
  source_dir: "controllers"
  migrations_dir: "db/migrations"
`
	_ = os.WriteFile("bourbon.yaml", []byte(manifest), 0644)

	var stdout, stderr bytes.Buffer
	runner := NewRunner(&stdout, &stderr)

	code := runner.Run([]string{"make", "controller", "product"})
	if code != 0 {
		t.Fatalf("make controller failed: %s", stderr.String())
	}

	if _, err := os.Stat("controllers/product_controller.go"); err != nil {
		t.Errorf("expected controllers/product_controller.go to exist: %v", err)
	}

	// Test overwrite refusal without --force
	stdout.Reset()
	stderr.Reset()
	code = runner.Run([]string{"make", "controller", "product"})
	if code == 0 {
		t.Errorf("expected make controller without --force to fail on existing file, got 0")
	}

	// Test overwrite success with --force
	stdout.Reset()
	stderr.Reset()
	code = runner.Run([]string{"make", "controller", "product", "--force"})
	if code != 0 {
		t.Errorf("expected make controller with --force to succeed, got %d (stderr: %s)", code, stderr.String())
	}
}
