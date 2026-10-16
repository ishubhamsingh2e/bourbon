# 🥃 Bourbon Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/ishubhamsingh2e/bourbon)](https://goreportcard.com/report/github.com/ishubhamsingh2e/bourbon)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A batteries-included, zero-dependency core web framework for Go.**

Bourbon brings the developer velocity and elegance of Django and Laravel to the Go ecosystem—delivering rich CLI scaffolding, background job queues, multi-guard authentication, and server-rendered view engines while strictly upholding Go's standard library philosophy: explicitness, zero unnecessary abstractions, static typing, low reflection, and sub-millisecond startup times.

---

## 📚 Documentation Portal

Comprehensive technical guides are available in the [docs/](docs/index.md) directory:

- **[Getting Started](docs/getting-started.md)**: Your first 5 minutes with Bourbon.
- **[Architecture & Design Philosophy](docs/architecture.md)**: Monorepo layout, zero-dependency core, and application lifecycle.
- **[Configuration & Dependency Injection](docs/config-and-di.md)**: Environment variable loading and the type-safe DI container.
- **[Routing & Middleware](docs/routing-and-middleware.md)**: Go 1.22+ ServeMux routing, versioning, REST resources, and middleware.
- **[Validation & Error Handling](docs/validation.md)**: Declarative validation rules, struct tags, and RFC 7807 Problem Details.
- **[Database, Models & Migrations](docs/database-and-orm.md)**: Pluggable database port, unit-of-work transactions, and embedded migrations.
- **[Authentication & RBAC](docs/authentication-and-rbac.md)**: Multi-guard auth (`SessionGuard`, `TokenGuard`) and wildcard RBAC permissions.
- **[Templating](docs/templating.md)**: Server-rendered HTML templates with layout composition, partials, and dev-mode hot reloading.
- **[Events, Background Jobs & Real-Time](docs/events-and-jobs.md)**: In-process event bus, job queue with worker pool, cron, and SSE.
- **[Storage & Mailers](docs/storage-and-mail.md)**: Local/cloud object storage abstractions and SMTP/devmail mailers.
- **[Observability & OpenAPI](docs/observability-and-openapi.md)**: Liveness/readiness health probes, Prometheus metrics, and OpenAPI 3.0 docs.
- **[Testing & Contract Verification](docs/testing.md)**: Fluent HTTP test client, in-memory database fake (`memdb`), and adapter compliance suites.
- **[CLI Reference](docs/cli-reference.md)**: Scaffolding, live-reloading dev server, code generators, and database tools.

---

## ⚡ Core Features

- **Zero-Dependency Core**: The core framework depends exclusively on the Go standard library and Go-team maintained `golang.org/x/*` packages.
- **Explicit Builder Pattern**: Clear, idiomatic `main.go` bootstrap without hidden reflection magic.
- **Zero-Allocation Routing**: Powered by Go 1.22+ `http.ServeMux` conventions with route grouping, API versioning (`r.Version("v1")`), and REST resource helpers.
- **Ergonomic Request Context**: Unified `*context.Context` providing discrete binding (`BindJSON`, `BindQuery`, `BindForm`), RFC 7807 Problem Details, and helper response methods.
- **Data Layer & Transactions**: Pluggable `database.DB` port, unit-of-work transaction wrapper (`database.WithTx`), generic repositories (`repository.Repository[T, ID]`), and an in-memory test fake (`database/memdb`).
- **Embedded Migrations**: Built-in migration runner supporting SQL and Go migration files across SQLite, PostgreSQL, and MySQL.
- **Multi-Guard Authentication**: Stateful cookie session guards and stateless token guards with wildcard RBAC permissions (`users:*`).
- **View & Templating**: Server-rendered `html/template` engine with layout/partial composition, built-in helpers (`csrf_token`, `auth`, `flash`), and dev-mode hot reloading.
- **Asynchronous Processing**: In-process event bus, background job queues with concurrency pools and retry backoff, cron task scheduling, and real-time Server-Sent Events (SSE).
- **Production Observability**: Built-in Kubernetes health probes (`/healthz`, `/readyz`), Prometheus metrics scraper (`/metrics`), and dynamic OpenAPI 3.0 documentation.
- **First-Class Developer CLI**: Live reload with request buffer proxy (`bourbon dev`), project scaffolding (`bourbon new`), route inspection (`bourbon routes`), and code generators (`bourbon make:*`).

---

## 🚀 Quick Start

### 1. Install the CLI

```bash
# Install the Bourbon CLI binary
go install github.com/ishubhamsingh2e/bourbon/cmd/bourbon@latest

# Verify installation
bourbon doctor
```

### 2. Scaffold a New Project

Choose between 4 presets (`api`, `web`, `fullstack`, `worker`) and 2 directory layouts (`feature` vertical-slice or `layered`):

```bash
# Create a REST API with vertical-slice feature organization
bourbon new my-api --preset=api --structure=feature

cd my-api
```

### 3. Start the Live-Reload Development Server

```bash
bourbon dev
```

The Bourbon development server starts an in-memory reverse proxy at `http://localhost:3000`. During code edits, browser requests are buffered in-memory until the Go binary finishes recompiling, ensuring zero dropped requests or half-rendered pages.

---

## 💡 Code Example

Bourbon keeps `main.go` clean, explicit, and immediately understandable:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/config"
	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func main() {
	// 1. Initialize application shell
	cfg := config.LoadFromEnv()
	application := app.New(
		app.WithConfig(cfg),
	)

	// 2. Hook in required subsystem providers
	application.Use(
		sqlite.NewProvider(sqlite.Config{DSN: "app.db"}),
	)

	r := application.Router()

	// 3. Global Middleware Stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(15 * time.Second))

	// 4. Register Routes (Native *core.Context handlers, zero wrappers required)
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	router.Version(r, "v1", func(v1 router.Router) {
		v1.Get("/hello", func(c *core.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Welcome to Bourbon!",
			})
		})
	})

	// 5. Start server with signal-based graceful shutdown
	log.Printf("Starting %s on %s:%d", cfg.App.Name, cfg.HTTP.Host, cfg.HTTP.Port)
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}
```

---

## 📊 Performance Benchmarks

Bourbon is benchmarked on Apple Silicon (M4, macOS) with zero external network overhead:

| Benchmark Target | Operations / sec | Latency | Memory / Op | Allocs / Op |
| :--- | :--- | :--- | :--- | :--- |
| **Router Static Matching** | ~4,180,000 ops/s | `323 ns/op` | 822 B/op | 10 allocs/op |
| **Router Parametric Matching** | ~2,825,000 ops/s | `361 ns/op` | 1,133 B/op | 11 allocs/op |
| **DI Container Singleton Resolution** | ~35,150,000 ops/s | `36 ns/op` | 8 B/op | 1 allocs/op |
| **Event Bus Sync Dispatch** | ~27,760,000 ops/s | `45 ns/op` | 72 B/op | 2 allocs/op |
| **Job Queue Enqueue/Dequeue** | ~7,300,000 ops/s | `157 ns/op` | 96 B/op | 2 allocs/op |
| **Validation Execution** | ~3,120,000 ops/s | `350 ns/op` | 53 B/op | 2 allocs/op |

Run the benchmarks locally:
```bash
go test -bench=. ./benchmarks
```

---

## 🛠️ Contrib Adapters

Bourbon maintains official adapters in `contrib/` for external cloud infrastructure and databases:

- **Databases**: [contrib/sqlite](contrib/sqlite), [contrib/postgres](contrib/postgres), [contrib/mysql](contrib/mysql)
- **Key-Value & Cache**: [contrib/redis](contrib/redis)
- **Object Storage**: [contrib/s3](contrib/s3)
- **Security & Tokens**: [contrib/jwt](contrib/jwt)
- **Telemetry & Logging**: [contrib/otel](contrib/otel), [contrib/zap](contrib/zap)

---

## 🤝 Contributing

Contributions to Bourbon are warmly welcomed. Please follow these guidelines:

1. Fork the repository and create your branch (`git checkout -b feature/my-feature`).
2. Run test suites locally (`go test ./...`).
3. Verify formatting using standard `gofmt`.
4. Open a pull request against `main`.

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
