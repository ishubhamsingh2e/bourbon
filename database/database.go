// Package database provides the core database port, transaction management, and connection routing contracts.
package database

import (
	"context"
	"database/sql"
	"errors"
)

var (
	// ErrNotFound is returned when a requested database record does not exist.
	ErrNotFound = errors.New("database: record not found")
	// ErrTxClosed is returned when an operation is attempted on a committed or rolled-back transaction.
	ErrTxClosed = errors.New("database: transaction already committed or rolled back")
	// ErrConnectionNotFound is returned when a requested named connection does not exist.
	ErrConnectionNotFound = errors.New("database: named connection not found")
)

// DB represents a pluggable database connection pool.
type DB interface {
	Ping(ctx context.Context) error
	Close() error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Replica() DB
	Connection(name string) (DB, error)
}

// Conn represents an active single database connection.
type Conn interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
}

// Tx represents an active database transaction.
type Tx interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
}

// Driver represents a database connection provider capable of opening a DB instance.
type Driver interface {
	Open() (DB, error)
}
