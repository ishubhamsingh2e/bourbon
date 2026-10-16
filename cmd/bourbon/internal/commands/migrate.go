package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/config"
	"github.com/ishubhamsingh2e/bourbon/database/migrator"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate [command]",
		Short: "Manage database migrations",
		Long: `Manage database migrations using your project's transparent migration runner (cmd/migrate/main.go).

Commands:
  migrate        Apply pending migrations (alias for 'migrate up')
  migrate up     Apply pending migrations
  migrate rollback Rollback the last applied migration (alias: down)
  migrate status Show migration status
  migrate create Generate a new migration file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectMigrate(cmd, "up")
		},
	}

	migrateCmd.AddCommand(
		newMigrateUpCmd(),
		newMigrateRollbackCmd(),
		newMigrateStatusCmd(),
		newMigrateCreateCmd(),
	)

	return migrateCmd
}

func newMigrateUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Apply all pending database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectMigrate(cmd, "up")
		},
	}
}

func newMigrateRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rollback",
		Aliases: []string{"down"},
		Short:   "Rollback the last applied database migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectMigrate(cmd, "rollback")
		},
	}
}

func newMigrateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display the status of all database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectMigrate(cmd, "status")
		},
	}
}

func newMigrateCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Generate a new SQL migration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			rootDir, _, err := config.FindProjectRoot(".")
			if err != nil || rootDir == "" {
				rootDir = "."
			}

			migDir := filepath.Join(rootDir, "db", "migrations")
			dest, err := migrator.Create(migDir, name, "", "")
			if err != nil {
				return fmt.Errorf("failed creating migration file: %w", err)
			}

			relDest, err := filepath.Rel(rootDir, dest)
			if err != nil {
				relDest = dest
			}
			cmd.Printf("Created migration: %s\n", relDest)
			return nil
		},
	}
}

func runProjectMigrate(cmd *cobra.Command, action string) error {
	rootDir, _, err := config.FindProjectRoot(".")
	if err != nil || rootDir == "" {
		rootDir = "."
	}

	runnerPath := filepath.Join(rootDir, "cmd", "migrate", "main.go")
	if _, err := os.Stat(runnerPath); os.IsNotExist(err) {
		return fmt.Errorf("no migration runner found at %s. Ensure cmd/migrate/main.go exists in your project", filepath.Join("cmd", "migrate", "main.go"))
	}

	c := exec.Command("go", "run", "./cmd/migrate", action)
	c.Dir = rootDir
	c.Stdout = cmd.OutOrStdout()
	c.Stderr = cmd.OutOrStderr()
	c.Stdin = cmd.InOrStdin()

	if err := c.Run(); err != nil {
		return fmt.Errorf("migration runner failed: %w", err)
	}
	return nil
}
