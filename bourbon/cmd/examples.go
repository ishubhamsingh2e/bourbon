package cmd

// This file contains examples of how to customize Bourbon's behavior.
// These examples are not executed - they're for documentation purposes.

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	bourbonHttp "github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

// Example 1: Custom initialization hook
// Use this to add routes and custom setup
func exampleCustomInit() {
	SetCustomInit(func(app *core.Application) error {
		// Add custom routes
		app.Router.Get("/custom", func(ctx *bourbonHttp.Context) error {
			return ctx.String(200, "Custom route")
		})

		return nil
	})
}

// Example 2: Register custom commands
// Add your own CLI commands
func exampleCustomCommands() {
	// Register a seed command
	RegisterCommand("seed", func(args []string) error {
		app := core.NewApplication("./settings.toml")
		if err := app.ConnectDB(); err != nil {
			return err
		}

		fmt.Println("Seeding database...")
		// Your seeding logic here
		return nil
	})

	// Register a clear cache command
	RegisterCommand("cache:clear", func(args []string) error {
		fmt.Println("Clearing cache...")
		// Your cache clearing logic here
		return nil
	})
}

// Example 3: Override existing commands
// You can replace built-in commands
func exampleOverrideCommand() {
	RegisterCommand("migrate", func(args []string) error {
		fmt.Println("Custom migration logic!")
		// Your custom migration logic here
		return nil
	})
}
