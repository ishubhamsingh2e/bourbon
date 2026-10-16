// Package memdb provides an in-core, zero-dependency in-memory database reference adapter.
package memdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/database"
)

type memResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r *memResult) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

func (r *memResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// MemDB implements database.DB entirely in memory.
type MemDB struct {
	mu          sync.RWMutex
	closed      bool
	data        map[string]any
	connections map[string]database.DB
	replica     database.DB
}

// New constructs an initialized MemDB instance.
func New() *MemDB {
	db := &MemDB{
		data:        make(map[string]any),
		connections: make(map[string]database.DB),
	}
	db.replica = db
	return db
}

// Ping verifies that the memory database is active.
func (m *MemDB) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return fmt.Errorf("memdb: database is closed")
	}
	return nil
}

// Close marks the in-memory database as closed.
func (m *MemDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.data = make(map[string]any)
	return nil
}

// Exec simulates query execution in memory.
func (m *MemDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, fmt.Errorf("memdb: database is closed")
	}
	if len(args) > 0 {
		m.data[query] = args[0]
	} else {
		m.data[query] = true
	}
	return &memResult{lastInsertID: 1, rowsAffected: 1}, nil
}

// Query returns nil rows for memory placeholder queries.
func (m *MemDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil, fmt.Errorf("memdb: database is closed")
	}
	return nil, nil
}

// QueryRow returns an empty row for memory queries.
func (m *MemDB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

// BeginTx starts a staged in-memory transaction.
func (m *MemDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, fmt.Errorf("memdb: database is closed")
	}

	snapshot := make(map[string]any, len(m.data))
	for k, v := range m.data {
		snapshot[k] = v
	}

	return &MemTx{
		db:      m,
		staged:  snapshot,
		pending: make(map[string]any),
	}, nil
}

// Replica returns the configured replica instance or itself.
func (m *MemDB) Replica() database.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.replica
}

// SetReplica configures the read-replica database.
func (m *MemDB) SetReplica(db database.DB) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replica = db
}

// Connection returns a named connection if registered.
func (m *MemDB) Connection(name string) (database.DB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", database.ErrConnectionNotFound, name)
	}
	return conn, nil
}

// SetConnection registers a named connection.
func (m *MemDB) SetConnection(name string, db database.DB) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[name] = db
}

// GetValue returns stored key data for verification.
func (m *MemDB) GetValue(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// MemTx implements database.Tx for in-memory transactions.
type MemTx struct {
	mu      sync.Mutex
	db      *MemDB
	staged  map[string]any
	pending map[string]any
	closed  bool
}

// Exec stages an execution in the transaction.
func (tx *MemTx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, database.ErrTxClosed
	}
	if len(args) > 0 {
		tx.pending[query] = args[0]
	} else {
		tx.pending[query] = true
	}
	return &memResult{lastInsertID: 1, rowsAffected: 1}, nil
}

// Query is a placeholder for transactional queries.
func (tx *MemTx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, database.ErrTxClosed
	}
	return nil, nil
}

// QueryRow is a placeholder for transactional single-row queries.
func (tx *MemTx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

// Commit applies all staged writes to the parent MemDB.
func (tx *MemTx) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return database.ErrTxClosed
	}
	tx.closed = true

	tx.db.mu.Lock()
	defer tx.db.mu.Unlock()
	for k, v := range tx.pending {
		tx.db.data[k] = v
	}
	return nil
}

// Rollback discards all pending writes in the transaction.
func (tx *MemTx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return database.ErrTxClosed
	}
	tx.closed = true
	tx.pending = make(map[string]any)
	return nil
}

// Ensure driver.Result compliance
var _ driver.Result = (*memResult)(nil)
