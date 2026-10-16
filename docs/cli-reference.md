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

### `bourbon migrate [command]`
Manage database migrations using your project's transparent migration runner (`cmd/migrate/main.go`).
- `bourbon migrate` / `bourbon migrate up`: Apply all pending database migrations in `db/migrations/`.
- `bourbon migrate rollback` / `bourbon migrate down`: Revert the last applied database migration.
- `bourbon migrate status`: Display a table showing applied vs pending migrations.
- `bourbon migrate create <name>`: Generate a new timestamped SQL migration file in `db/migrations/`.

### `bourbon routes`
Display all registered HTTP route patterns, methods, and handlers.

### `bourbon doctor`
Inspect local Go environment, dependencies, toolchain, and manifests.

### `bourbon upgrade`
Checks for newer framework releases.

### `bourbon config`
Inspect and validate Bourbon configuration.

---

## Code Generators (`bourbon make`)

- `bourbon make app <name>`: Generate a complete vertical feature slice.
- `bourbon make controller <name>`: Generate an HTTP controller.
- `bourbon make service <name>`: Generate a domain business service.
- `bourbon make repository <name>`: Generate a data repository.
- `bourbon make model <name>`: Generate a database model struct.
- `bourbon make migration <name>`: Generate a timestamped SQL migration file in `db/migrations/`.
- `bourbon make middleware <name>`: Generate an HTTP middleware.
- `bourbon make job <name>`: Generate a background job worker.
- `bourbon make test <name>`: Generate a test suite skeleton.
