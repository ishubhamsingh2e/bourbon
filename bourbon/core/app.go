package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ishubhamsingh2e/bourbon/bourbon/core/gormigrate"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core/registry"
	"github.com/ishubhamsingh2e/bourbon/bourbon/database/orm"
	bourbon "github.com/ishubhamsingh2e/bourbon/bourbon/http"
	"github.com/ishubhamsingh2e/bourbon/bourbon/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App represents the main application structure
type App struct {
	Config             *Config                      // Application configuration
	Router             *bourbon.Router              // HTTP router
	Server             *http.Server                 // HTTP server
	Logger             *logging.Logger              // Structured logger
	ErrorStore         *logging.ErrorStore          // Error store for logging server errors to database
	Registry           *registry.Registry           // Global registry for app components
	DB                 *gorm.DB                     // Database connection
	BasePath           string                       // Base path for the application
	Apps               []string                     // List of registered apps/modules
	GormigrateRunner   *gormigrate.GormigrateRunner // Gormigrate migration runner
	MiddlewareRegistry *registry.MiddlewareRegistry // Middleware registry
	middlewareStack    []registry.MiddlewareFunc    // Ordered list of middlewares
	middlewareMu       sync.RWMutex                 // Mutex for middleware stack
}

type Application = App

// NewApp creates a new instance of App with default values
func NewApp() *App {
	logger, _ := logging.NewLogger(logging.DefaultConfig())
	return &App{
		Router:             bourbon.NewRouter(),
		Logger:             logger,
		Registry:           registry.NewRegistry(),
		BasePath:           ".",
		Apps:               make([]string, 0),
		MiddlewareRegistry: registry.NewMiddlewareRegistry(),
		middlewareStack:    make([]registry.MiddlewareFunc, 0),
	}
}

func NewApplication(configPath string) *Application {
	app := NewApp()

	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	app.Config = config

	// Initialize logger with config
	loggerConfig := &logging.LoggerConfig{
		FileLogging: config.Logging.FileLogging,
		StoragePath: config.Logging.StoragePath,
		Rotation:    logging.LogRotation(config.Logging.Rotation),
		MaxSize:     config.Logging.MaxSize,
		MaxAge:      config.Logging.MaxAge,
		MaxBackups:  config.Logging.MaxBackups,
		Compress:    config.Logging.Compress,
		Level:       config.Logging.Level,
		Development: config.App.Debug,
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	app.Logger = logger

	// Initialize error store if database error logging is enabled
	if config.Logging.StoreErrorsInDB {
		app.ErrorStore = logging.NewErrorStore(app.DB, true)
		// Run migration for error logs table if DB is connected
		if app.DB != nil {
			if err := app.ErrorStore.Migrate(); err != nil {
				app.Logger.Warn("Failed to migrate error logs table", zap.Error(err))
			}
		}
	}

	if config.Templates.Directory != "" {
		engine := bourbon.NewTemplateEngine(
			config.Templates.Directory,
			config.Templates.Extension,
			config.Templates.AutoReload,
		)

		if err := engine.Load(); err != nil {
			app.Logger.Warn("Failed to load templates", zap.Error(err), zap.String("directory", config.Templates.Directory))
		} else {
			app.Router.TemplateEngine = engine
		}
	}

	return app
}

// RegisterMiddleware registers a named middleware in the app's registry
func (a *App) RegisterMiddleware(name string, middleware registry.MiddlewareFunc) {
	a.MiddlewareRegistry.Register(name, middleware)
}

// UseMiddleware adds a registered middleware to the stack by name
func (a *App) UseMiddleware(name string) error {
	middleware, exists := a.MiddlewareRegistry.Get(name)
	if !exists {
		return fmt.Errorf("middleware '%s' not registered", name)
	}

	a.middlewareMu.Lock()
	defer a.middlewareMu.Unlock()
	a.middlewareStack = append(a.middlewareStack, middleware)
	return nil
}

// UseMiddlewareFunc adds a middleware function directly to the stack
func (a *App) UseMiddlewareFunc(middleware registry.MiddlewareFunc) {
	a.middlewareMu.Lock()
	defer a.middlewareMu.Unlock()
	a.middlewareStack = append(a.middlewareStack, middleware)
}

// Use is an alias for UseMiddlewareFunc for convenience
func (a *App) Use(middleware registry.MiddlewareFunc) {
	a.UseMiddlewareFunc(middleware)
}

// ClearMiddlewares removes all middlewares from the stack
func (a *App) ClearMiddlewares() {
	a.middlewareMu.Lock()
	defer a.middlewareMu.Unlock()
	a.middlewareStack = make([]registry.MiddlewareFunc, 0)
}

// GetMiddlewares returns a copy of the current middleware stack
func (a *App) GetMiddlewares() []registry.MiddlewareFunc {
	a.middlewareMu.RLock()
	defer a.middlewareMu.RUnlock()

	stack := make([]registry.MiddlewareFunc, len(a.middlewareStack))
	copy(stack, a.middlewareStack)
	return stack
}

// buildHandler applies all middlewares in the stack to the router
func (a *App) buildHandler() http.Handler {
	a.middlewareMu.RLock()
	defer a.middlewareMu.RUnlock()

	handler := http.Handler(a.Router)

	// Apply middlewares in reverse order (last registered wraps first)
	for i := len(a.middlewareStack) - 1; i >= 0; i-- {
		handler = a.middlewareStack[i](handler)
	}

	return handler
}

func (a *App) LoadConfig(path string) error {
	config, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.Config = config
	return nil
}

func (a *App) RegisterApp(name string) {
	a.Apps = append(a.Apps, name)
	a.Logger.Info("Registered app", zap.String("name", name))
}

func (app *Application) Run() error {
	app.printStartupBanner()

	// Build handler with middleware stack
	handler := app.buildHandler()

	// Create server if not already created
	if app.Server == nil {
		app.Server = &http.Server{
			Addr:           fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port),
			Handler:        handler,
			ReadTimeout:    time.Duration(app.Config.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(app.Config.Server.WriteTimeout) * time.Second,
			MaxHeaderBytes: app.Config.Server.MaxHeaderBytes,
		}
	} else {
		// Update handler if server already exists
		app.Server.Handler = handler
	}

	if app.Config.Static.Directory != "" && app.Config.Static.URLPrefix != "" {
		app.Static(app.Config.Static.URLPrefix, app.Config.Static.Directory)
		app.Logger.Info("Static files mounted",
			zap.String("prefix", app.Config.Static.URLPrefix),
			zap.String("directory", app.Config.Static.Directory))
	}

	go func() {
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("Server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	app.Logger.Info("Server stopped")
	return nil
}

func (a *App) Static(prefix, root string) {
	a.Router.Static(prefix, root)
}

func (a *App) AddTemplateFunc(name string, fn interface{}) {
	if a.Router.TemplateEngine != nil {
		a.Router.TemplateEngine.AddFunc(name, fn)
		if a.Config.Templates.AutoReload {
			_ = a.Router.TemplateEngine.Load()
		}
	}
}

func (a *App) AddTemplateFuncs(funcs map[string]interface{}) {
	if a.Router.TemplateEngine != nil {
		for name, fn := range funcs {
			a.Router.TemplateEngine.AddFunc(name, fn)
		}
		if a.Config.Templates.AutoReload {
			_ = a.Router.TemplateEngine.Load()
		}
	}
}

func (app *Application) printStartupBanner() {
	host := app.Config.Server.Host
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}

	protocol := "http"
	url := fmt.Sprintf("%s://%s:%d", protocol, host, app.Config.Server.Port)

	fmt.Printf("Application: %s\n", app.Config.App.Name)
	fmt.Printf("Environment: %s\n", app.Config.App.Env)
	fmt.Printf("Debug Mode:  %v\n", app.Config.App.Debug)
	fmt.Printf("Host:        %s\n", app.Config.Server.Host)
	fmt.Printf("Port:        %d\n", app.Config.Server.Port)
	fmt.Printf("URL:         %s\n", url)
	fmt.Println()
	fmt.Printf("> Press Ctrl+C to stop\n")
	fmt.Println()
}

// ConnectDB establishes database connection using the application configuration
func (a *App) ConnectDB() error {
	if a.Config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Convert Config.Database to orm.DatabaseConfig
	dbConfig := orm.DatabaseConfig{
		Driver:          a.Config.Database.Driver,
		Host:            a.Config.Database.Host,
		Port:            a.Config.Database.Port,
		Name:            a.Config.Database.Name,
		User:            a.Config.Database.User,
		Password:        a.Config.Database.Password,
		Path:            a.Config.Database.Path,
		MaxOpenConns:    a.Config.Database.MaxOpenConns,
		MaxIdleConns:    a.Config.Database.MaxIdleConns,
		ConnMaxLifetime: a.Config.Database.ConnMaxLifetime,
		Options: orm.DatabaseOptions{
			SSLMode:    a.Config.Database.Options.SSLMode,
			LogQueries: a.Config.Database.Options.LogQueries,
		},
	}

	db, err := orm.ConnectDatabase(dbConfig, a.Config.App.Debug)
	if err != nil {
		return err
	}

	a.DB = db
	return nil
}

// InitMigrations initializes the gormigrate runner with registered migrations
func (a *App) InitMigrations() error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	a.GormigrateRunner = gormigrate.NewGormigrateRunner(a.DB)
	migrations := gormigrate.GetGormigrateMigrations()

	if len(migrations) > 0 {
		a.GormigrateRunner.AddMigrations(migrations)
		if err := a.GormigrateRunner.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize migrations: %w", err)
		}
	} else {
		a.Logger.Warn("No migrations registered")
	}

	return nil
}

// Migrate runs all pending migrations
func (a *App) Migrate() error {
	if a.GormigrateRunner == nil {
		if err := a.InitMigrations(); err != nil {
			return err
		}
	}
	return a.GormigrateRunner.Migrate()
}

// RollbackLast rolls back the last migration
func (a *App) RollbackLast() error {
	if a.GormigrateRunner == nil {
		if err := a.InitMigrations(); err != nil {
			return err
		}
	}
	return a.GormigrateRunner.RollbackLast()
}

// RollbackTo rolls back to a specific migration
func (a *App) RollbackTo(migrationID string) error {
	if a.GormigrateRunner == nil {
		if err := a.InitMigrations(); err != nil {
			return err
		}
	}
	return a.GormigrateRunner.RollbackTo(migrationID)
}

// MigrateTo migrates to a specific migration
func (a *App) MigrateTo(migrationID string) error {
	if a.GormigrateRunner == nil {
		if err := a.InitMigrations(); err != nil {
			return err
		}
	}
	return a.GormigrateRunner.MigrateTo(migrationID)
}
