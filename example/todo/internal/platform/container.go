package platform

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/mail/devmail"
	"github.com/ishubhamsingh2e/bourbon/observability"
	"github.com/ishubhamsingh2e/bourbon/storage"
	"todo/config"
)

// Container manages all technical infrastructure adapters and connection pools.
type Container struct {
	DB      database.DB
	Cache   cache.Cache
	Bus     events.Bus
	Mailer  *devmail.Mailer
	Metrics *observability.MetricsCollector
	Storage storage.Storage
}

// NewContainer initializes infrastructure adapters according to application settings.
func NewContainer(cfg *config.Settings) (*Container, error) {
	var db database.DB
	if driver := cfg.DBDriver(); driver != nil {
		var err error
		db, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening database: %w", err)
		}
	}

	var ca cache.Cache
	if driver := cfg.CacheDriver(); driver != nil {
		var err error
		ca, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening cache: %w", err)
		}
	}

	var st storage.Storage
	if driver := cfg.StorageDriver(); driver != nil {
		var err error
		st, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening storage: %w", err)
		}
	}

	var bus events.Bus
	if driver := cfg.EventDriver(); driver != nil {
		var err error
		bus, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening event bus: %w", err)
		}
	}

	mailer := devmail.New()
	metrics := observability.NewMetricsCollector()

	return &Container{
		DB:      db,
		Cache:   ca,
		Bus:     bus,
		Mailer:  mailer,
		Metrics: metrics,
		Storage: st,
	}, nil
}
