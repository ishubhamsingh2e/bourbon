package orm

import (
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	Path     string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	Options DatabaseOptions
}

// DatabaseOptions holds database connection options
type DatabaseOptions struct {
	SSLMode    string
	LogQueries bool
}
