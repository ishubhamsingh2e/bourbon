# Observability & OpenAPI Specification in Bourbon

Bourbon provides zero-dependency production observability primitives, including Kubernetes-ready liveness/readiness health probes, Prometheus-compatible metrics exposition, structured logging adapters, and dynamic OpenAPI 3.0 schema generation.

---

## 1. Health Probes (`/healthz` and `/readyz`)

The `observability.HealthChecker` provides standard endpoints for orchestrators like Kubernetes or Docker Swarm:

```go
package routes

import (
	"context"
	"database/sql"

	"github.com/ishubhamsingh2e/bourbon/observability"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func RegisterHealthRoutes(r router.Router, db *sql.DB) {
	checker := observability.NewHealthChecker()

	// Register readiness dependency checks
	checker.AddCheck("database", func(ctx context.Context) error {
		return db.PingContext(ctx)
	})

	// /healthz returns 200 OK {"status":"up"} immediately
	r.Get("/healthz", checker.Liveness())

	// /readyz runs all checks; returns 200 OK or 503 Service Unavailable
	r.Get("/readyz", checker.Readiness())
}
```

---

## 2. Prometheus Metrics (`/metrics`)

Bourbon includes an in-core Prometheus metrics collector (`observability.MetricsCollector`) that tracks total requests, durations, and per-route HTTP status codes:

```go
package main

import (
	"github.com/ishubhamsingh2e/bourbon/observability"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func SetupMetrics(r router.Router) {
	collector := observability.NewMetricsCollector()

	// Apply metrics recording middleware globally
	r.Use(collector.Middleware())

	// Expose standard Prometheus text-format scrape endpoint
	r.Get("/metrics", collector.Handler())
}
```

Scrape output example:
```text
# HELP http_requests_total Total number of HTTP requests processed
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/users",status="200"} 42
# HELP http_request_duration_seconds Summary of request processing latency
# TYPE http_request_duration_seconds summary
http_request_duration_seconds_sum 0.0423
http_request_duration_seconds_count 42
```

---

## 3. Dynamic OpenAPI 3.0 Generation

Bourbon allows programmatically constructing an OpenAPI 3.0 specification from registered routes and exporting it as JSON or Swagger UI:

```go
package docs

import (
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/openapi"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func RegisterOpenAPIRoutes(r router.Router) {
	doc := openapi.NewBuilder().
		SetTitle("Bourbon Todo API").
		SetVersion("1.0.0").
		SetDescription("Production REST API built with Bourbon").
		AddRoute("GET", "/api/v1/todos", "List all todos", "200", "Returns array of todo items").
		AddRoute("POST", "/api/v1/todos", "Create new todo", "201", "Returns created todo item")

	// Serves raw OpenAPI 3.0 JSON specification
	r.Get("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		doc.WriteJSON(w)
	})

	// Serves embedded Swagger UI documentation page
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(doc.SwaggerUIHTML("/openapi.json")))
	})
}
```
