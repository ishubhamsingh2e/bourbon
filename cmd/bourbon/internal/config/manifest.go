// Package config provides project manifest (bourbon.yaml) parsing and conventions.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfig describes project metadata.
type ProjectConfig struct {
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	Structure string `json:"structure"`
	GoModule  string `json:"go_module"`
}

// PathsConfig describes project file layout paths.
type PathsConfig struct {
	SourceDir     string `json:"source_dir"`
	MigrationsDir string `json:"migrations_dir"`
}

// Manifest represents the minimal bourbon.yaml configuration file.
type Manifest struct {
	Version string        `json:"version"`
	Project ProjectConfig `json:"project"`
	Paths   PathsConfig   `json:"paths"`
}

// DefaultManifest constructs a Manifest with convention-over-configuration defaults.
func DefaultManifest(name, module, profile, structure string) *Manifest {
	if profile == "" {
		profile = "fullstack"
	}
	if structure == "" {
		structure = "feature"
	}
	if module == "" {
		module = name
	}

	srcDir := "internal/modules"
	if structure == "layered" {
		srcDir = "controllers"
	}

	return &Manifest{
		Version: "1",
		Project: ProjectConfig{
			Name:      name,
			Profile:   profile,
			Structure: structure,
			GoModule:  module,
		},
		Paths: PathsConfig{
			SourceDir:     srcDir,
			MigrationsDir: "db/migrations",
		},
	}
}

// ParseSimpleYAML parses key-value YAML format without external dependencies.
func ParseSimpleYAML(data []byte) (*Manifest, error) {
	m := DefaultManifest("unnamed", "unnamed", "fullstack", "feature")

	lines := strings.Split(string(data), "\n")
	var section string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if strings.HasSuffix(trimmed, ":") {
			if indent == 0 {
				section = strings.TrimSuffix(trimmed, ":")
			}
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		if indent == 0 && key == "version" {
			m.Version = val
			continue
		}

		switch section {
		case "project":
			switch key {
			case "name":
				m.Project.Name = val
			case "profile", "preset":
				m.Project.Profile = val
			case "structure":
				m.Project.Structure = val
			case "go_module":
				m.Project.GoModule = val
			}
		case "paths":
			switch key {
			case "source_dir":
				m.Paths.SourceDir = val
			case "migrations_dir":
				m.Paths.MigrationsDir = val
			}
		}
	}

	if m.Project.Name == "" || m.Project.Name == "unnamed" {
		return nil, errors.New("config: manifest missing project name")
	}

	return m, nil
}

// LoadManifest reads and parses a bourbon.yaml manifest file.
func LoadManifest(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("config: failed reading manifest '%s': %w", filePath, err)
	}

	return ParseSimpleYAML(data)
}

// FindProjectRoot locates the nearest Bourbon project directory containing bourbon.yaml or project indicators.
func FindProjectRoot(startDir string) (string, *Manifest, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", nil, err
	}

	for {
		manifestPath := filepath.Join(dir, "bourbon.yaml")
		if manifest, err := LoadManifest(manifestPath); err == nil {
			return dir, manifest, nil
		}

		// Also check for standard Bourbon project signatures (go.mod with config/settings.go)
		if _, err := os.Stat(filepath.Join(dir, "config", "settings.go")); err == nil {
			modName := "app"
			if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "module ") {
						modName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
						break
					}
				}
			}
			structType := "feature"
			if _, err := os.Stat(filepath.Join(dir, "controllers")); err == nil {
				structType = "layered"
			}
			return dir, DefaultManifest(filepath.Base(dir), modName, "fullstack", structType), nil
		}
		if _, err := os.Stat(filepath.Join(dir, "internal", "config", "settings.go")); err == nil {
			modName := "app"
			if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "module ") {
						modName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
						break
					}
				}
			}
			structType := "feature"
			if _, err := os.Stat(filepath.Join(dir, "controllers")); err == nil {
				structType = "layered"
			}
			return dir, DefaultManifest(filepath.Base(dir), modName, "fullstack", structType), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil, errors.New("not in a Bourbon project directory (bourbon.yaml not found)\nRun 'bourbon new <name>' to create a project, or 'cd' into an existing Bourbon project")
}

// FormatYAML serializes the manifest into a clean, minimal bourbon.yaml format.
func (m *Manifest) FormatYAML() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("version: \"%s\"\n\n", m.Version))

	sb.WriteString("project:\n")
	sb.WriteString(fmt.Sprintf("  name: \"%s\"\n", m.Project.Name))
	sb.WriteString(fmt.Sprintf("  profile: \"%s\"\n", m.Project.Profile))
	sb.WriteString(fmt.Sprintf("  structure: \"%s\"\n", m.Project.Structure))
	sb.WriteString(fmt.Sprintf("  go_module: \"%s\"\n\n", m.Project.GoModule))

	sb.WriteString("paths:\n")
	sb.WriteString(fmt.Sprintf("  source_dir: \"%s\"\n", m.Paths.SourceDir))
	sb.WriteString(fmt.Sprintf("  migrations_dir: \"%s\"\n", m.Paths.MigrationsDir))

	return sb.String()
}
