package gormigrate

import (
	"sync"

	"github.com/go-gormigrate/gormigrate/v2"
)

// AppMigration wraps a gormigrate migration with app metadata
type AppMigration struct {
	*gormigrate.Migration
	AppName string
}

// GormigrateMigrationRegistry holds all registered gormigrate migrations
type GormigrateMigrationRegistry struct {
	migrations []*AppMigration
	mu         sync.RWMutex
}

var gormigrateRegistry = &GormigrateMigrationRegistry{
	migrations: make([]*AppMigration, 0),
}

// RegisterGormigrateMigration registers a migration in the global registry
func RegisterGormigrateMigration(migration *gormigrate.Migration) {
	// Extract app name from migration ID (format: timestamp_name or app/timestamp_name)
	appName := "default"
	gormigrateRegistry.mu.Lock()
	defer gormigrateRegistry.mu.Unlock()
	gormigrateRegistry.migrations = append(gormigrateRegistry.migrations, &AppMigration{
		Migration: migration,
		AppName:   appName,
	})
}

// RegisterAppMigration registers a migration with explicit app name
func RegisterAppMigration(appName string, migration *gormigrate.Migration) {
	gormigrateRegistry.mu.Lock()
	defer gormigrateRegistry.mu.Unlock()
	gormigrateRegistry.migrations = append(gormigrateRegistry.migrations, &AppMigration{
		Migration: migration,
		AppName:   appName,
	})
}

// RegisterGormigrateMigrations registers multiple migrations at once
func RegisterGormigrateMigrations(migrations []*gormigrate.Migration) {
	gormigrateRegistry.mu.Lock()
	defer gormigrateRegistry.mu.Unlock()
	for _, m := range migrations {
		gormigrateRegistry.migrations = append(gormigrateRegistry.migrations, &AppMigration{
			Migration: m,
			AppName:   "default",
		})
	}
}

// GetGormigrateMigrations returns all registered migrations
func GetGormigrateMigrations() []*gormigrate.Migration {
	gormigrateRegistry.mu.RLock()
	defer gormigrateRegistry.mu.RUnlock()

	// Return just the gormigrate.Migration part
	result := make([]*gormigrate.Migration, len(gormigrateRegistry.migrations))
	for i, m := range gormigrateRegistry.migrations {
		result[i] = m.Migration
	}
	return result
}

// GetAppMigrations returns all registered migrations with app metadata
func GetAppMigrations() []*AppMigration {
	gormigrateRegistry.mu.RLock()
	defer gormigrateRegistry.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*AppMigration, len(gormigrateRegistry.migrations))
	copy(result, gormigrateRegistry.migrations)
	return result
}

// GetMigrationsByApp returns migrations grouped by app name
func GetMigrationsByApp() map[string][]*AppMigration {
	gormigrateRegistry.mu.RLock()
	defer gormigrateRegistry.mu.RUnlock()

	grouped := make(map[string][]*AppMigration)
	for _, m := range gormigrateRegistry.migrations {
		grouped[m.AppName] = append(grouped[m.AppName], m)
	}
	return grouped
}

// ClearGormigrateMigrations clears all registered migrations (useful for testing)
func ClearGormigrateMigrations() {
	gormigrateRegistry.mu.Lock()
	defer gormigrateRegistry.mu.Unlock()
	gormigrateRegistry.migrations = make([]*AppMigration, 0)
}
