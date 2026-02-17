package orm

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DialectorFunc func(cfg DatabaseConfig) (gorm.Dialector, error)

var (
	driverRegistry = make(map[string]DialectorFunc)
	driverMutex    sync.RWMutex
)

func RegisterDriver(name string, fn DialectorFunc) {
	driverMutex.Lock()
	defer driverMutex.Unlock()
	driverRegistry[name] = fn
}

func GetDialector(name string) (DialectorFunc, bool) {
	driverMutex.RLock()
	defer driverMutex.RUnlock()
	fn, ok := driverRegistry[name]
	return fn, ok
}

func ListDrivers() []string {
	driverMutex.RLock()
	defer driverMutex.RUnlock()
	drivers := make([]string, 0, len(driverRegistry))
	for name := range driverRegistry {
		drivers = append(drivers, name)
	}
	return drivers
}

// ConnectDatabase creates a new database connection
func ConnectDatabase(cfg DatabaseConfig, debug bool) (*gorm.DB, error) {
	driverFunc, ok := GetDialector(cfg.Driver)
	if !ok {
		return nil, fmt.Errorf("unsupported or unavailable database driver: %s (use build tags: -tags=postgres or -tags=mysql or -tags=sqlite or -tags=all_drivers)", cfg.Driver)
	}

	dialector, err := driverFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialector: %w", err)
	}

	gormLogger := logger.Default.LogMode(logger.Silent)
	if debug {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Apply connection pool settings
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 25
	}
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 5
	}
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

