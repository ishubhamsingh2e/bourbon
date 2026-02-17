//go:build postgres || all_drivers

package orm

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver("postgres", postgresDialector)
}

func postgresDialector(cfg DatabaseConfig) (gorm.Dialector, error) {
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
