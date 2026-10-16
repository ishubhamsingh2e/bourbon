package dialect

func init() {
	Register("mysql", &MySQLDialect{})
}

// MySQLDialect implements the Dialect interface for MySQL and MariaDB.
type MySQLDialect struct{}

// Name returns the dialect name.
func (d *MySQLDialect) Name() string {
	return "mysql"
}

// DriverName returns the standard driver identifier.
func (d *MySQLDialect) DriverName() string {
	return "mysql"
}

// QuoteIdentifier wraps identifiers in backticks.
func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

// Rebind returns positional query unaltered as MySQL/MariaDB uses '?'.
func (d *MySQLDialect) Rebind(query string) string {
	return query
}

// AutoIncrementKeyword returns the autoincrement clause for MySQL/MariaDB.
func (d *MySQLDialect) AutoIncrementKeyword() string {
	return "AUTO_INCREMENT"
}

// DefaultTimestamp returns the current timestamp expression.
func (d *MySQLDialect) DefaultTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// UUIDType returns the column type for UUIDs.
func (d *MySQLDialect) UUIDType() string {
	return "VARCHAR(36)"
}

// BooleanType returns the column type for booleans.
func (d *MySQLDialect) BooleanType() string {
	return "TINYINT(1)"
}

// JSONType returns the column type for JSON.
func (d *MySQLDialect) JSONType() string {
	return "JSON"
}
