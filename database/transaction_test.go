package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type MockTx struct {
	committed  bool
	rolledBack bool
	closed     bool
}

func (m *MockTx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *MockTx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *MockTx) Commit() error {
	if m.closed {
		return ErrTxClosed
	}
	m.committed = true
	m.closed = true
	return nil
}

func (m *MockTx) Rollback() error {
	if m.closed {
		return ErrTxClosed
	}
	m.rolledBack = true
	m.closed = true
	return nil
}

type MockDB struct {
	tx *MockTx
}

func (m *MockDB) Ping(ctx context.Context) error { return nil }
func (m *MockDB) Close() error                   { return nil }
func (m *MockDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *MockDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	m.tx = &MockTx{}
	return m.tx, nil
}
func (m *MockDB) Replica() DB { return m }
func (m *MockDB) Connection(name string) (DB, error) {
	return m, nil
}

func TestWithTxSuccess(t *testing.T) {
	db := &MockDB{}

	err := WithTx(context.Background(), db, func(ctx context.Context, tx Tx) error {
		return nil
	})

	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}
	if !db.tx.committed {
		t.Errorf("expected tx to be committed")
	}
	if db.tx.rolledBack {
		t.Errorf("tx should not be rolled back")
	}
}

func TestWithTxErrorRollback(t *testing.T) {
	db := &MockDB{}
	expectedErr := errors.New("business error")

	err := WithTx(context.Background(), db, func(ctx context.Context, tx Tx) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if db.tx.committed {
		t.Errorf("tx should not be committed")
	}
	if !db.tx.rolledBack {
		t.Errorf("expected tx to be rolled back")
	}
}

func TestWithTxPanicRollback(t *testing.T) {
	db := &MockDB{}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic to propagate, got nil")
		}
		if db.tx.committed {
			t.Errorf("tx should not be committed on panic")
		}
		if !db.tx.rolledBack {
			t.Errorf("expected tx to be rolled back on panic")
		}
	}()

	_ = WithTx(context.Background(), db, func(ctx context.Context, tx Tx) error {
		panic("database hardware fault")
	})
}
