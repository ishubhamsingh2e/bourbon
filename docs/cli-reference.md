# CLI Reference (`bourbon`)

The `bourbon` CLI manages application scaffolding, live reloading, code generation, diagnostics, and build lifecycles.

---

## Commands

### `bourbon new <name>`
Scaffold a new Bourbon application.
- `--preset=api|web|worker|fullstack`: Preset selection.
- `--structure=feature|layered`: Directory layout convention.
- `--module=<path>`: Custom Go module name.

### `bourbon dev`
Starts the live-reload development server with reverse-proxy request buffering.

### `bourbon build`
Compiles an optimized production binary into `bin/app`.

### `bourbon test`
Runs the Go test suite with `-race` detector and `-cover` profile.

### `bourbon makemigrations [app]`
Inspects Go struct models via AST parsing, calculates diffs against previous state, prompts interactively for field/model renames and default values, and writes delta SQL migration files.
- `--name=<name>`: Custom migration name.
- `--empty`: Create an empty migration skeleton.
- `--no-input` / `--noinput`: Non-interactive mode applying defaults.

### `bourbon migrate [app]`
Apply all pending database migrations in chronological order.

### `bourbon showmigrations [app]`
Display all project migrations with `[X]` (applied) and `[ ]` (unapplied) state.

### `bourbon sqlmigrate <app> <migration_name>`
Print the raw SQL statements for a migration without executing them.
- `--backwards`: Print rollback (Down) SQL statements.

### `bourbon db <action>`
Manage database migrations:
- `bourbon db makemigrations [app]`: Inspect models and generate migration files.
- `bourbon db create [name]`: Create a new timestamped SQL migration file (`name` is optional, defaults to `migration`).
- `bourbon db migrate`: Run pending migrations.
- `bourbon db rollback`: Rollback the last applied migration.
- `bourbon db status`: Display migration status table.
- `bourbon db showmigrations`: Display all migrations with applied status.
- `bourbon db sqlmigrate <name>`: Print raw SQL for a migration.
- `bourbon db version`: Display active database schema version.
- `bourbon db reset`: Roll back all applied migrations.
- `bourbon db seed`: Seed database fixtures.

### `bourbon routes`
Display all registered HTTP route patterns, methods, and handlers.

### `bourbon doctor`
Inspect local Go environment, dependencies, toolchain, and manifests.

### `bourbon upgrade`
Checks for newer framework releases.

---

## Code Generators

- `bourbon make:controller <name>`: Generate an HTTP controller.
- `bourbon make:service <name>`: Generate a domain business service.
- `bourbon make:repository <name>`: Generate a data repository.
- `bourbon make:model <name>`: Generate a database model struct.
- `bourbon make:migration [name]`: Generate a timestamped SQL migration file with table boilerplate (`name` is optional).
- `bourbon make:middleware <name>`: Generate an HTTP middleware.
- `bourbon make:job <name>`: Generate a background job worker.
