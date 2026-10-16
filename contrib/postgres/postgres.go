// Package postgres provides a PostgreSQL adapter for database.DB.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config configures the PostgreSQL connection pool.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DSN             string
}

// DB implements database.DB for PostgreSQL.
type DB struct {
	primary  *sql.DB
	replicas []*sql.DB
	named    map[string]*sql.DB
}

// New constructs a Postgres database instance wrapping *sql.DB.
func New(primary *sql.DB, replicas ...*sql.DB) *DB {
	return &DB{
		primary:  primary,
		replicas: replicas,
		named:    make(map[string]*sql.DB),
	}
}

// Open initializes a connection to PostgreSQL using standard DSN.
func Open(cfg Config) (*DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		if cfg.Port == 0 {
			cfg.Port = 5432
		}
		if cfg.SSLMode == "" {
			cfg.SSLMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed opening connection: %w", err)
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
	return db.primary.ExecContext(ctx, query, args...)
}

// Query executes a query returning rows.
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.primary.QueryContext(ctx, query, args...)
}

// QueryRow executes a query returning at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.primary.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Tx, error) {
	tx, err := db.primary.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Ping verifies connectivity.
func (db *DB) Ping(ctx context.Context) error {
	return db.primary.PingContext(ctx)
}

// Close closes all connections.
func (db *DB) Close() error {
	return db.primary.Close()
}

// Replica returns a read replica if configured.
func (db *DB) Replica() database.DB {
	if len(db.replicas) > 0 {
		return New(db.replicas[0])
	}
	return db
}

// Connection returns a named connection.
func (db *DB) Connection(name string) (database.DB, error) {
	if namedDB, ok := db.named[name]; ok {
		return New(namedDB), nil
	}
	return nil, database.ErrConnectionNotFound
}

// RegisterConnection binds a named connection pool.
func (db *DB) RegisterConnection(name string, namedDB *sql.DB) {
	db.named[name] = namedDB
}

// SQLDB returns the underlying primary *sql.DB connection pool.
func (db *DB) SQLDB() *sql.DB {
	return db.primary
}

// DriverName returns the database driver name for migrations.
func (db *DB) DriverName() string {
	return "postgres"
}

// Tx wraps a standard sql.Tx.
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
