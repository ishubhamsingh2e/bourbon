# Configuration & Dependency Injection in Bourbon

Bourbon utilizes an explicit configuration loader and a type-safe Dependency Injection (DI) Container to manage infrastructure dependencies and service lifecycles without global state.

---

## 1. Application Configuration

Bourbon loads configuration from environment variables and `.env` files using `config.LoadFromEnv()`:

```go
package main

import (
	"github.com/ishubhamsingh2e/bourbon/config"
)

func loadAppConfig() *config.Config {
	cfg := config.LoadFromEnv()
	// cfg.App: Name, Key, Debug, URL, Timezone
	// cfg.HTTP: Host, Port, ReadTimeout, WriteTimeout, IdleTimeout
	// cfg.Env: EnvDevelopment, EnvStaging, EnvProduction, EnvTesting
	return cfg
}
```

### Supported Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_NAME` | Application identifier | `"Bourbon App"` |
| `APP_ENV` | Runtime environment (`development`, `staging`, `production`, `testing`) | `"development"` |
| `APP_KEY` | 32-byte secret key for signing tokens and sessions | `""` |
| `APP_DEBUG` | Boolean flag enabling debug logs and stack traces | `true` |
| `APP_URL` | Base canonical URL of the application | `"http://localhost:8000"` |
| `HTTP_HOST` | Host interface for HTTP binding | `"0.0.0.0"` |
| `HTTP_PORT` | Port number for HTTP binding | `8000` |
| `HTTP_READ_TIMEOUT` | HTTP server read timeout duration (e.g. `15s`) | `15s` |
| `HTTP_WRITE_TIMEOUT`| HTTP server write timeout duration (e.g. `15s`) | `15s` |

---

## 2. Type-Safe Dependency Injection Container

Bourbon includes a concurrent, reflection-based DI container (`app.Container`) supporting three binding styles:

1. **Instances (`c.Singleton(val)`)**: Binds an already instantiated value.
2. **Singleton Factories (`c.Singleton(fn)`)**: Factory evaluated once on first resolution; cached thereafter.
3. **Transient Factories (`c.Transient(fn)`)**: Factory invoked on every resolution call.

```go
package bootstrap

import (
	"database/sql"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
)

func RegisterServices(c *app.Container) {
	// 1. Bind an instance singleton
	c.Singleton(sqlite.New(dbConn))

	// 2. Bind a factory with dependency autowiring
	c.Singleton(func(db *sqlite.DB) *UserRepository {
		return NewUserRepository(db)
	})

	// 3. Bind a transient service
	c.Transient(func(repo *UserRepository) *UserService {
		return NewUserService(repo)
	})
}
```

---

## 3. Resolving Dependencies

To retrieve dependencies from the container, pass a pointer to the target variable:

```go
var userService *UserService
if err := container.Resolve(&userService); err != nil {
	log.Fatalf("failed resolving UserService: %v", err)
}
```

The container automatically inspects constructor function arguments, resolves sub-dependencies recursively, and reports clear errors for missing bindings or circular dependency cycles (`app.ErrCircularDependency`).

---

## 4. Modular Service Providers

To keep large codebases clean and decoupled, encapsulate subsystem lifecycle into modular `core.Provider` implementations:

```go
package providers

import (
	"context"
	"database/sql"

	"github.com/ishubhamsingh2e/bourbon/core"
)

type DatabaseProvider struct {
	db *sql.DB
}

func (p *DatabaseProvider) Register(c core.Container) error {
	c.Singleton(func() (*sql.DB, error) {
		db, err := sql.Open("sqlite", "app.db")
		if err != nil {
			return nil, err
		}
		p.db = db
		return db, nil
	})
	return nil
}

func (p *DatabaseProvider) Boot(ctx context.Context, app core.Application) error {
	// Execute health ping or warm up connection pool
	var db *sql.DB
	if err := app.Container().Resolve(&db); err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func (p *DatabaseProvider) Shutdown(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
```

Register providers during application startup using `application.Use(...)`:

```go
application := app.New()
application.Use(&providers.DatabaseProvider{})
```
