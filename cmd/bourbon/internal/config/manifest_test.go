package config

import (
	"testing"
)

func TestManifestRoundtrip(t *testing.T) {
	manifest := DefaultManifest("shop-api", "github.com/example/shop-api", "api", "feature")

	yamlStr := manifest.FormatYAML()

	parsed, err := ParseSimpleYAML([]byte(yamlStr))
	if err != nil {
		t.Fatalf("failed parsing generated YAML: %v", err)
	}

	if parsed.Project.Name != "shop-api" {
		t.Errorf("expected name 'shop-api', got: %s", parsed.Project.Name)
	}
	if parsed.Project.Profile != "api" {
		t.Errorf("expected profile 'api', got: %s", parsed.Project.Profile)
	}
	if parsed.Project.Structure != "feature" {
		t.Errorf("expected structure 'feature', got: %s", parsed.Project.Structure)
	}
	if parsed.Paths.SourceDir != "internal/modules" {
		t.Errorf("expected source_dir 'internal/modules', got: %s", parsed.Paths.SourceDir)
	}
}
