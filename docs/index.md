# Bourbon Documentation Portal

Welcome to the official documentation for **Bourbon**, a batteries-included web framework for Go designed to deliver the developer velocity and elegance of Django and Laravel while strictly upholding Go's standard library philosophy: explicitness, zero unnecessary abstractions, static typing, low reflection, and sub-millisecond startup times.

---

## Documentation Map

### 1. Foundations & Core Concepts
- **[Getting Started](getting-started.md)**: Installation, CLI verification, creating projects from presets, and booting your first application.
- **[Architecture & Design Philosophy](architecture.md)**: Monorepo layout, zero-dependency core principles, port-and-adapter architecture, and application lifecycles.
- **[Configuration & Dependency Injection](config-and-di.md)**: Environment variable loading, typed configuration structures, service providers, and the type-safe DI container.

### 2. HTTP Engine & Request Pipelines
- **[Routing & Middleware](routing-and-middleware.md)**: Go 1.22+ ServeMux routing, route parameters, grouping, API versioning, REST resource definitions, and the built-in middleware stack.
- **[Validation & Error Handling](validation.md)**: Declarative struct validation, rule builders, field-level errors, and RFC 7807 Problem Details serialization.

### 3. Data & State Management
- **[Database, Models & Migrations](database-and-orm.md)**: Pluggable `database.DB` port, unit-of-work transactions (`WithTx`), dialects (SQLite, Postgres, MySQL), model metadata conventions, and SQL migration management.
- **[Authentication & RBAC](authentication-and-rbac.md)**: Multi-guard authentication (`SessionGuard`, `TokenGuard`), role-to-permission inheritance, wildcard access rules, and password hashing.

### 4. View & Frontend Layer
- **[Server-Rendered Templating](templating.md)**: Server-rendered HTML templates with layout composition, partials, template helpers (`csrf_token`, `auth`, `flash`), and dev-mode hot reloading.

### 5. Asynchronous Processing & Real-Time Hub
- **[Events, Background Jobs & Real-Time](events-and-jobs.md)**: In-process event bus, background job queues with retry backoffs, cron task scheduling, and Server-Sent Events (SSE) broadcasting.

### 6. Infrastructure & Operations
- **[Storage, Mailers & Observability](storage-and-mail.md)**: Local and cloud file storage abstractions, SMTP and development mailers, health probes, Prometheus metrics, and dynamic OpenAPI 3.0 specs.
- **[Observability & OpenAPI](observability-and-openapi.md)**: Structured `slog` integration, liveness/readiness probes, metrics collectors, and OpenAPI documentation endpoints.

### 7. Developer Tooling & Testing
- **[CLI Reference](cli-reference.md)**: Complete guide to `bourbon new`, `bourbon dev`, `bourbon make:*`, `bourbon routes`, and `bourbon doctor`.
- **[Testing & Contract Verification](testing.md)**: Unit testing with fluent test clients (`testutil`), the in-memory fake database (`memdb`), and adapter compliance contract suites.

---

## Architectural Principles at a Glance

Bourbon is constructed on four primary architectural tenets:

1. **Zero-Dependency Core**: The core module (`github.com/ishubhamsingh2e/bourbon`) relies exclusively on the Go standard library and Go-team maintained `golang.org/x/*` packages. Third-party drivers and cloud SDKs are decoupled into the `contrib/` ecosystem.
2. **Explicit Builder Pattern**: No hidden reflection or global state magic. Applications explicitly register configuration and service providers in `main.go`.
3. **Convention with Explicit Overridability**: Standard directory structures (`feature` and `layered` presets) are codified in a root `bourbon.yaml` manifest that developers can modify without framework resistance.
4. **Standard Library Ergonomics**: Standard `net/http` handlers, Go 1.22+ routing patterns, standard `html/template` layouts, and standard `log/slog` logging formatters.
