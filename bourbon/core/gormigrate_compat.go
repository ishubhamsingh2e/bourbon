package core

import (
	"github.com/go-gormigrate/gormigrate/v2"
	gormigratePackage "github.com/ishubhamsingh2e/bourbon/bourbon/core/gormigrate"
)

// Re-export gormigrate types and functions for backward compatibility
// These maintain the old API while delegating to the new package structure

// AppMigration is re-exported from gormigrate package
type AppMigration = gormigratePackage.AppMigration

// RegisterGormigrateMigration registers a migration in the global registry
// This function is re-exported for backward compatibility
func RegisterGormigrateMigration(migration *gormigrate.Migration) {
	gormigratePackage.RegisterGormigrateMigration(migration)
}

// RegisterAppMigration registers a migration with explicit app name
// This function is re-exported for backward compatibility
func RegisterAppMigration(appName string, migration *gormigrate.Migration) {
	gormigratePackage.RegisterAppMigration(appName, migration)
}

// RegisterGormigrateMigrations registers multiple migrations at once
// This function is re-exported for backward compatibility
func RegisterGormigrateMigrations(migrations []*gormigrate.Migration) {
	gormigratePackage.RegisterGormigrateMigrations(migrations)
}

// GetGormigrateMigrations returns all registered migrations
// This function is re-exported for backward compatibility
func GetGormigrateMigrations() []*gormigrate.Migration {
	return gormigratePackage.GetGormigrateMigrations()
}

// GetAppMigrations returns all registered migrations with app metadata
// This function is re-exported for backward compatibility
func GetAppMigrations() []*AppMigration {
	return gormigratePackage.GetAppMigrations()
}

// GetMigrationsByApp returns migrations grouped by app name
// This function is re-exported for backward compatibility
func GetMigrationsByApp() map[string][]*AppMigration {
	return gormigratePackage.GetMigrationsByApp()
}

// ClearGormigrateMigrations clears all registered migrations
// This function is re-exported for backward compatibility
func ClearGormigrateMigrations() {
	gormigratePackage.ClearGormigrateMigrations()
}
