// Package mysql provides MySQL database driver for Bourbon framework.
// Import this package to enable MySQL support:
//
// import _ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/mysql"
package mysql

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/bourbon/database/orm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	orm.RegisterDriver("mysql", mysqlDialector)
}

func mysqlDialector(cfg orm.DatabaseConfig) (gorm.Dialector, error) {
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
