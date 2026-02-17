// Package postgres provides PostgreSQL database driver for Bourbon framework.
// Import this package to enable PostgreSQL support:
//
// import _ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/postgres"
package postgres

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/bourbon/database/orm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	orm.RegisterDriver("postgres", postgresDialector)
}

func postgresDialector(cfg orm.DatabaseConfig) (gorm.Dialector, error) {
	sslMode := cfg.Options.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		sslMode,
	)
	return postgres.Open(dsn), nil
}
