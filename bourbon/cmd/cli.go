package cmd

import (
	"fmt"
	"os"

	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/bourbon/middleware"
	_ "github.com/ishubhamsingh2e/bourbon/bourbon/database/drivers"
	"go.uber.org/zap"
)

// CommandHandler is a function that handles a command
type CommandHandler func(args []string) error

// commandRegistry holds all registered commands
var commandRegistry = map[string]CommandHandler{
	"make:migration":   handleMakeMigration,
	"migrate":          handleMigrate,
	"migrate:status":   handleMigrateStatus,
	"migrate:rollback": handleMigrateRollback,
}

// RegisterCommand allows users to register custom commands
func RegisterCommand(name string, handler CommandHandler) {
	commandRegistry[name] = handler
}

// Run is the main entry point for Bourbon applications
// It handles both CLI commands and server startup
func Run(configPath string) {
	if len(os.Args) > 1 {
		if err := HandleCommand(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Normal server startup
	StartServer(configPath)
}

// HandleCommand processes CLI commands
func HandleCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	handler, exists := commandRegistry[command]
	if !exists {
		return fmt.Errorf("unknown command: %s", command)
	}

	return handler(args[1:])
}

// StartServer initializes and starts the Bourbon server
func StartServer(configPath string) {
	app := core.NewApplication(configPath)

	// Initialize database
	if err := app.ConnectDB(); err != nil {
		app.Logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}

	// Call custom initialization hook if registered
	// This is where user's middleware.go SetupMiddleware is called
	if customInit != nil {
		if err := customInit(app); err != nil {
			app.Logger.Error("Custom initialization failed", zap.Error(err))
			os.Exit(1)
		}
	} else {
		// If no custom init, setup default middlewares as fallback
		SetupDefaultMiddlewares(app)
	}

	// Start the server
	if err := app.Run(); err != nil {
		app.Logger.Error("Server error", zap.Error(err))
	}
}

// SetupDefaultMiddlewares configures the default middleware stack
func SetupDefaultMiddlewares(app *core.Application) {
	app.RegisterMiddleware("recovery", middleware.Recovery(app.Logger, app.ErrorStore))
	app.UseMiddleware("recovery")

	app.RegisterMiddleware("logger", middleware.Logger(app.Logger, app.ErrorStore))
	app.UseMiddleware("logger")
}

// Custom initialization hook
var customInit func(*core.Application) error

// SetCustomInit allows users to set a custom initialization function
func SetCustomInit(fn func(*core.Application) error) {
	customInit = fn
}

// handleMakeMigration handles the make:migration command
func handleMakeMigration(args []string) error {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	return GenerateMigration(name)
}

// handleMigrate handles the migrate command
func handleMigrate(args []string) error {
	app := core.NewApplication("./settings.toml")

	if err := app.ConnectDB(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Running migrations...")
	if err := core.RunMigrations(app); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}

// handleMigrateStatus handles the migrate:status command
func handleMigrateStatus(args []string) error {
	app := core.NewApplication("./settings.toml")

	if err := app.ConnectDB(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return core.ShowMigrationStatus(app)
}

// handleMigrateRollback handles the migrate:rollback command
func handleMigrateRollback(args []string) error {
	app := core.NewApplication("./settings.toml")

	if err := app.ConnectDB(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Rolling back last migration...")
	if err := core.RollbackLastMigration(app); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback completed successfully")
	return nil
}
