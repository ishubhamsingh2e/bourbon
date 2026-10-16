// Package mysql provides a MySQL and MariaDB database adapter for database.DB.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ishubhamsingh2e/bourbon/database"
)

// Config configures MySQL and MariaDB connection settings.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB implements database.DB for MySQL and MariaDB.
type DB struct {
	db      *sql.DB
	replica *sql.DB
	named   map[string]*sql.DB
}

// New constructs a MySQL/MariaDB database instance wrapping *sql.DB.
func New(db *sql.DB) *DB {
	return &DB{
		db:    db,
		named: make(map[string]*sql.DB),
	}
}

// Open opens a MySQL or MariaDB database connection pool.
func Open(cfg Config) (*DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		return nil, fmt.Errorf("mysql: dsn is required")
	}

	// Ensure parseTime=true is set so time.Time scans properly
	if !strings.Contains(dsn, "parseTime=") {
		if strings.Contains(dsn, "?") {
			dsn += "&parseTime=true"
		} else {
			dsn += "?parseTime=true"
		}
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: failed opening database: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	return New(db), nil
}

// Open satisfies the database.Driver interface.
func (cfg Config) Open() (database.DB, error) {
	return Open(cfg)
}

// Exec executes a statement without returning rows.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

// Query executes a query returning rows.
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query returning at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a MySQL transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Ping verifies database connectivity.
func (db *DB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// Close closes the connection pool.
func (db *DB) Close() error {
	return db.db.Close()
}

// Replica returns the read replica if configured, or primary.
func (db *DB) Replica() database.DB {
	if db.replica != nil {
		return New(db.replica)
	}
	return db
}

// SetReplica sets a read-only replica connection pool.
func (db *DB) SetReplica(replica *sql.DB) {
	db.replica = replica
}

// Connection returns a named connection if registered.
func (db *DB) Connection(name string) (database.DB, error) {
	if namedDB, ok := db.named[name]; ok {
		return New(namedDB), nil
	}
	return nil, database.ErrConnectionNotFound
}

// RegisterConnection registers a named connection pool.
func (db *DB) RegisterConnection(name string, conn *sql.DB) {
	db.named[name] = conn
}

// SQLDB returns the underlying *sql.DB connection pool.
func (db *DB) SQLDB() *sql.DB {
	return db.db
}

// DriverName returns the database driver name for migrations.
func (db *DB) DriverName() string {
	return "mysql"
}

// Tx wraps a standard sql.Tx for MySQL.
type Tx struct {
	tx *sql.Tx
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

var _ database.DB = (*DB)(nil)
var _ database.Tx = (*Tx)(nil)
