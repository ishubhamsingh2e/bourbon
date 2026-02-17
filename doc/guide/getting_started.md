# Getting Started with Bourbon

## Prerequisites

- Go 1.21 or higher
- Git

## Installation

### 1. Install the CLI Tool

Clone the Bourbon repository and build the CLI tool:

```bash
git clone https://github.com/ishubhamsingh2e/bourbon
cd bourbon
go build -tags=all_drivers -o bourbon-cli ./cmd/bourbon
```

Optionally, move the `bourbon-cli` to a directory in your PATH:

```bash
sudo mv bourbon-cli /usr/local/bin/bourbon
```

Now you can use the `bourbon` command globally.

### 2. Create a New Project

To create a new Bourbon project, run:

```bash
bourbon new myblog
```

This creates a new directory `myblog` with the following structure:

```
myblog/
├── apps/                # Application modules
│   └── myblog/          # Default app
├── static/              # Static files (CSS, JS, images)
├── storage/             # Database and logs
├── templates/           # HTML templates
├── db.go                # Database setup
├── go.mod               # Dependencies
├── main.go              # Application entry point
└── settings.toml        # Configuration
```

### 3. Initialize Dependencies

Navigate into your project directory and tidy up dependencies:

```bash
cd myblog
go mod tidy
```

### 4. Configure Database

Bourbon uses SQLite by default, which requires no extra setup. If you want to use PostgreSQL or MySQL, update `settings.toml`.

### 5. Create Your First App

A default app is created for you, but you can create additional apps using:

```bash
bourbon create:app posts
```

This creates a new directory structure inside `apps/posts` with `models.go`, `controllers.go`, and `routes.go`.

**Important:** Don't forget to register your new app in `settings.toml` under `[apps.installed]`:

```toml
[apps]
installed = [
    "myblog",
    "posts"
]
```

### 6. Define Models

Open `apps/posts/models.go` and define your data models using standard Go structs with GORM tags:

```go
package posts

import (
    "github.com/ishubhamsingh2e/bourbon/bourbon/models"
)

type Post struct {
    models.BaseModel
    Title   string `gorm:"size:255" json:"title"`
    Content string `gorm:"type:text" json:"content"`
}
```

### 7. Make Migrations

Bourbon automatically detects changes in your models and generates migration files:

```bash
bourbon make:migration
```

This will create a new migration file in `apps/posts/migrations/`.

### 8. Run Migrations

To apply the migrations, you run your application. Ensure that migration running is enabled in `main.go` or `db.go`. By default, `main.go` includes a call to run migrations:

```go
// main.go
if err := core.RunMigrations(app); err != nil {
    app.Logger.Error("Migration failed", zap.Error(err))
    os.Exit(1)
}
```

Simply run your app:

```bash
go run main.go
```

The migrations will be applied automatically on startup.

### 9. Start the Server

```bash
go run main.go
```

Your server is now running at `http://localhost:8000`.

Visit `http://localhost:8000/api/health` to verify the API status.
