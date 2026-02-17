# 🥃 Bourbon Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/ishubhamsingh2e/bourbon)](https://goreportcard.com/report/github.com/ishubhamsingh2e/bourbon)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A Django-inspired MVC framework for Go with built-in ORM, smart migrations, and CLI scaffolding.**

Bourbon brings the elegance and productivity of Django to the Go ecosystem—providing a structured, batteries-included environment with sensible defaults and minimal boilerplate.

---

## 📚 Documentation

- **[Architecture Overview](doc/ARCHITECTURE.md)** - Detailed framework architecture and components
- **[Quick Reference Diagram](doc/DIAGRAM.md)** - Visual guide to request flow and package structure

---

## Features

- **Django-inspired Architecture** - Familiar App-based modular structure.
- **Built-in ORM** - Seamless [GORM](https://gorm.io) integration supporting PostgreSQL, MySQL, and SQLite.
- **Smart Migrations** - Auto-detects model changes and generates Go-based migrations (like `makemigrations`).
- **Robust Router** - RESTful routing with grouping, path parameters (`:id`), and middleware support.
- **Template Engine** - Powered by Go `html/template` with auto-reload and custom functions.
- **Structured Logging** - High-performance logging via Uber Zap with file rotation and error storage.
- **CLI Scaffolding** - Quick generation of projects, apps, and migrations.
- **Async Jobs** - Built-in async dispatcher interface for background task processing.
- **Middleware System** - Named middleware registry with per-route and global application.
- **Error Storage** - Automatic panic and 5xx error capture to database for debugging.
- **SQLite by Default** - Zero-config start with no external database required.

---

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/ishubhamsingh2e/bourbon
cd bourbon

# Build and install the CLI
go build -o bourbon ./cmd/bourbon
sudo mv bourbon /usr/local/bin/
```

### 2. Create a Project

```bash
# Create a new project with SQLite (default)
bourbon new myblog
cd myblog

# Or choose your database (sqlite, postgres, mysql)
bourbon new myblog --db=postgres

# Install dependencies
go mod tidy

# Start the server
go run .
```

Your application is now running at http://localhost:8000!

### 3. Clean & Minimal Code

The generated `main.go` is intentionally minimal - all boilerplate is abstracted into the framework:

```go
package main

import (
	_ "myblog/database/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
)

func main() {
	cmd.Run("./settings.toml")
}
```

**That's it!** This single line:
- ✅ Handles CLI commands (migrate, make:migration, etc.)
- ✅ Initializes the application
- ✅ Sets up default middlewares
- ✅ Connects to database
- ✅ Runs migrations
- ✅ Starts the HTTP server

**Want customization?** It's fully overridable! See [ABSTRACTION_GUIDE.md](ABSTRACTION_GUIDE.md) for details.

---

## Documentation

Comprehensive documentation is available in the [doc/](doc/index.md) directory:

- **[Getting Started](doc/guide/getting_started.md)** - Your first 5 minutes with Bourbon.
- **[Directory Structure](doc/guide/directory_structure.md)** - Understanding the project layout.
- **[Routing & Middleware](doc/core/routing.md)** - Defining endpoints and request pipelines.
- **[Models & Migrations](doc/database/models.md)** - Managing your data layer.
- **[CLI Reference](doc/cli/reference.md)** - Full list of available commands.
- **[Deployment](doc/deployment/deployment.md)** - Moving your app to production.
- **[API Reference](doc/API_REFERENCE.md)** - Quick lookup for common APIs and patterns.

## Roadmap

Bourbon is actively developed with the goal of achieving feature parity with Django, Laravel, and Spring Boot. See our comprehensive **[TODO.md](TODO.md)** for:

- 📋 **6 Development Phases** - From core stability to enterprise features
- 🎯 **Feature Comparison Matrix** - Track progress against Django, Laravel, and Spring Boot
- 🎯 **Priority Levels** - High, medium, and future enhancements
- 🤝 **Contribution Opportunities** - Find features to work on

**Current Focus**: Phase 1 - Authentication, validation, sessions, and security

---

## Example

Bourbon makes it easy to define clean, readable APIs:

```go
func RegisterRoutes(app *core.App) {
    // Basic route
    app.Router.Get("/", func(c *http.Context) error {
        return c.Render("index.html", http.H{"Title": "Home"})
    })

    // Grouped API routes with middleware
    api := app.Router.Group("/api/v1")
    {
        api.Get("/posts", listPosts)
        api.Post("/posts", createPost)
    }
}
```

---

## The Migration Workflow

Bourbon handles database schema changes just like Django. When you update your `models.go`:

```bash
# 1. Detect changes and generate migration files
bourbon make:migration --app=posts --name=add_author_field

# 2. Run the application
go run main.go
```

Migrations are automatically applied on startup, ensuring your database is always in sync with your code.

---

## Project Structure

```text
myproject/
├── apps/                # Your modular application logic
│   └── blog/
│       ├── models.go    # Data models
│       ├── routes.go    # App-specific routes
│       └── migrations/  # Generated migrations
├── static/              # CSS, JS, Images
├── templates/           # HTML Templates
├── storage/             # SQLite DB and Logs
├── settings.toml        # Unified configuration
└── main.go              # App entry point
```

---

## Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are greatly appreciated.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Acknowledgments

- Django for the timeless architectural inspiration.
- GORM for the powerful ORM capabilities.
- Cobra for the CLI framework.

---

**Bourbon v0.0.1** - Made for Go developers who love Django's philosophy.
