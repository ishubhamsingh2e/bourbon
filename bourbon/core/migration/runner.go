package migration

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MigrationRegistry holds all registered migrations
type MigrationRegistry struct {
	migrations map[string][]Migration // app -> migrations
	mu         sync.RWMutex
}

var registry = &MigrationRegistry{
	migrations: make(map[string][]Migration),
}

// RegisterMigration registers a migration for an app
func RegisterMigration(migration Migration) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	app := migration.App()
	registry.migrations[app] = append(registry.migrations[app], migration)
}

// GetMigrations returns all migrations for an app, sorted by version
func GetMigrations(app string) []Migration {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	migrations := registry.migrations[app]
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version() < sorted[j].Version()
	})

	return sorted
}

// GetAllApps returns all apps that have registered migrations
func GetAllApps() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	apps := make([]string, 0, len(registry.migrations))
	for app := range registry.migrations {
		apps = append(apps, app)
	}
	sort.Strings(apps)
	return apps
}

// RunRegisteredMigrations runs all registered migrations for an app
func (mr *MigrationRunner) RunRegisteredMigrations(app string) (int, error) {
	migrations := GetMigrations(app)
	if len(migrations) == 0 {
		return 0, nil
	}

	count := 0
	for _, migration := range migrations {
		if err := mr.RunMigration(migration); err != nil {
			return count, err
		}

		// Check if it was actually applied (not skipped)
		applied, _ := mr.IsApplied(migration.App(), migration.Name())
		if applied {
			count++
		}
	}

	return count, nil
}

// RunAllRegisteredMigrations runs migrations for all registered apps
func (mr *MigrationRunner) RunAllRegisteredMigrations() (int, error) {
	apps := GetAllApps()
	totalCount := 0

	for _, app := range apps {
		fmt.Printf("\n%s:\n", app)
		count, err := mr.RunRegisteredMigrations(app)
		if err != nil {
			return totalCount, fmt.Errorf("error migrating %s: %w", app, err)
		}
		totalCount += count
	}

	return totalCount, nil
}

// MigrationRecord represents a migration entry in the database
type MigrationRecord struct {
	ID        string    `gorm:"primaryKey;size:255"`
	AppName   string    `gorm:"size:100;index"`
	AppliedAt time.Time `gorm:"index"`
}

// TableName specifies the table name for migration records
func (MigrationRecord) TableName() string {
	return "bourbon_migrations"
}

// MigrationTracker tracks migrations per application with timestamps
type MigrationTracker struct {
	db *gorm.DB
}

// NewMigrationTracker creates a new migration tracker
func NewMigrationTracker(db *gorm.DB) *MigrationTracker {
	return &MigrationTracker{db: db}
}

// EnsureTable creates the migrations table if it doesn't exist
func (mt *MigrationTracker) EnsureTable() error {
	return mt.db.AutoMigrate(&MigrationRecord{})
}

// RecordMigration records a migration as applied
func (mt *MigrationTracker) RecordMigration(id, appName string) error {
	record := MigrationRecord{
		ID:        id,
		AppName:   appName,
		AppliedAt: time.Now(),
	}
	return mt.db.Create(&record).Error
}

// RemoveMigration removes a migration record (for rollback)
func (mt *MigrationTracker) RemoveMigration(id string) error {
	return mt.db.Where("id = ?", id).Delete(&MigrationRecord{}).Error
}

// IsMigrationApplied checks if a migration has been applied
func (mt *MigrationTracker) IsMigrationApplied(id string) (bool, error) {
	var count int64
	err := mt.db.Model(&MigrationRecord{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// GetAppliedMigrations returns all applied migrations
func (mt *MigrationTracker) GetAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	err := mt.db.Order("applied_at ASC").Find(&records).Error
	return records, err
}

// GetAppliedMigrationsByApp returns applied migrations for a specific app
func (mt *MigrationTracker) GetAppliedMigrationsByApp(appName string) ([]MigrationRecord, error) {
	var records []MigrationRecord
	err := mt.db.Where("app_name = ?", appName).Order("applied_at ASC").Find(&records).Error
	return records, err
}

// GetAppliedMigrationIDs returns just the IDs of applied migrations
func (mt *MigrationTracker) GetAppliedMigrationIDs() ([]string, error) {
	var ids []string
	err := mt.db.Model(&MigrationRecord{}).Pluck("id", &ids).Error
	return ids, err
}

// GetMigrationsByApp groups migrations by app name
func (mt *MigrationTracker) GetMigrationsByApp() (map[string][]MigrationRecord, error) {
	var records []MigrationRecord
	err := mt.db.Order("app_name, applied_at ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]MigrationRecord)
	for _, record := range records {
		grouped[record.AppName] = append(grouped[record.AppName], record)
	}

	return grouped, nil
}
