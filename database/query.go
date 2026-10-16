// Package database provides the core database port, transaction management, and connection routing contracts.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// Queryer defines operations for executing SQL read queries.
type Queryer interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
}

// Execer defines operations for executing SQL write queries.
type Execer interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Params represents key-value named parameters for raw queries.
type Params map[string]any

type queryAdapter struct {
	q Queryer
}

func (a *queryAdapter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if qc, ok := a.q.(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}); ok {
		return qc.QueryContext(ctx, query, args...)
	}
	return a.q.Query(ctx, query, args...)
}

func toSQLScanQuerier(q Queryer) sqlscan.Querier {
	if qc, ok := q.(sqlscan.Querier); ok {
		return qc
	}
	return &queryAdapter{q: q}
}

// QueryAll executes a raw SQL query with safe parameter bindings and scans the resulting rows into a slice of type T.
func QueryAll[T any](ctx context.Context, q Queryer, query string, args ...any) ([]T, error) {
	if q == nil {
		return nil, errors.New("database: queryer cannot be nil")
	}

	normQuery, normArgs, err := normalizeParams(query, args)
	if err != nil {
		return nil, fmt.Errorf("database: query error: %w", err)
	}

	var results []T
	querier := toSQLScanQuerier(q)
	if err := sqlscan.Select(ctx, querier, &results, normQuery, normArgs...); err != nil {
		return nil, err
	}

	if results == nil {
		results = make([]T, 0)
	}
	return results, nil
}

// QueryOne executes a raw SQL query and scans a single resulting row into type T.
// Returns ErrNotFound if no rows matched the query.
func QueryOne[T any](ctx context.Context, q Queryer, query string, args ...any) (*T, error) {
	if q == nil {
		return nil, errors.New("database: queryer cannot be nil")
	}

	normQuery, normArgs, err := normalizeParams(query, args)
	if err != nil {
		return nil, fmt.Errorf("database: query error: %w", err)
	}

	var result T
	querier := toSQLScanQuerier(q)
	if err := sqlscan.Get(ctx, querier, &result, normQuery, normArgs...); err != nil {
		if sqlscan.NotFound(err) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &result, nil
}

// Exec executes a SQL write statement with safe parameter bindings.
func Exec(ctx context.Context, e Execer, query string, args ...any) (sql.Result, error) {
	if e == nil {
		return nil, errors.New("database: execer cannot be nil")
	}

	normQuery, normArgs, err := normalizeParams(query, args)
	if err != nil {
		return nil, fmt.Errorf("database: exec error: %w", err)
	}

	return e.Exec(ctx, normQuery, normArgs...)
}

// RawQuery represents a fluent query builder for raw parameterized SQL statements.
type RawQuery struct {
	q     Queryer
	e     Execer
	query string
	args  []any
	err   error
}

// Raw creates a fluent raw SQL query statement with positional or named parameters.
func Raw(target any, query string, args ...any) *RawQuery {
	rq := &RawQuery{
		query: query,
	}

	if q, ok := target.(Queryer); ok {
		rq.q = q
	}
	if e, ok := target.(Execer); ok {
		rq.e = e
	}

	normQuery, normArgs, err := normalizeParams(query, args)
	if err != nil {
		rq.err = err
		return rq
	}

	rq.query = normQuery
	rq.args = normArgs
	return rq
}

// ScanOne scans a single row into the destination struct or primitive pointer.
func (r *RawQuery) ScanOne(ctx context.Context, dst any) error {
	if r.err != nil {
		return r.err
	}
	if r.q == nil {
		return errors.New("database: queryer not available on raw query")
	}

	querier := toSQLScanQuerier(r.q)
	if err := sqlscan.Get(ctx, querier, dst, r.query, r.args...); err != nil {
		if sqlscan.NotFound(err) || errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ScanAll scans multiple rows into the destination slice pointer.
func (r *RawQuery) ScanAll(ctx context.Context, dst any) error {
	if r.err != nil {
		return r.err
	}
	if r.q == nil {
		return errors.New("database: queryer not available on raw query")
	}

	querier := toSQLScanQuerier(r.q)
	return sqlscan.Select(ctx, querier, dst, r.query, r.args...)
}

// Exec executes the raw query as a write statement.
func (r *RawQuery) Exec(ctx context.Context) (sql.Result, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.e == nil {
		return nil, errors.New("database: execer not available on raw query")
	}

	return r.e.Exec(ctx, r.query, r.args...)
}

var namedParamPattern = regexp.MustCompile(`:([a-zA-Z0-9_]+)`)

func normalizeParams(query string, args []any) (string, []any, error) {
	if len(args) == 0 {
		return query, args, nil
	}

	// Check if first arg is Params or map[string]any
	if len(args) == 1 && args[0] != nil {
		val := reflect.ValueOf(args[0])
		if val.Kind() == reflect.Map {
			paramsMap := make(map[string]any)
			for _, key := range val.MapKeys() {
				if key.Kind() == reflect.String {
					paramsMap[key.String()] = val.MapIndex(key).Interface()
				}
			}
			return bindNamed(query, paramsMap)
		}
	}

	return query, args, nil
}

func bindNamed(query string, params map[string]any) (string, []any, error) {
	var boundArgs []any
	var missingParams []string

	replaced := namedParamPattern.ReplaceAllStringFunc(query, func(match string) string {
		paramName := strings.TrimPrefix(match, ":")
		val, exists := params[paramName]
		if !exists {
			missingParams = append(missingParams, paramName)
			return match
		}
		boundArgs = append(boundArgs, val)
		return "?"
	})

	if len(missingParams) > 0 {
		return "", nil, fmt.Errorf("missing named parameters: %s", strings.Join(missingParams, ", "))
	}

	return replaced, boundArgs, nil
}
