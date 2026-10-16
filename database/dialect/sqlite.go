package dialect

func init() {
	Register("sqlite", &SQLiteDialect{})
}

// SQLiteDialect implements the Dialect interface for SQLite.
type SQLiteDialect struct{}

// Name returns the dialect name.
func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

// DriverName returns the standard driver identifier.
func (d *SQLiteDialect) DriverName() string {
	return "sqlite"
}

// QuoteIdentifier wraps identifiers in double quotes.
func (d *SQLiteDialect) QuoteIdentifier(name string) string {
	return `"` + name + `"`
}

// Rebind returns positional query unaltered as SQLite uses '?'.
func (d *SQLiteDialect) Rebind(query string) string {
	return query
}

// AutoIncrementKeyword returns the autoincrement clause.
func (d *SQLiteDialect) AutoIncrementKeyword() string {
	return "AUTOINCREMENT"
}

// DefaultTimestamp returns the current timestamp expression.
func (d *SQLiteDialect) DefaultTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// UUIDType returns the column type for UUIDs.
func (d *SQLiteDialect) UUIDType() string {
	return "VARCHAR(36)"
}

// BooleanType returns the column type for booleans.
func (d *SQLiteDialect) BooleanType() string {
	return "BOOLEAN"
}

// JSONType returns the column type for JSON.
func (d *SQLiteDialect) JSONType() string {
	return "TEXT"
}
