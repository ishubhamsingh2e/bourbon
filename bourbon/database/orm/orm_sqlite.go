//go:build sqlite || all_drivers

package orm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver("sqlite", sqliteDialector)
}

func sqliteDialector(cfg DatabaseConfig) (gorm.Dialector, error) {
	path := cfg.Path
	if path == "" {
		path = cfg.Name
	}
	if path == "" {
		path = "bourbon.db"
	}
	return sqlite.Open(path), nil
}
