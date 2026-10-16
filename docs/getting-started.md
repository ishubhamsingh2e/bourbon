# Getting Started with Bourbon

Welcome to **Bourbon**, a batteries-included, zero-dependency core Go web framework.

---

## 1. Installation

Install the Bourbon CLI tool:

```bash
go install github.com/ishubhamsingh2e/bourbon/cmd/bourbon@latest
```

Verify your installation:

```bash
bourbon --version
bourbon doctor
```

---

## 2. Creating a New Project

Bourbon provides 4 official presets and 2 structural layouts (`feature` vertical-slice or `layered` architecture):

```bash
# Create a REST API project (feature vertical-slice layout)
bourbon new my-api --preset=api --structure=feature

# Create a web application (clean layered layout)
bourbon new my-app --preset=web --structure=layered
```

Available presets:
- `api`: Lightweight JSON REST API with router, context binders, and CORS.
- `web`: Server-rendered HTML templates with session management and CSRF defense.
- `fullstack`: Hybrid web app with server-rendered views, API endpoints, sessions, and auth guards.
- `worker`: Headless background daemon with event bus and job queues.

---

## 3. Starting the Development Server

Navigate to your project directory and start the live-reload development server:

```bash
cd my-api
bourbon dev
```

The Bourbon dev server launches an in-memory reverse proxy at `http://localhost:3000`. During code edits, incoming browser requests are buffered until the Go binary finishes recompiling in the background.

---

## 4. Application Bootstrap (`main.go`)

Bourbon uses an explicit builder pattern in `main.go`—clear, standard Go without hidden reflection or magic:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/config"
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/middleware"
)

func main() {
	cfg := config.LoadFromEnv()

	// Initialize application shell implementing core.Application
	application := app.New(
		app.WithConfig(cfg),
	)

	r := application.Router()

	// 1. Global Middleware Stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))

	// 2. Health Endpoint with native *core.Context handler
	r.Get("/healthz", func(c *core.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// 3. Start server with signal-based graceful shutdown
	log.Printf("Starting %s on %s:%d", cfg.App.Name, cfg.HTTP.Host, cfg.HTTP.Port)
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}
```

---

## 5. Running Tests & Compiling

```bash
# Run unit and integration tests with race detection
bourbon test

# Compile an optimized production binary
bourbon build
```
