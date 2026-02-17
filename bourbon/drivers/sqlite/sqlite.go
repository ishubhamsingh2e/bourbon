// Package sqlite provides SQLite database driver for Bourbon framework.
// Import this package to enable SQLite support:
//
// import _ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/sqlite"
package sqlite

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/database/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	orm.RegisterDriver("sqlite", sqliteDialector)
}

func sqliteDialector(cfg orm.DatabaseConfig) (gorm.Dialector, error) {
	path := cfg.Path
	if path == "" {
		path = cfg.Name
	}
	if path == "" {
		path = "bourbon.db"
	}
	return sqlite.Open(path), nil
}
