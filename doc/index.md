# Bourbon Documentation

Welcome to the documentation for **Bourbon**, a Django-inspired MVC framework for Go.

## Overview

Bourbon brings the elegance and structure of Django to the Go programming language. It provides a robust set of tools for building modern web applications, including:

- **MVC Architecture:** Organized structure with Apps, Models, Views (Controllers), and Templates.
- **Built-in ORM:** Seamless integration with GORM for database operations.
- **Smart Migrations:** Auto-detects model changes and generates migration files, similar to Django's `makemigrations`.
- **Powerful Routing:** Expressive HTTP routing with path parameters, grouping, and middleware support.
- **Template Engine:** Go html/template with auto-reload and custom function registration.
- **Async Jobs:** Background task processing with pluggable dispatcher backends.
- **Error Storage:** Automatic panic and 5xx error capture to database.
- **CLI Tools:** Scaffolding for projects, apps, and migrations.

## Getting Started

If you are new to Bourbon, start with the [Getting Started](guide/getting_started.md) guide to create your first project.

## Core Concepts

- **[Routing](core/routing.md):** Learn how to define URL patterns and handle requests.
- **[Requests & Responses](core/requests_responses.md):** Dive into the `Context` object, data binding, and response formats.
- **[Middleware](core/middleware.md):** Understand how to intercept and process requests globally or per-route.
- **[Templates & Static Files](core/templates_static.md):** Learn how to serve HTML and static assets.
- **[Async Jobs](core/async_jobs.md):** Process background tasks with the async dispatcher system.

## Database

- **[Models](database/models.md):** Define your data structure using GORM structs.
- **[Migrations](database/migrations.md):** Manage database schema changes with auto-detection.

## CLI Reference

Check out the [CLI Reference](cli/reference.md) for a complete list of commands including:
- Project scaffolding (`bourbon new`)
- App generation (`bourbon create:app`)
- Migration management (`make:migration`, `migrate`, `migrate:rollback`)

## Quick Reference

- **[API Reference](API_REFERENCE.md)** - Quick lookup for common Bourbon APIs and patterns
- **[Changelog](../CHANGELOG.md)** - Version history and recent changes

## Configuration

Learn about the [Configuration System](guide/configuration.md) and how to customize your application via `settings.toml`.

## Architecture

For a deep dive into the framework internals, see:
- **[Architecture Overview](../ARCHITECTURE.md)** - Framework components and design patterns
- **[Quick Reference Diagram](../DIAGRAM.md)** - Visual guide to request flow

## Contributing

Bourbon is an open-source project. We welcome contributions, bug reports, and feature requests.
