package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/database"
)

func TestSQLiteAdapterStructure(t *testing.T) {
	fakeDB := &sql.DB{}
	db := New(fakeDB)

	if db.db != fakeDB {
		t.Errorf("db pointer mismatch")
	}

	replica := db.Replica()
	if replica == nil {
		t.Errorf("expected non-nil replica")
	}
}

func TestSQLiteProviderLifecycle(t *testing.T) {
	provider := NewProvider(Config{
		DSN: "file::memory:?cache=shared",
	})

	application := app.New()
	application.Use(provider)

	// Test registration
	c := application.Container()
	if err := provider.Register(c); err != nil {
		t.Fatalf("failed registering sqlite provider: %v", err)
	}

	// Resolve DB
	var db database.DB
	if err := c.Resolve(&db); err != nil {
		t.Fatalf("failed resolving database.DB: %v", err)
	}

	// Test boot
	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting sqlite provider: %v", err)
	}

	// Test shutdown
	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down sqlite provider: %v", err)
	}
}
