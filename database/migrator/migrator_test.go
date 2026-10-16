package migrator_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database/migrator"
	_ "modernc.org/sqlite"
)

func TestMigratorLifecycle(t *testing.T) {
	tempDir := t.TempDir()

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	defer db.Close()

	// 1. Create a migration file
	upSQL := "CREATE TABLE users (id VARCHAR(36) PRIMARY KEY, email VARCHAR(255) NOT NULL);"
	downSQL := "DROP TABLE users;"
	path, err := migrator.Create(tempDir, "create_users_table", upSQL, downSQL)
	if err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migration file was not written: %v", err)
	}

	m, err := migrator.New(db, tempDir, "sqlite3")
	if err != nil {
		t.Fatalf("failed to construct migrator: %v", err)
	}

	// 2. Status before migration
	statuses, err := m.Status()
	if err != nil {
		t.Fatalf("failed getting status: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Applied {
		t.Errorf("expected 1 unapplied migration, got %+v", statuses)
	}

	// 3. Migrate Up
	if err := m.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// 4. Verify table created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("table users query failed: %v", err)
	}

	// 5. Version check
	ver, err := m.Version()
	if err != nil || ver == 0 {
		t.Fatalf("expected version > 0, got %d, err: %v", ver, err)
	}

	// 6. Rollback
	if err := m.Rollback(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify table dropped
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err == nil {
		t.Errorf("expected error querying dropped users table, but got none")
	}
}

func TestCreateMigrationFile(t *testing.T) {
	tempDir := t.TempDir()
	path, err := migrator.Create(tempDir, "add-orders-table", "", "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if filepath.Base(path) == "" {
		t.Fatalf("invalid path returned: %s", path)
	}

	// Test optional empty name
	pathEmpty, err := migrator.Create(tempDir, "", "", "")
	if err != nil {
		t.Fatalf("Create with empty name failed: %v", err)
	}
	if !filepath.IsAbs(pathEmpty) && filepath.Base(pathEmpty) == "" {
		t.Fatalf("invalid path returned for empty name: %s", pathEmpty)
	}
}

type mockProvider struct {
	db     *sql.DB
	driver string
}

func (m *mockProvider) SQLDB() *sql.DB {
	return m.db
}

func (m *mockProvider) DriverName() string {
	return m.driver
}

func TestInCodeMigrateHelper(t *testing.T) {
	tempDir := t.TempDir()

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	defer db.Close()

	_, err = migrator.Create(tempDir, "create_items_table",
		"CREATE TABLE items (id VARCHAR(36) PRIMARY KEY, title VARCHAR(255) NOT NULL);",
		"DROP TABLE items;",
	)
	if err != nil {
		t.Fatalf("failed creating migration file: %v", err)
	}

	provider := &mockProvider{db: db, driver: "sqlite3"}

	// 1. Status via provider
	statuses, err := migrator.Status(provider, tempDir)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Applied {
		t.Errorf("expected 1 unapplied migration, got %+v", statuses)
	}

	// 2. Migrate via provider
	if err := migrator.Migrate(provider, tempDir); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// 3. Version via provider
	v, err := migrator.Version(provider, tempDir)
	if err != nil || v == 0 {
		t.Fatalf("expected version > 0, got %d, err: %v", v, err)
	}

	// 4. Rollback via provider
	if err := migrator.Rollback(provider, tempDir); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}
