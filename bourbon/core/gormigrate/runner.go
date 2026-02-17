package gormigrate

import (
	"fmt"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core/migration"
	"gorm.io/gorm"
)

// GormigrateRunner wraps gormigrate for managing migrations
type GormigrateRunner struct {
	db         *gorm.DB
	migrator   *gormigrate.Gormigrate
	migrations []*gormigrate.Migration
	tracker    *migration.MigrationTracker
}

// NewGormigrateRunner creates a new gormigrate-based migration runner
func NewGormigrateRunner(db *gorm.DB) *GormigrateRunner {
	return &GormigrateRunner{
		db:         db,
		migrations: make([]*gormigrate.Migration, 0),
		tracker:    migration.NewMigrationTracker(db),
	}
}

// AddMigration adds a migration to the runner
func (gr *GormigrateRunner) AddMigration(id string, migrate gormigrate.MigrateFunc, rollback gormigrate.RollbackFunc) {
	gr.migrations = append(gr.migrations, &gormigrate.Migration{
		ID:       id,
		Migrate:  migrate,
		Rollback: rollback,
	})
}

// AddMigrations adds multiple migrations at once
func (gr *GormigrateRunner) AddMigrations(migrations []*gormigrate.Migration) {
	gr.migrations = append(gr.migrations, migrations...)
}

// Initialize creates the gormigrate instance with all registered migrations
func (gr *GormigrateRunner) Initialize() error {
	if len(gr.migrations) == 0 {
		return fmt.Errorf("no migrations registered")
	}

	// Configure gormigrate to use bourbon_migrations table
	options := gormigrate.DefaultOptions
	options.TableName = "bourbon_migrations"

	gr.migrator = gormigrate.New(gr.db, options, gr.migrations)
	return nil
}

// Migrate runs all pending migrations
func (gr *GormigrateRunner) Migrate() error {
	if gr.migrator == nil {
		if err := gr.Initialize(); err != nil {
			return err
		}
	}

	log.Println("Running migrations...")
	if err := gr.migrator.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// RollbackLast rolls back the last migration
func (gr *GormigrateRunner) RollbackLast() error {
	if gr.migrator == nil {
		if err := gr.Initialize(); err != nil {
			return err
		}
	}

	log.Println("Rolling back last migration...")
	if err := gr.migrator.RollbackLast(); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Println("Rollback completed successfully")
	return nil
}

// RollbackTo rolls back to a specific migration ID
func (gr *GormigrateRunner) RollbackTo(migrationID string) error {
	if gr.migrator == nil {
		if err := gr.Initialize(); err != nil {
			return err
		}
	}

	log.Printf("Rolling back to migration: %s...\n", migrationID)
	if err := gr.migrator.RollbackTo(migrationID); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Println("Rollback completed successfully")
	return nil
}

// MigrateTo migrates to a specific migration ID
func (gr *GormigrateRunner) MigrateTo(migrationID string) error {
	if gr.migrator == nil {
		if err := gr.Initialize(); err != nil {
			return err
		}
	}

	log.Printf("Migrating to: %s...\n", migrationID)
	if err := gr.migrator.MigrateTo(migrationID); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migration completed successfully")
	return nil
}

// GetMigrations returns all registered migrations
func (gr *GormigrateRunner) GetMigrations() []*gormigrate.Migration {
	return gr.migrations
}

// GetTracker returns the migration tracker
func (gr *GormigrateRunner) GetTracker() *migration.MigrationTracker {
	return gr.tracker
}

// InitSchema can be used to initialize the database schema from scratch
// This is optional and useful for first-time setups
func (gr *GormigrateRunner) InitSchema(initFunc gormigrate.InitSchemaFunc) {
	if gr.migrator != nil {
		gr.migrator.InitSchema(initFunc)
	}
}
