package config

import (
	"time"

	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/cache/memcache"
	bconfig "github.com/ishubhamsingh2e/bourbon/config"
	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/events/inprocess"
	"github.com/ishubhamsingh2e/bourbon/storage"
	"github.com/ishubhamsingh2e/bourbon/storage/localfs"
)

// Settings represents the strongly-typed application configuration.
type Settings struct {
	AppName     string `env:"APP_NAME,default=todo"`
	Environment string `env:"APP_ENV,default=development"`
	Port        int    `env:"PORT,default=8080"`
	Host        string `env:"HOST,default=0.0.0.0"`
	DatabaseURL string `env:"DATABASE_URL,default=file:storage.db?cache=shared&mode=rwc"`
}

// Load initializes application settings from .env and environment variables.
func Load() *Settings {
	var s Settings
	_ = bconfig.Load(&s)
	return &s
}

// DBDriver returns the configured database driver provider.
func (s *Settings) DBDriver() database.Driver {
	return sqlite.Config{
		DSN: s.DatabaseURL,
	}
}

// CacheDriver returns the configured cache driver provider.
func (s *Settings) CacheDriver() cache.Driver {
	return memcache.Config{
		CleanupInterval: 1 * time.Minute,
	}
}

// StorageDriver returns the configured file storage driver provider.
func (s *Settings) StorageDriver() storage.Driver {
	return localfs.Config{
		RootDir: "storage/uploads",
		BaseURL: "/storage",
	}
}

// EventDriver returns the configured event bus driver provider.
func (s *Settings) EventDriver() events.Driver {
	return inprocess.Config{
		BufferSize: 1024,
	}
}

// AppConfig translates settings into the core Bourbon framework configuration.
func (s *Settings) AppConfig() *bconfig.Config {
	env := bconfig.EnvDevelopment
	switch s.Environment {
	case "production":
		env = bconfig.EnvProduction
	case "staging":
		env = bconfig.EnvStaging
	}

	return &bconfig.Config{
		App: bconfig.AppConfig{
			Name: s.AppName,
		},
		HTTP: bconfig.HTTPConfig{
			Host: s.Host,
			Port: s.Port,
		},
		Env: env,
	}
}
