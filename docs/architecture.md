# Architecture & Design Philosophy

Bourbon is architected around the Hexagonal (Ports and Adapters) pattern, providing modularity and low coupling while strictly enforcing a zero-third-party-runtime dependency core.

---

## 1. Monorepo & Module Architecture

Bourbon is structured to clearly separate core contracts, standard subsystem implementations, and developer tooling:

```text
bourbon/
├── core/                 # Core contracts: Application, Container, Provider, Context
├── app/                  # Application orchestrator, lifecycle, and DI container
├── router/               # High-performance Radix Trie router with native Context handler support
├── context/              # Backward-compatible re-exports for core.Context
├── middleware/           # Recovery, logging, CORS, CSRF, security headers
├── database/             # database.DB port, transaction manager, schema & migrator
├── model/                # Base models, timestamps, and metadata
├── repository/           # Repository structural conventions
├── cache/                # Cache port with atomic operations & TTLs (in-memory default)
├── session/              # Stateful session engine & flash messaging
├── auth/                 # Multi-guard authentication, RBAC, and crypto
├── validation/           # Declarative schema and struct validation
├── view/                 # View rendering abstraction
├── template/             # Go html/template engine with layout composition
├── events/               # Event bus with sync & async dispatchers (in-process default)
├── jobs/                 # Background job queue with worker pool (in-process default)
├── schedule/             # Cron scheduling subsystem
├── realtime/             # SSE real-time room broadcaster
├── storage/              # File storage port & local filesystem driver
├── mail/                 # Mailer port & SMTP/devmail adapters
├── observability/        # Health checks & Prometheus metrics
├── openapi/              # Dynamic OpenAPI 3.0 specification builder
├── contrib/              # Ecosystem drivers (SQLite, Postgres, MySQL, Redis, S3, OTel)
└── cmd/bourbon/          # Developer CLI tool binary
```

---

## 2. Minimal Core Application Shell

`core.Application` defines the minimal contract required to execute an HTTP service. It exposes:
- **`Container() core.Container`**: Dependency injection registry and service resolution.
- **`Router() router.Router`**: HTTP routing engine with native `func(c *core.Context) error` handlers.
- **`Logger() logger.Logger`**: Structured application logger (defaulting to stdlib `log/slog`).
- **`Use(providers ...core.Provider) core.Application`**: Modular service provider registration.
- **`Run(ctx context.Context) error`**: Provider registration, boot sequence, and graceful HTTP listening.
- **`Shutdown(ctx context.Context) error`**: Graceful shutdown of HTTP server and providers in reverse order.

---

## 3. The 3-Phase Provider Lifecycle

Bourbon decouples all subsystems (databases, queues, caches, storage) from the shell via `core.Provider`:

```
   1. Register(c Container) error
        │  Bind singleton/transient factories, configs, and adapters into the DI container.
        ▼
   2. Boot(ctx Context, app Application) error
        │  Active initialization (ping database, connect to queues, mount routes) after all providers register.
        ▼
   3. Shutdown(ctx Context) error
        │  Executed in REVERSE registration order upon SIGINT/SIGTERM to close connection pools.
```

### Standard Subsystem Providers

| Subsystem | Contract | Bourbon Default Driver | Enterprise Driver Example |
| :--- | :--- | :--- | :--- |
| **Database** | `database.DB` | `contrib/sqlite.NewProvider(cfg)` | PostgreSQL, MySQL, CockroachDB |
| **Cache** | `cache.Cache` | `cache/memcache.NewProvider(cfg)` | Redis, Memcached |
| **Job Queue** | `jobs.Queue` | `jobs/inprocess.NewProvider()` | Redis Queue, SQS, RabbitMQ |
| **Event Bus** | `events.Bus` | `events/inprocess.NewProvider()` | Kafka, NATS, RabbitMQ |
| **Storage** | `storage.Storage` | `storage/localfs.NewProvider(cfg)` | AWS S3, Cloudflare R2 |
| **Mailer** | `mail.Mailer` | `mail/devmail.NewProvider()` | SMTP, SendGrid, SES |

---

## 4. Bootstrapping with Providers

```go
package main

import (
	"context"
	"log"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
	"github.com/ishubhamsingh2e/bourbon/core"
)

func main() {
	application := app.New()

	// Hook in required subsystem providers
	application.Use(
		sqlite.NewProvider(sqlite.Config{DSN: "app.db"}),
	)

	r := application.Router()

	// Native *core.Context handler: zero wrappers needed
	r.Get("/healthz", func(c *core.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})

	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("application terminated: %v", err)
	}
}
```

---

## 5. The Context Engine (`*core.Context`)

The request lifecycle revolves around `*core.Context`, eliminating package name collisions with Go's standard `"context"` package:

```go
package main

import (
	"context" // Standard library context
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/core" // Bourbon context
)

func HandleUpdateProfile(c *core.Context) error {
	var req UpdateProfileDTO
	if err := c.BindJSON(&req); err != nil {
		return err
	}

	user := c.Auth()
	db := c.DB()

	return c.JSON(http.StatusOK, map[string]any{
		"status": "profile updated",
		"user":   user.ID(),
	})
}
```
