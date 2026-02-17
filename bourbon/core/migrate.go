package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/bourbon/core/gormigrate"
)

// RunMigrations executes all pending migrations from registered apps
// This should be called from main.go after importing migration packages
func RunMigrations(app *Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if app.DB == nil {
		return fmt.Errorf("database not initialized - call ConnectDB() first")
	}

	// Initialize and run migrations
	if err := app.InitMigrations(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Get registered migrations
	appMigrations := gormigrate.GetAppMigrations()
	if len(appMigrations) == 0 {
		fmt.Println("WARNING: No migrations found!")
		return nil
	}

	// Check which migrations are already applied (from gormigrate's table)
	var appliedIDs []string
	app.DB.Table("bourbon_migrations").Pluck("id", &appliedIDs)

	appliedMap := make(map[string]bool)
	for _, id := range appliedIDs {
		appliedMap[id] = true
	}

	// Count pending migrations
	pendingCount := 0
	for _, m := range appMigrations {
		if !appliedMap[m.ID] {
			pendingCount++
		}
	}

	fmt.Printf("\nMigration System\n")
	fmt.Printf("═══════════════════════════════════════════\n")
	fmt.Printf("Found %d migration(s) (%d pending)\n\n", len(appMigrations), pendingCount)

	if pendingCount == 0 {
		fmt.Println("All migrations already applied!")
		return nil
	}

	// Run migrations via gormigrate (it handles tracking in bourbon_migrations)
	if err := app.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("\nMigrations completed successfully!")
	return nil
}

// ShowMigrationStatus displays the status of all migrations
func ShowMigrationStatus(app *Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if app.DB == nil {
		return fmt.Errorf("database not initialized - call ConnectDB() first")
	}

	// Get registered migrations grouped by app
	appMigrations := gormigrate.GetAppMigrations()
	if len(appMigrations) == 0 {
		fmt.Println("WARNING: No migrations found!")
		return nil
	}

	// Get applied migrations from gormigrate's table
	var appliedIDs []string
	app.DB.Table("bourbon_migrations").Pluck("id", &appliedIDs)

	// Create lookup map
	appliedMap := make(map[string]bool)
	for _, id := range appliedIDs {
		appliedMap[id] = true
	}

	// Group migrations by app
	groupedMigrations := make(map[string][]*gormigrate.AppMigration)
	for _, m := range appMigrations {
		groupedMigrations[m.AppName] = append(groupedMigrations[m.AppName], m)
	}

	// Calculate totals
	totalApplied := len(appliedIDs)
	totalPending := len(appMigrations) - totalApplied

	fmt.Printf("\nMigration Status\n")
	fmt.Printf("════════════════════════════════════════════\n")
	fmt.Printf("\nTotal migrations: %d\n", len(appMigrations))
	fmt.Printf("Applied: %d\n", totalApplied)
	fmt.Printf("Pending: %d\n\n", totalPending)

	// Show migrations grouped by app
	for appName, migrations := range groupedMigrations {
		fmt.Printf("\nApp: %s\n", appName)
		fmt.Println("────────────────────────────────────────────────────────────────")

		appApplied := 0
		for _, m := range migrations {
			if appliedMap[m.ID] {
				appApplied++
			}
		}

		fmt.Printf("  Migrations: %d total, %d applied, %d pending\n\n",
			len(migrations), appApplied, len(migrations)-appApplied)

		for i, m := range migrations {
			status := "PENDING"

			if appliedMap[m.ID] {
				status = "APPLIED"
			}

			fmt.Printf("  %2d. [%s] %s\n", i+1, status, m.ID)
		}
	}
	fmt.Println("\n────────────────────────────────────────────────────────────────")

	return nil
}

// RollbackLastMigration rolls back the last applied migration
func RollbackLastMigration(app *Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if app.DB == nil {
		return fmt.Errorf("database not initialized - call ConnectDB() first")
	}

	if err := app.InitMigrations(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := app.RollbackLast(); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback completed successfully")
	return nil
}

// MigrateToVersion migrates to a specific migration ID
func MigrateToVersion(app *Application, migrationID string) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if app.DB == nil {
		return fmt.Errorf("database not initialized - call ConnectDB() first")
	}

	if err := app.InitMigrations(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	fmt.Printf("Migrating to: %s...\n", migrationID)
	if err := app.MigrateTo(migrationID); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully")
	return nil
}

// RollbackToVersion rolls back to a specific migration ID
func RollbackToVersion(app *Application, migrationID string) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if app.DB == nil {
		return fmt.Errorf("database not initialized - call ConnectDB() first")
	}

	if err := app.InitMigrations(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	fmt.Printf("Rolling back to: %s...\n", migrationID)
	if err := app.RollbackTo(migrationID); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback completed successfully")
	return nil
}

// getProjectModule reads the go.mod file to get the module name
func getProjectModule() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module not found in go.mod")
}

// hasModels checks if an app has models defined
func hasModels(modelsPath string) bool {
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(modelsPath)
	if err != nil {
		return false
	}

	content := string(data)
	return strings.Contains(content, "gorm.Model") ||
		strings.Contains(content, "type") && strings.Contains(content, "struct")
}

// hasModelChanges checks if app has uncommitted model changes (simplified check)
func hasModelChanges(appName string) bool {
	modelsPath := filepath.Join("apps", appName, "models.go")
	migrationsDir := filepath.Join("apps", appName, "migrations")

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// If no migrations exist but models do, there are changes
		return hasModels(modelsPath)
	}

	// Check if migrations directory has any migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return false
	}

	hasMigrationFiles := false
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
			name := entry.Name()
			if name != ".gitkeep" && !strings.HasSuffix(name, "_test.go") {
				hasMigrationFiles = true
				break
			}
		}
	}

	// If models exist but no migration files, there are changes
	return hasModels(modelsPath) && !hasMigrationFiles
}
