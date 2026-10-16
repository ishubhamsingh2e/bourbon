// Package config defines the application configuration models and environment types.
package config

import "time"

// Environment represents the execution runtime environment.
type Environment string

const (
	// EnvDevelopment represents local development environment.
	EnvDevelopment Environment = "development"
	// EnvStaging represents staging pre-production environment.
	EnvStaging Environment = "staging"
	// EnvProduction represents production live environment.
	EnvProduction Environment = "production"
	// EnvTesting represents automated test runner environment.
	EnvTesting Environment = "testing"
)

// Config aggregates all subsystem configuration settings.
type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	Env  Environment
}

// AppConfig defines application metadata and behavior settings.
type AppConfig struct {
	Name     string
	Key      string
	Debug    bool
	URL      string
	Timezone string
}

// HTTPConfig defines HTTP server transport and timeout settings.
type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}
