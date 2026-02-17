package migration


import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Migration interface defines the contract for database migrations
// Django-style: includes App() to track which app the migration belongs to
type Migration interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
	Name() string
	Version() int
	App() string
}

// BaseMigration provides default implementations
type BaseMigration struct{}

func (m *BaseMigration) Up(db *gorm.DB) error {
	return nil
}

func (m *BaseMigration) Down(db *gorm.DB) error {
	return nil
}

func (m *BaseMigration) Name() string {
	return "base_migration"
}

func (m *BaseMigration) Version() int {
	return 0
}

func (m *BaseMigration) App() string {
	return "default"
}

// DjangoMigration represents the django_migrations table for tracking applied migrations
type DjangoMigration struct {
	ID      uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	App     string    `gorm:"type:varchar(255);not null" json:"app"`
	Name    string    `gorm:"type:varchar(255);not null" json:"name"`
	Applied time.Time `gorm:"autoCreateTime" json:"applied"`
}

// TableName sets the table name for Bourbon migrations
func (DjangoMigration) TableName() string {
	return "bourbon_migrations"
}

// MigrationRunner handles running and tracking migrations
type MigrationRunner struct {
	db *gorm.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *gorm.DB) *MigrationRunner {
	return &MigrationRunner{db: db}
}

// InitMigrationTable creates the django_migrations table if it doesn't exist
func (mr *MigrationRunner) InitMigrationTable() error {
	return mr.db.AutoMigrate(&DjangoMigration{})
}

// IsApplied checks if a migration has been applied
func (mr *MigrationRunner) IsApplied(app, name string) (bool, error) {
	var count int64
	err := mr.db.Model(&DjangoMigration{}).
		Where("app = ? AND name = ?", app, name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RecordMigration records a migration as applied
func (mr *MigrationRunner) RecordMigration(app, name string) error {
	migration := &DjangoMigration{
		App:  app,
		Name: name,
	}
	return mr.db.Create(migration).Error
}

// RemoveMigration removes a migration record (for rollback)
func (mr *MigrationRunner) RemoveMigration(app, name string) error {
	return mr.db.Where("app = ? AND name = ?", app, name).
		Delete(&DjangoMigration{}).Error
}

// GetAppliedMigrations returns all applied migrations
func (mr *MigrationRunner) GetAppliedMigrations() ([]DjangoMigration, error) {
	var migrations []DjangoMigration
	err := mr.db.Order("id ASC").Find(&migrations).Error
	return migrations, err
}

// GetAppMigrations returns all applied migrations for a specific app
func (mr *MigrationRunner) GetAppMigrations(app string) ([]DjangoMigration, error) {
	var migrations []DjangoMigration
	err := mr.db.Where("app = ?", app).Order("id ASC").Find(&migrations).Error
	return migrations, err
}

// ShowMigrationStatus displays migration status like Django's showmigrations
func (mr *MigrationRunner) ShowMigrationStatus() error {
	if err := mr.InitMigrationTable(); err != nil {
		return err
	}

	migrations, err := mr.GetAppliedMigrations()
	if err != nil {
		return err
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("=================")

	if len(migrations) == 0 {
		fmt.Println("No migrations applied yet.")
		return nil
	}

	// Group by app
	appMigrations := make(map[string][]DjangoMigration)
	for _, m := range migrations {
		appMigrations[m.App] = append(appMigrations[m.App], m)
	}

	for app, migs := range appMigrations {
		fmt.Printf("\n%s:\n", app)
		for _, m := range migs {
			fmt.Printf(" [X] %s (applied: %s)\n", m.Name, m.Applied.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

// RunMigration runs a single migration and records it
func (mr *MigrationRunner) RunMigration(migration Migration) error {
	app := migration.App()
	name := migration.Name()

	// Check if already applied
	applied, err := mr.IsApplied(app, name)
	if err != nil {
		return fmt.Errorf("error checking migration status: %w", err)
	}

	if applied {
		fmt.Printf("  [SKIP] %s.%s (already applied)\n", app, name)
		return nil
	}

	// Run the migration
	if err := migration.Up(mr.db); err != nil {
		return fmt.Errorf("error running migration %s.%s: %w", app, name, err)
	}

	// Record the migration
	if err := mr.RecordMigration(app, name); err != nil {
		return fmt.Errorf("error recording migration %s.%s: %w", app, name, err)
	}

	fmt.Printf("  [OK] %s.%s\n", app, name)
	return nil
}

// RollbackMigration rolls back a single migration
func (mr *MigrationRunner) RollbackMigration(migration Migration) error {
	app := migration.App()
	name := migration.Name()

	// Check if applied
	applied, err := mr.IsApplied(app, name)
	if err != nil {
		return fmt.Errorf("error checking migration status: %w", err)
	}

	if !applied {
		fmt.Printf("  [SKIP] %s.%s (not applied)\n", app, name)
		return nil
	}

	// Run the rollback
	if err := migration.Down(mr.db); err != nil {
		return fmt.Errorf("error rolling back migration %s.%s: %w", app, name, err)
	}

	// Remove the migration record
	if err := mr.RemoveMigration(app, name); err != nil {
		return fmt.Errorf("error removing migration record %s.%s: %w", app, name, err)
	}

	fmt.Printf("  [ROLLBACK] %s.%s\n", app, name)
	return nil
}
