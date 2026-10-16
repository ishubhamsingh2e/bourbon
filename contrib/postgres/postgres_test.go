package postgres

import (
	"database/sql"
	"testing"
)

func TestPostgresAdapterStructure(t *testing.T) {
	fakeDB := &sql.DB{}
	db := New(fakeDB)

	if db.primary != fakeDB {
		t.Errorf("primary mismatch")
	}

	replica := db.Replica()
	if replica == nil {
		t.Errorf("expected non-nil replica fallback")
	}

	_, err := db.Connection("unknown")
	if err == nil {
		t.Errorf("expected error for unknown connection")
	}
}
