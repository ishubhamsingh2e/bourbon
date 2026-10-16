package dialect

import (
	"strconv"
	"strings"
)

func init() {
	Register("postgres", &PostgresDialect{})
}

// PostgresDialect implements the Dialect interface for PostgreSQL.
type PostgresDialect struct{}

// Name returns the dialect name.
func (d *PostgresDialect) Name() string {
	return "postgres"
}

// DriverName returns the standard driver identifier.
func (d *PostgresDialect) DriverName() string {
	return "pgx"
}

// QuoteIdentifier wraps identifiers in double quotes.
func (d *PostgresDialect) QuoteIdentifier(name string) string {
	return `"` + name + `"`
}

// Rebind converts positional '?' placeholders to '$1, $2, ...' syntax.
func (d *PostgresDialect) Rebind(query string) string {
	var builder strings.Builder
	builder.Grow(len(query) + 16)

	paramIdx := 1
	inSingleQuote := false
	inDoubleQuote := false
	runes := []rune(query)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
			builder.WriteRune(r)
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
			builder.WriteRune(r)
		case '?':
			if !inSingleQuote && !inDoubleQuote {
				builder.WriteByte('$')
				builder.WriteString(strconv.Itoa(paramIdx))
				paramIdx++
			} else {
				builder.WriteRune(r)
			}
		default:
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// AutoIncrementKeyword returns the autoincrement clause for PostgreSQL.
func (d *PostgresDialect) AutoIncrementKeyword() string {
	return "BIGSERIAL"
}

// DefaultTimestamp returns the current timestamp expression.
func (d *PostgresDialect) DefaultTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// UUIDType returns the column type for UUIDs.
func (d *PostgresDialect) UUIDType() string {
	return "UUID"
}

// BooleanType returns the column type for booleans.
func (d *PostgresDialect) BooleanType() string {
	return "BOOLEAN"
}

// JSONType returns the column type for JSON.
func (d *PostgresDialect) JSONType() string {
	return "JSONB"
}
