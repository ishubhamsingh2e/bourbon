package memdb

import (
	"context"
	"errors"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database"
)

func TestMemDBBasicOperations(t *testing.T) {
	ctx := context.Background()
	db := New()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	res, err := db.Exec(ctx, "insert_user", "Alice")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Errorf("expected 1 row affected, got %d", n)
	}

	val, ok := db.GetValue("insert_user")
	if !ok || val != "Alice" {
		t.Errorf("GetValue = %v, %v, want Alice, true", val, ok)
	}
}

func TestMemDBTransactions(t *testing.T) {
	ctx := context.Background()
	db := New()

	// 1. Transaction Commit
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	_, _ = tx.Exec(ctx, "tx_key", "committed_value")
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit failed: %v", err)
	}

	val, ok := db.GetValue("tx_key")
	if !ok || val != "committed_value" {
		t.Errorf("committed value not found in db: %v", val)
	}

	// 2. Transaction Rollback
	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	_, _ = tx2.Exec(ctx, "rollback_key", "staged_value")
	if err := tx2.Rollback(); err != nil {
		t.Fatalf("tx2.Rollback failed: %v", err)
	}

	if _, exists := db.GetValue("rollback_key"); exists {
		t.Errorf("rolled back key should not exist in db")
	}

	// 3. Double Commit/Rollback error
	if err := tx2.Commit(); !errors.Is(err, database.ErrTxClosed) {
		t.Errorf("expected ErrTxClosed on closed tx, got %v", err)
	}
}

func TestMemDBReplicasAndNamedConnections(t *testing.T) {
	db := New()
	replicaDB := New()
	analyticsDB := New()

	db.SetReplica(replicaDB)
	db.SetConnection("analytics", analyticsDB)

	if db.Replica() != replicaDB {
		t.Errorf("Replica() did not return configured replica")
	}

	conn, err := db.Connection("analytics")
	if err != nil || conn != analyticsDB {
		t.Errorf("Connection(analytics) failed: %v", err)
	}

	_, err = db.Connection("missing_db")
	if !errors.Is(err, database.ErrConnectionNotFound) {
		t.Errorf("expected ErrConnectionNotFound, got %v", err)
	}
}
