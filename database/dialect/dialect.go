// Package dialect provides cross-database SQL dialect abstraction, parameter rebinding, and schema DDL generators.
package dialect

import (
	"fmt"
	"strings"
	"sync"
)

// Dialect defines the contracts for database-specific SQL syntax, parameter translation, and identifiers.
type Dialect interface {
	Name() string
	DriverName() string
	QuoteIdentifier(name string) string
	Rebind(query string) string
	AutoIncrementKeyword() string
	DefaultTimestamp() string
	UUIDType() string
	BooleanType() string
	JSONType() string
}

var (
	dialectsMu sync.RWMutex
	dialects   = make(map[string]Dialect)
)

// Register registers a new SQL dialect by name.
func Register(name string, d Dialect) {
	dialectsMu.Lock()
	defer dialectsMu.Unlock()
	dialects[strings.ToLower(name)] = d
}

// Get returns the dialect registered for the given driver or database name.
func Get(name string) (Dialect, error) {
	dialectsMu.RLock()
	dialectsMu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(name))
	if d, ok := dialects[norm]; ok {
		return d, nil
	}

	// Alias lookups
	switch norm {
	case "sqlite3", "sqlite":
		if d, ok := dialects["sqlite"]; ok {
			return d, nil
		}
	case "postgres", "postgresql", "pgx", "pq":
		if d, ok := dialects["postgres"]; ok {
			return d, nil
		}
	case "mysql", "mariadb":
		if d, ok := dialects["mysql"]; ok {
			return d, nil
		}
	}

	return nil, fmt.Errorf("dialect: unsupported database dialect %q", name)
}
