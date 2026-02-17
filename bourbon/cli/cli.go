package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "bourbon",
	Short:   "Bourbon - A Django-like MVC framework for Go",
	Long:    `Bourbon is a lightweight, Django-inspired MVC framework for Go with built-in ORM, migrations, and code generators.`,
	Version: "1.0.0",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Bourbon v1.0.0")
		fmt.Println("A Django-like MVC framework for Go")
	},
}

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, _ := cmd.Flags().GetString("db")
		createProjectWithDB(args[0], db)
	},
}

var createAppCmd = &cobra.Command{
	Use:   "create:app [app-name]",
	Short: "Create a new application module",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createApp(args[0])
	},
}

var makeMigrationCmd = &cobra.Command{
	Use:   "make:migration",
	Short: "Create migrations (auto-detects changes if no app specified)",
	Run: func(cmd *cobra.Command, args []string) {
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		force, _ := cmd.Flags().GetBool("force")

		if app == "" {
			// Auto-detect changes in all apps (like Django)
			makeMigrationsForAllApps(name, force)
		} else {
			// Create migration for specific app
			makeMigrationForApp(app, name, force)
		}
	},
}

func init() {
	makeMigrationCmd.Flags().String("app", "", "Application name (optional, auto-detects all apps if not provided)")
	makeMigrationCmd.Flags().String("name", "", "Migration name (optional, uses sequential numbering if not provided)")
	makeMigrationCmd.Flags().Bool("force", false, "Force migration creation even if no changes detected")

	newCmd.Flags().String("db", "sqlite", "Database driver (sqlite, postgres, mysql)")

	rootCmd.AddCommand(
		versionCmd,
		newCmd,
		createAppCmd,
		makeMigrationCmd,
	)
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
