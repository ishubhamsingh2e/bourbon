package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"todo/config"

	"github.com/ishubhamsingh2e/bourbon/database/migrator"
)

func main() {
	cfg := config.Load()
	driver := cfg.DBDriver()
	if driver == nil {
		log.Fatal("migrator: no database driver configured in config.Settings")
	}

	db, err := driver.Open()
	if err != nil {
		log.Fatalf("migrator: failed to open database: %v", err)
	}
	defer db.Close()

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	migrationDir := "db/migrations"

	switch action {
	case "up", "migrate":
		if err := migrator.Migrate(db, migrationDir); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully.")

	case "rollback", "down":
		if err := migrator.Rollback(db, migrationDir); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully.")

	case "status":
		statuses, err := migrator.Status(db, migrationDir)
		if err != nil {
			log.Fatalf("status check failed: %v", err)
		}

		if len(statuses) == 0 {
			fmt.Println("No migrations found in " + migrationDir)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tMIGRATION\tAPPLIED AT")
		for _, s := range statuses {
			statusStr := "Pending"
			appliedAtStr := "-"
			if s.Applied {
				statusStr = "Applied"
				if s.AppliedAt != nil {
					appliedAtStr = s.AppliedAt.Format("2006-01-02 15:04:05")
				}
			}
			fmt.Fprintf(w, "%s\t%d_%s\t%s\n", statusStr, s.Version, s.Name, appliedAtStr)
		}
		_ = w.Flush()

	default:
		fmt.Fprintf(os.Stderr, "Unknown migration command: %s (supported: up, rollback, status)\n", action)
		os.Exit(1)
	}
}
