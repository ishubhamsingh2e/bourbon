# CLI Reference

Bourbon provides a command-line interface (CLI) to help you create and manage your projects.

## Installation Commands

These commands are run using the `bourbon` CLI tool after installation.

### `bourbon new`

Creates a new Bourbon project structure with your choice of database.

**Usage:**

```bash
bourbon new <project-name> [--db=<database>]
```

**Flags:**

- `--db`: Database driver to use (sqlite, postgres, mysql). Default: sqlite

**Examples:**

```bash
# Create with SQLite (default)
bourbon new myblog
bourbon new myblog --db=sqlite

# Create with PostgreSQL
bourbon new myblog --db=postgres

# Create with MySQL
bourbon new myblog --db=mysql
```

This creates a new directory with:
- Project structure (apps/, templates/, static/, storage/)
- main.go with correct database driver import
- settings.toml configured for chosen database
- Basic app module matching project name

**Database-specific setup:**
- **SQLite**: Zero config, database file created automatically
- **PostgreSQL**: Update settings.toml with your database credentials
- **MySQL**: Update settings.toml with your database credentials

### `bourbon create:app`

Creates a new application module within your project.

**Usage:**

```bash
bourbon create:app <app-name>
```

**Example:**

```bash
bourbon create:app posts
```

This creates a directory `apps/posts` with `models.go`, `controllers.go`, and `routes.go`.

**Note:** After creating an app, remember to add it to `settings.toml` under `[apps.installed]`.

### `bourbon version`

Displays the current version of the Bourbon CLI.

**Usage:**

```bash
bourbon version
```

---

## Runtime Commands

These commands are run through your application's main entry point after building your project.

### `go run main.go` (or `go run .`)

Starts the development server with default settings. Automatically runs pending migrations on startup.

**Usage:**

```bash
go run main.go
# or
go run .
```

### `make:migration`

Detects changes in your models and creates a new migration file.

**Usage:**

```bash
go run main.go make:migration [flags]
# or
go run . make:migration [flags]
```

**Flags:**

- `--app string`: Specify the application name to check for changes. If omitted, checks all apps.
- `--name string`: Provide a descriptive name for the migration.
- `--force`: Force creation of migration even if no changes are detected.

**Examples:**

```bash
# Auto-detect changes in all apps
go run . make:migration

# Create migration for specific app
go run . make:migration --app=posts

# Provide a custom name
go run . make:migration --name=add_author_id

# Force creation without changes
go run . make:migration --force
```

**Note:** When you modify models, the system will:
1. Scan your models.go files for changes
2. Detect additions, deletions, and type changes
3. Warn about destructive changes (field deletions)
4. Generate a timestamped migration file

### `migrate`

Runs all pending migrations for your application.

**Usage:**

```bash
go run . migrate
```

**Note:** Migrations are automatically run on application startup, so this command is typically only needed for deployment scripts or manual migration management.

### `migrate:status`

Shows the status of all migrations, grouped by application.

**Usage:**

```bash
go run . migrate:status
```

**Output Example:**

```
Migration Status:
=================

App: posts
  [✓] 20260101120000_initial.go (Applied)
  [✓] 20260215100000_add_author_field.go (Applied)
  [ ] 20260217120000_add_tags.go (Pending)

App: users
  [✓] 20260110150000_create_users.go (Applied)
```

### `migrate:rollback`

Rolls back the last applied migration or to a specific version.

**Usage:**

```bash
# Rollback last migration
go run . migrate:rollback

# Rollback to specific migration ID
go run . migrate:rollback --to=20260215100000
```

**Options:**

- `--to string`: Migration ID to rollback to (exclusive - rolls back everything after this ID)

**Warning:** Rollbacks can cause data loss. Always backup your database before rolling back.

## Global Flags

- `--help`: Show help for any command.

## Shell Completion

You can generate shell completion scripts for your shell using standard Cobra commands if enabled (check `bourbon completion --help`).

## Common Workflows

### Creating a New Project

1. `bourbon new myproject`
2. `cd myproject`
3. `go mod tidy`
4. `go run main.go`

### Adding a New Feature

1. `bourbon create:app comments`
2. Define `Comment` struct in `apps/comments/models.go`.
3. `bourbon make:migration --app=comments --name=create_comments`
4. `go run main.go` (Runs migration automatically)
