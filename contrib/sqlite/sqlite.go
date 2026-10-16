// Package sqlite provides an embedded SQLite adapter for database.DB.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/database"
	_ "modernc.org/sqlite"
)

// Config configures SQLite connection settings.
type Config struct {
	DSN             string
	EnableWAL       bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DB implements database.DB for SQLite.
type DB struct {
	db    *sql.DB
	named map[string]*sql.DB
}

// New constructs an SQLite database instance wrapping *sql.DB.
func New(db *sql.DB) *DB {
	return &DB{
		db:    db,
		named: make(map[string]*sql.DB),
	}
}

// Open opens an SQLite database.
func Open(cfg Config) (*DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed opening database: %w", err)
	}

	maxOpen := 1
	if cfg.MaxOpenConns > 0 {
		maxOpen = cfg.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

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

// BeginTx starts an SQLite transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Ping verifies SQLite connectivity.
func (db *DB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// Close closes the SQLite connection.
func (db *DB) Close() error {
	return db.db.Close()
}

// Replica returns the primary database since SQLite does not use read replicas.
func (db *DB) Replica() database.DB {
	return db
}

// Connection returns a named connection if registered.
func (db *DB) Connection(name string) (database.DB, error) {
	if namedDB, ok := db.named[name]; ok {
		return New(namedDB), nil
	}
	return nil, database.ErrConnectionNotFound
}

// SQLDB returns the underlying *sql.DB connection pool.
func (db *DB) SQLDB() *sql.DB {
	return db.db
}

// DriverName returns the database driver name for migrations.
func (db *DB) DriverName() string {
	return "sqlite3"
}

// Tx wraps a standard sql.Tx for SQLite.
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

// Provider implements core.Provider for SQLite database integration.
type Provider struct {
	cfg Config
	db  *DB
}

// NewProvider returns a new SQLite Provider with the given configuration.
func NewProvider(cfg Config) *Provider {
	return &Provider{cfg: cfg}
}

// Register registers database.DB into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	c.Singleton(func() (database.DB, error) {
		db, err := Open(p.cfg)
		if err != nil {
			return nil, err
		}
		p.db = db
		return db, nil
	})
	return nil
}

// Boot verifies connectivity by pinging the database.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var db database.DB
	if err := app.Container().Resolve(&db); err != nil {
		return err
	}
	return db.Ping(ctx)
}

// Shutdown gracefully closes the SQLite database connection.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

var _ core.Provider = (*Provider)(nil)
