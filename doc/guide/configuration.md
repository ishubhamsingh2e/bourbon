# Configuration

Bourbon uses `settings.toml` for application configuration. This file contains sensible defaults for a typical web application.

## Example `settings.toml`

```toml
[app]
name = "myblog"
debug = true
secret_key = "change-me-in-production"
timezone = "UTC"
env = "development"

[server]
host = "127.0.0.1"
port = 8000
read_timeout = 30
write_timeout = 30
max_header_bytes = 1048576

[database]
driver = "sqlite"
path = "storage/database.db"

# For PostgreSQL
# driver = "postgres"
# host = "localhost"
# port = 5432
# name = "myblog_db"
# user = "dbuser"
# password = "dbpass"
# max_open_conns = 25
# max_idle_conns = 5
# conn_max_lifetime = 3600

[database.options]
ssl_mode = "disable"
log_queries = false

[apps]
installed = [
    "myblog"
]

[middleware]
enabled = [
    "Logger",
    "Recovery",
    "CORS"
]

[templates]
directory = "templates"
extension = ".html"
auto_reload = true

[static]
directory = "static"
url_prefix = "/static"

[logging]
level = "info"
format = "json"
output = "stdout"
file_logging = false
storage_path = "storage/logs"
rotation = "daily"  # Options: hourly, daily, weekly, none
max_size = 100      # MB per log file
max_age = 30        # days to retain logs
max_backups = 10    # number of old log files to keep
compress = true     # compress old logs
store_errors_db = false  # store 5xx errors in database

[security]
allowed_hosts = ["localhost", "127.0.0.1"]
cors_origins = ["http://localhost:3000"]
```

## Key Configuration Sections

### `[app]`

- `name`: The name of your application.
- `debug`: Enable debug mode (e.g., development error pages).
- `secret_key`: Used for signing cookies and sessions. Change this in production.
- `timezone`: Default timezone for the application.
- `env`: Environment (e.g., `development`, `production`).

### `[server]`

- `host`: The IP address to bind to. `127.0.0.1` is for local access only, `0.0.0.0` for all interfaces.
- `port`: The port to listen on.
- `read_timeout`: Timeout for reading requests (seconds).
- `write_timeout`: Timeout for writing responses (seconds).

### `[database]`

- `driver`: Supported drivers: `sqlite`, `postgres`, `mysql`.
- `path`: Path to SQLite database file.
- `host`, `port`, `name`, `user`, `password`: Connection details for PostgreSQL/MySQL.
- `max_open_conns`: Maximum number of open connections to the database.
- `max_idle_conns`: Maximum number of idle connections.
- `conn_max_lifetime`: Maximum lifetime of a connection (seconds).

### `[middleware]`

- `enabled`: List of middleware names to enable globally.

### `[logging]`

- `level`: Minimum log level (`debug`, `info`, `warn`, `error`).
- `format`: Output format (`json` or `console`).
- `rotation`: Log rotation frequency (`daily`, `hourly`, `weekly`, `none`).
- `file_logging`: Enable logging to files.
- `store_errors_db`: If true, stores 500 errors in the database.

### `[security]`

- `allowed_hosts`: List of allowed hostnames/IPs for incoming requests.
- `cors_origins`: Allowed origins for CORS requests.

## Environment Variables

Environment variables override settings in `settings.toml`. The convention is `BOURBON_<SECTION>_<KEY>`.

Example:

```bash
export BOURBON_DATABASE_PASSWORD="secret"
export BOURBON_SERVER_PORT="8080"
```
