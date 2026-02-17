package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateMigration creates a new migration file in the default app
func GenerateMigration(name string) error {
	// Find the default app (first app in apps/ directory)
	appName, err := getDefaultApp()
	if err != nil {
		return err
	}

	return GenerateMigrationForApp(appName, name)
}

// GenerateMigrationForApp creates a new migration file for a specific app
func GenerateMigrationForApp(appName, name string) error {
	// Scan models to detect changes
	models, err := ScanModels(appName)
	if err != nil {
		return fmt.Errorf("failed to scan models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models found in apps/%s/models.go - create models first", appName)
	}

	// Detect all changes
	changes, err := DetectAllChanges(appName, models)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	if !changes.HasChanges() {
		fmt.Println("No changes detected - models are up to date")
		return nil
	}

	// Show all destructive changes and ask for confirmation
	if changes.HasDestructiveChanges() {
		fmt.Println("\nWARNING: Destructive changes detected!")
		
		if len(changes.DeletedModels) > 0 {
			fmt.Println("\nModels to be DELETED:")
			for _, modelName := range changes.DeletedModels {
				fmt.Printf("  - %s (table: %s)\n", modelName, toSnakeCase(modelName))
			}
		}
		
		if len(changes.DeletedFields) > 0 {
			fmt.Println("\nFields to be DELETED:")
			for modelName, fields := range changes.DeletedFields {
				fmt.Printf("  Model: %s\n", modelName)
				for _, field := range fields {
					fmt.Printf("    - %s %s\n", field.Name, field.Type)
				}
			}
		}
		
		fmt.Println("\nThese changes CANNOT be undone!")
		fmt.Print("\nContinue? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "y" {
			fmt.Println("Migration cancelled.")
			return nil
		}
	}

	// Create migrations directory if it doesn't exist
	migrationsDir := filepath.Join("apps", appName, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")
	
	// Generate filename and migration ID
	var fileName, migrationID string
	if name == "" {
		// Use only timestamp if no name provided
		fileName = fmt.Sprintf("%s.go", timestamp)
		migrationID = timestamp
	} else {
		// Use timestamp + name
		cleanName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
		fileName = fmt.Sprintf("%s_%s.go", timestamp, cleanName)
		migrationID = fmt.Sprintf("%s_%s", timestamp, cleanName)
	}
	
	filePath := filepath.Join(migrationsDir, fileName)

	// Generate migration code following gormigrate best practices
	migrateCode := GenerateMigrationCodeFromChanges(changes)
	rollbackCode := GenerateRollbackCodeFromChanges(changes)

	// Check if we need time import (only for CreateTable with BaseModel fields)
	needsTimeImport := len(changes.NewModels) > 0
	timeImport := ""
	if needsTimeImport {
		timeImport = "\t\"time\"\n\n"
	}

	// Migration template following gormigrate best practices
	template := fmt.Sprintf(`package migrations

import (
%s	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	"gorm.io/gorm"
)

func init() {
	core.RegisterGormigrateMigration(&gormigrate.Migration{
		ID: "%s",
		Migrate: func(tx *gorm.DB) error {
%s
		},
		Rollback: func(tx *gorm.DB) error {
%s
		},
	})
}
`, timeImport, migrationID, migrateCode, rollbackCode)

	// Write file
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	// Update migration state
	if err := UpdateMigrationState(appName, models, migrationID); err != nil {
		return fmt.Errorf("failed to update migration state: %w", err)
	}

	fmt.Printf("Created migration: %s\n", filePath)
	fmt.Printf("  Models: %s\n", getModelNames(models))
	return nil
}

// getModelNames returns a comma-separated list of model names
func getModelNames(models []ModelInfo) string {
	names := make([]string, len(models))
	for i, model := range models {
		names[i] = model.Name
	}
	return strings.Join(names, ", ")
}

// getDefaultApp returns the first app found in apps/ directory
func getDefaultApp() (string, error) {
	appsDir := "apps"
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return "", fmt.Errorf("apps directory not found. Are you in the project root?")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no apps found in apps/ directory")
}

