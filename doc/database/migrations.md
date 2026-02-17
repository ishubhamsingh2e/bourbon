# Migrations

Bourbon uses a Django-like migration system to manage database schema changes.

## Creating Migrations

When you change your models, you create a migration file using the CLI:

```bash
bourbon make:migration
```

This command:
1.  Analyzes your `models.go` file.
2.  Compares it with the previous migration's state.
3.  Generates a new migration file in `apps/<app_name>/migrations/`.

### Options

- `bourbon make:migration --name add_category_id`: Provide a descriptive name for the migration.
- `bourbon make:migration --app posts`: Only check the `posts` app for changes.

## Migration Files

Migration files are auto-generated Go files that define `Up` and `Down` logic.

```go
package migrations

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "github.com/ishubhamsingh2e/bourbon/bourbon/core"
    "gorm.io/gorm"
)

func init() {
    migration := &gormigrate.Migration{
        ID: "20240215120000_CreatePostsTable",
        Migrate: func(tx *gorm.DB) error {
            return tx.AutoMigrate(&posts.Post{})
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Migrator().DropTable("posts")
        },
    }
    core.RegisterAppMigration("posts", migration)
}
```

## Running Migrations

Migrations are run automatically when your application starts.

This is configured in `main.go`:

```go
// main.go
if err := core.RunMigrations(app); err != nil {
    app.Logger.Error("Migration failed", zap.Error(err))
    os.Exit(1)
}
```

This function:
1.  Connects to the database.
2.  Initializes the `gormigrate` runner.
3.  Registers all migrations found in the `migrations` package of your apps.
4.  Executes pending migrations in order.

### Important: Importing Migrations

Ensure your migration packages are imported in `main.go` or `db.go` so their `init()` functions run and register the migrations.

```go
import (
    _ "myblog/apps/posts/migrations"
    _ "myblog/apps/users/migrations"
)
```

## Migration Status

You can check the status of applied migrations by querying the `gorm_migrations` table in your database.

Currently, there is no CLI command to list migration status, but you can inspect the database directly or implement a custom route to display it.
