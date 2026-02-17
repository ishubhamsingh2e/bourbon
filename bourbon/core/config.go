package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Apps       AppsConfig       `mapstructure:"apps"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
	Templates  TemplatesConfig  `mapstructure:"templates"`
	Static     StaticConfig     `mapstructure:"static"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Security   SecurityConfig   `mapstructure:"security"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	Debug     bool   `mapstructure:"debug"`
	SecretKey string `mapstructure:"secret_key"`
	Timezone  string `mapstructure:"timezone"`
}

type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Path     string `mapstructure:"path"`

	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	Options DatabaseOptions `mapstructure:"options"`
}

type DatabaseOptions struct {
	SSLMode    string `mapstructure:"ssl_mode"`
	LogQueries bool   `mapstructure:"log_queries"`
}

type AppsConfig struct {
	Installed []string `mapstructure:"installed"`
}

type MiddlewareConfig struct {
	Enabled []string `mapstructure:"enabled"`
}

type TemplatesConfig struct {
	Directory  string   `mapstructure:"directory"`
	Extension  string   `mapstructure:"extension"`
	AutoReload bool     `mapstructure:"auto_reload"`
	Funcs      []string `mapstructure:"funcs"`
}

type StaticConfig struct {
	Directory string `mapstructure:"directory"`
	URLPrefix string `mapstructure:"url_prefix"`
}

type LoggingConfig struct {
	Level           string `mapstructure:"level"`
	Format          string `mapstructure:"format"`
	Output          string `mapstructure:"output"`
	FileLogging     bool   `mapstructure:"file_logging"`
	StoragePath     string `mapstructure:"storage_path"`
	Rotation        string `mapstructure:"rotation"`        // hourly, daily, weekly, none
	MaxSize         int    `mapstructure:"max_size"`        // MB
	MaxAge          int    `mapstructure:"max_age"`         // days
	MaxBackups      int    `mapstructure:"max_backups"`     // number of backups
	Compress        bool   `mapstructure:"compress"`        // compress old logs
	StoreErrorsInDB bool   `mapstructure:"store_errors_db"` // store 5xx errors in database
}

type SecurityConfig struct {
	AllowedHosts      []string `mapstructure:"allowed_hosts"`
	CorsOrigins       []string `mapstructure:"cors_origins"`
	CSRFEnabled       bool     `mapstructure:"csrf_enabled"`
	SessionTimeout    int      `mapstructure:"session_timeout"`
	SessionCookieName string   `mapstructure:"session_cookie_name"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	setGlobalDefaults(v)

	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvPrefix("BOURBON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	config.loadEnvOverrides()

	return &config, nil
}

func setGlobalDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "bourbon-app")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.secret_key", "change-me-in-production")
	v.SetDefault("app.timezone", "UTC")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8000)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.max_header_bytes", 1048576)

	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "bourbon.db")
	v.SetDefault("database.path", "storage/database.db")
	v.SetDefault("database.user", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.options.ssl_mode", "disable")
	v.SetDefault("database.options.log_queries", false)

	v.SetDefault("apps.installed", []string{})

	v.SetDefault("middleware.enabled", []string{"Logger", "Recovery"})

	v.SetDefault("templates.directory", "templates")
	v.SetDefault("templates.extension", ".html")
	v.SetDefault("templates.auto_reload", true)

	v.SetDefault("static.directory", "static")
	v.SetDefault("static.url_prefix", "/static")

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file_logging", false)
	v.SetDefault("logging.storage_path", "storage/logs")
	v.SetDefault("logging.rotation", "daily")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_age", 30)
	v.SetDefault("logging.max_backups", 10)
	v.SetDefault("logging.compress", true)
	v.SetDefault("logging.store_errors_db", false)

	v.SetDefault("security.allowed_hosts", []string{"localhost", "127.0.0.1"})
	v.SetDefault("security.cors_origins", []string{"*"})
	v.SetDefault("security.csrf_enabled", false)
	v.SetDefault("security.session_timeout", 3600)

}

func (c *Config) loadEnvOverrides() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Database.Port = p
		}
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		c.Database.Name = name
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		c.Database.Password = pass
	}

	if debug := os.Getenv("DEBUG"); debug != "" {
		c.App.Debug = debug == "true"
	}
	if secret := os.Getenv("SECRET_KEY"); secret != "" {
		c.App.SecretKey = secret
	}
}

func GetViper(configPath string) (*viper.Viper, error) {
	v := viper.New()
	setGlobalDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("toml")
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, err
			}
		}
	}

	v.SetEnvPrefix("BOURBON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v, nil
}
