//go:build mysql || all_drivers

package orm

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver("mysql", mysqlDialector)
}

func mysqlDialector(cfg DatabaseConfig) (gorm.Dialector, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
	return mysql.Open(dsn), nil
}
