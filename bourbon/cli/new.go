package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createProjectWithDB(name, database string) {
	// Validate database choice
	validDatabases := map[string]bool{
		"sqlite":   true,
		"postgres": true,
		"mysql":    true,
	}

	if !validDatabases[database] {
		fmt.Printf("Error: Invalid database '%s'. Must be: sqlite, postgres, or mysql\n", database)
		return
	}

	fmt.Printf("🥃 Creating new Bourbon project: %s\n", name)
	fmt.Printf("📦 Database: %s\n", database)

	if err := os.MkdirAll(name, 0755); err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		return
	}

	appName := strings.ReplaceAll(name, "-", "")

	dirs := []string{
		"static/css",
		"static/js",
		"templates",
		"storage",
		"storage/logs",
		".bourbon",
		filepath.Join("apps", appName),
		filepath.Join("apps", appName, "migrations"),
	}
	for _, dir := range dirs {
		path := filepath.Join(name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("Error creating %s: %v\n", dir, err)
			return
		}
	}

	// Select driver import based on database
	var driverImport string
	switch database {
	case "sqlite":
		driverImport = `_ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/sqlite"`
	case "postgres":
		driverImport = `_ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/postgres"`
	case "mysql":
		driverImport = `_ "github.com/ishubhamsingh2e/bourbon/bourbon/drivers/mysql"`
	}

	// Select settings template based on database
	var settingsContent string
	switch database {
	case "sqlite":
		settingsContent = settingsTemplateSQLite
	case "postgres":
		settingsContent = settingsTemplatePostgres
	case "mysql":
		settingsContent = settingsTemplateMySQL
	}

	files := map[string]string{
		"main.go":                                mainTemplate,
		"middleware.go":                          middlewareTemplate,
		"settings.toml":                          settingsContent,
		"go.mod":                                 goModTemplate,
		".gitignore":                             gitignoreTemplate,
		"README.md":                              readmeTemplate,
		filepath.Join("templates", "index.html"): indexHTMLTemplate,
		filepath.Join("static", "css", "style.css"):                   cssTemplate,
		filepath.Join("storage", ".gitkeep"):                          "",
		filepath.Join("storage", "logs", ".gitkeep"):                  "",
		filepath.Join("apps", appName, "models.go"):                   appModelsTemplate,
		filepath.Join("apps", appName, "controllers.go"):              appControllersTemplate,
		filepath.Join("apps", appName, "routes.go"):                   appRoutesTemplate,
		filepath.Join("apps", appName, "migrations", "migrations.go"): migrationsPackageTemplate,
	}

	data := map[string]string{
		"ProjectName":  name,
		"ModulePath":   fmt.Sprintf("github.com/yourusername/%s", name),
		"AppName":      appName,
		"Database":     database,
		"DriverImport": driverImport,
	}

	for filename, templateStr := range files {
		filePath := filepath.Join(name, filename)
		content := renderTemplate(templateStr, data)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("Error creating %s: %v\n", filename, err)
			return
		}
	}

	fmt.Printf("\n✅ Project '%s' created successfully!\n\n", name)
	fmt.Println("📋 Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  go mod tidy                      # Install dependencies")
	fmt.Println("  go run . make:migration          # Create migrations")
	fmt.Println("  go run .                         # Start server")
	fmt.Println("\n🥃 Happy coding with Bourbon!")
}

func renderTemplate(tmpl string, data map[string]string) string {
	result := tmpl
	for key, value := range data {
		placeholder := "{{." + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

const mainTemplate = `package main

import (
	{{.DriverImport}}
	"{{.ModulePath}}/apps/{{.AppName}}"
	_ "{{.ModulePath}}/apps/{{.AppName}}/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

func main() {
	cmd.SetCustomInit(func(app *core.Application) error {
		SetupMiddleware(app)
		{{.AppName}}.RegisterRoutes(app, "/")
		return nil
	})
	cmd.Run("./settings.toml")
}
`

const settingsTemplateSQLite = `[app]
name = "{{.ProjectName}}"
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

[database.options]
log_queries = false

# Middleware configuration
# Middlewares are registered in middleware.go and enabled here
# They are applied in the order listed below
[middleware]
enabled = [
    "recovery",  # Must be first - handles panics
    "logger",    # Request/response logging
    # "cors",    # Uncomment to enable CORS
    # "custom",  # Your custom middleware from middleware.go
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
`

const settingsTemplatePostgres = `[app]
name = "{{.ProjectName}}"
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
driver = "postgres"
host = "localhost"
port = 5432
name = "{{.ProjectName}}_db"
user = "postgres"
password = "postgres"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = 3600

[database.options]
ssl_mode = "disable"
log_queries = false

# Middleware configuration
# Middlewares are registered in middleware.go and enabled here
# They are applied in the order listed below
[middleware]
enabled = [
    "recovery",  # Must be first - handles panics
    "logger",    # Request/response logging
    # "cors",    # Uncomment to enable CORS
    # "custom",  # Your custom middleware from middleware.go
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
`

const settingsTemplateMySQL = `[app]
name = "{{.ProjectName}}"
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
driver = "mysql"
host = "localhost"
port = 3306
name = "{{.ProjectName}}_db"
user = "root"
password = "root"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = 3600

[database.options]
charset = "utf8mb4"
parse_time = "true"
loc = "Local"
log_queries = false

# Middleware configuration
# Middlewares are registered in middleware.go and enabled here
# They are applied in the order listed below
[middleware]
enabled = [
    "recovery",  # Must be first - handles panics
    "logger",    # Request/response logging
    # "cors",    # Uncomment to enable CORS
    # "custom",  # Your custom middleware from middleware.go
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
`

const indexHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <div class="container">
        <h1>{{.AppName}}</h1>
        <p>{{.Message}}</p>
        <div class="info">
            <p>* Template engine is working!</p>
            <p>* Static files are served from <code>/static</code></p>
            <p>* SQLite database ready to use</p>
            <p>Check out <a href="/api/health">/api/health</a> for API status</p>
        </div>
    </div>
</body>
</html>
`

const goModTemplate = `module {{.ModulePath}}

go 1.21

require (
	github.com/ishubhamsingh2e/bourbon v1.0.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

// LOCAL DEVELOPMENT: Uncomment the line below and fix the path
// Once Bourbon is published to GitHub, you can remove this line
replace github.com/ishubhamsingh2e/bourbon => /Volumes/External/Git/Bourbon
`

const gitignoreTemplate = `# Binaries
*.exe
*.dll
*.so
*.dylib
main
{{.ProjectName}}

# Test files
*.test
*.out

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Application
*.log
storage/database.db
storage/logs/

# Bourbon state (local development)
.bourbon/
`

const readmeTemplate = `# {{.ProjectName}}

A Bourbon web application.

## Getting Started

### Install Dependencies

` + "```bash" + `
go mod tidy
` + "```" + `

### Run Server

` + "```bash" + `
go run .
` + "```" + `

Your app will be running at http://localhost:8000

## Available Commands

### Migration Commands

` + "```bash" + `
# Create a new migration with a name
go run . make:migration create_users_table

# Create a new migration with just timestamp (no name)
go run . make:migration

# Run pending migrations (manual - not automatic)
go run . migrate

# Show migration status
go run . migrate:status

# Rollback last migration
go run . migrate:rollback
` + "```" + `

## Customization

The default ` + "`main.go`" + ` shows clear app registration with URL prefixes:

` + "```go" + `
package main

import (
	"{{.ModulePath}}/apps/{{.AppName}}"
	_ "{{.ModulePath}}/apps/{{.AppName}}/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

func main() {
	cmd.SetCustomInit(func(app *core.Application) error {
		// Setup custom middleware from middleware.go
		SetupMiddleware(app)
		
		// Register app routes under their URL prefixes
		{{.AppName}}.RegisterRoutes(app, "/")
		return nil
	})
	cmd.Run("./settings.toml")
}
` + "```" + `

### Middleware Configuration

Middleware are registered in ` + "`middleware.go`" + ` and enabled in ` + "`settings.toml`" + `:

` + "```toml" + `
[middleware]
enabled = [
    "recovery",  # Must be first
    "logger",
    "cors",
    "custom",    # Your custom middleware
]
` + "```" + `

In ` + "`middleware.go`" + `:

` + "```go" + `
func SetupMiddleware(app *core.Application) {
	// Register built-in middleware
	app.RegisterMiddleware("recovery", middleware.Recovery(app.Logger, app.ErrorStore))
	app.RegisterMiddleware("logger", middleware.Logger(app.Logger, app.ErrorStore))
	
	// Register custom middleware
	app.RegisterMiddleware("custom", MyCustomMiddleware())
	
	// Load from config
	for _, name := range app.Config.Middleware.Enabled {
		app.UseMiddleware(name)
	}
}
` + "```" + `

### Route Grouping (Django-style URL Patterns)

Each app can be mounted at a different URL prefix:

` + "```go" + `
package main

import (
	"myproject/apps/users"
	"myproject/apps/api"
	_ "myproject/apps/users/migrations"
	_ "myproject/apps/api/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

func main() {
	cmd.SetCustomInit(func(app *core.Application) error {
		SetupMiddleware(app)
		
		// Mount apps at different URL prefixes
		users.RegisterRoutes(app, "/")        // Root URL
		api.RegisterRoutes(app, "/api")       // /api/...
		// admin.RegisterRoutes(app, "/admin") // /admin/...
		
		return nil
	})
	cmd.Run("./settings.toml")
}
` + "```" + `

In your app's ` + "`routes.go`" + `:

` + "```go" + `
func RegisterRoutes(app *core.Application, prefix string) {
	group := app.Router.Group(prefix)
	
	group.Get("/items", listItemsHandler)       // /api/items
	group.Post("/items", createItemHandler)     // /api/items
	group.Get("/items/:id", getItemHandler)     // /api/items/123
}
` + "```" + `

### Adding Custom Routes

` + "```go" + `
package main

import (
	"{{.ModulePath}}/apps/{{.AppName}}"
	_ "{{.ModulePath}}/apps/{{.AppName}}/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	bourbonHttp "github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

func main() {
	cmd.SetCustomInit(func(app *core.Application) error {
		SetupMiddleware(app)
		
		// Register your app routes
		{{.AppName}}.RegisterRoutes(app, "/")
		
		// Add additional custom routes
		app.Router.Get("/hello", func(ctx *bourbonHttp.Context) error {
			return ctx.String(200, "Hello World!")
		})
		return nil
	})

	cmd.Run("./settings.toml")
}
` + "```" + `

### Adding Custom Commands

` + "```go" + `
package main

import (
	"fmt"
	"{{.ModulePath}}/apps/{{.AppName}}"
	_ "{{.ModulePath}}/apps/{{.AppName}}/migrations"
	"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

func init() {
	// Register a database seed command
	cmd.RegisterCommand("seed", func(args []string) error {
		app := core.NewApplication("./settings.toml")
		if err := app.ConnectDB(); err != nil {
			return err
		}
		fmt.Println("Seeding database...")
		// Your seeding logic
		return nil
	})
}

func main() {
	cmd.SetCustomInit(func(app *core.Application) error {
		SetupMiddleware(app)
		{{.AppName}}.RegisterRoutes(app, "/")
		return nil
	})
	cmd.Run("./settings.toml")
}
` + "```" + `

Then run: ` + "`go run main.go seed`" + `

### Full Control

For complete control over the startup process:

` + "```go" + `
package main

import (
"os"
_ "{{.ModulePath}}/database/migrations"
"github.com/ishubhamsingh2e/bourbon/bourbon/cmd"
"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

func main() {
// Handle CLI commands
if len(os.Args) > 1 {
cmd.HandleCommand(os.Args[1:])
return
}

// Manual server setup
app := core.NewApplication("./settings.toml")

// Custom middleware configuration
app.RegisterMiddleware("custom", myMiddleware)
app.UseMiddleware("custom")

// Setup default middlewares
cmd.SetupDefaultMiddlewares(app)

// Database connection
if err := app.ConnectDB(); err != nil {
app.Logger.Fatal("DB connection failed")
}

// Your custom logic here
setupRoutes(app)

// Start server
app.Run()
}
` + "```" + `

## Database

### Default Setup (SQLite)

By default, SQLite is configured in ` + "`settings.toml`" + `:

` + "```toml" + `
[database]
driver = "sqlite"
path = "storage/database.db"
` + "```" + `

The database will be created automatically in ` + "`storage/database.db`" + `.

### Migrations

Migrations are manual - run them when you're ready (not automatic on server startup):

1. **Create a migration:**

` + "```bash" + `
# With a descriptive name
go run . make:migration create_users_table

# Or just timestamp (no name)
go run . make:migration
` + "```" + `

This creates a new migration file in ` + "`apps/{{.AppName}}/migrations/`" + `.

2. **Edit the migration file** to add your schema changes:

` + "```go" + `
func init() {
core.RegisterGormigrateMigration(&gormigrate.Migration{
ID: "20260215215006_create_users_table",
Migrate: func(db *gorm.DB) error {
type User struct {
ID        uint   ` + "`gorm:\"primaryKey\"`" + `
Email     string ` + "`gorm:\"unique;not null\"`" + `
Name      string
CreatedAt time.Time
}
return db.AutoMigrate(&User{})
},
Rollback: func(db *gorm.DB) error {
return db.Migrator().DropTable("users")
},
})
}
` + "```" + `

3. **Run migrations:**

` + "```bash" + `
go run main.go migrate
` + "```" + `

4. **Check status:**

` + "```bash" + `
go run main.go migrate:status
` + "```" + `

### Switch to PostgreSQL

1. Add PostgreSQL driver:

` + "```bash" + `
go get gorm.io/driver/postgres
` + "```" + `

2. Update ` + "`settings.toml`" + `:

` + "```toml" + `
[database]
driver = "postgres"
host = "localhost"
port = 5432
name = "{{.ProjectName}}_db"
user = "dbuser"
password = "dbpass"
` + "```" + `

## Project Structure

` + "```" + `
{{.ProjectName}}/
├── main.go                    # Application entry point (clean & minimal)
├── settings.toml              # Configuration file
├── apps/                      # Your application modules
│   └── {{.AppName}}/          # Default app
│       ├── models.go          # Data models
│       ├── controllers.go     # Request handlers
│       ├── routes.go          # URL routing
│       └── migrations/        # App-specific migrations
├── templates/                 # HTML templates
├── static/                    # Static files (CSS, JS, images)
└── storage/                   # Database and logs
` + "```" + `

## License

MIT
`

const cssTemplate = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    line-height: 1.6;
    color: #333;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
}

.container {
    background: white;
    padding: 40px;
    border-radius: 10px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.2);
    max-width: 600px;
    text-align: center;
}

h1 {
    color: #8B4513;
    margin-bottom: 20px;
    font-size: 2.5em;
}

p {
    color: #666;
    font-size: 1.1em;
    margin-bottom: 15px;
}

.info {
    margin-top: 30px;
    padding: 20px;
    background: #f8f9fa;
    border-radius: 8px;
    text-align: left;
}

.info p {
    margin-bottom: 10px;
}

code {
    background: #e9ecef;
    padding: 2px 8px;
    border-radius: 4px;
    font-family: 'Courier New', monospace;
}

a {
    color: #667eea;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}
`

const migrationsPackageTemplate = `package migrations

// Migrations variable is used to ensure this package is imported
// All migration files in this directory will auto-register via init()
var Migrations = "migrations"
`

const appModelsTemplate = `package {{.AppName}}

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/database/orm"
)

// User model - example of a basic model
// Remove or modify this based on your needs
type User struct {
	orm.BaseModel
	Name  string ` + "`json:\"name\" gorm:\"not null\"`" + `
	Email string ` + "`json:\"email\" gorm:\"uniqueIndex;not null\"`" + `
}
`

const appControllersTemplate = `package {{.AppName}}

import (
"net/http"
"github.com/ishubhamsingh2e/bourbon/bourbon/core"
bourbonHttp "github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

type HomeController struct {
App *core.Application
}

func NewHomeController(app *core.Application) *HomeController {
return &HomeController{App: app}
}

func (c *HomeController) Index(ctx *bourbonHttp.Context) error {
data := bourbonHttp.H{
"Title":   "Welcome to {{.ProjectName}}",
"AppName": "{{.ProjectName}}",
"Message": "Your Bourbon application is running!",
}
return ctx.Render("index.html", data)
}

func (c *HomeController) HealthCheck(ctx *bourbonHttp.Context) error {
return ctx.JSON(http.StatusOK, bourbonHttp.H{
"status": "healthy",
"app":    c.App.Config.App.Name,
})
}
`

const appRoutesTemplate = `package {{.AppName}}

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
)

// RegisterRoutes registers all routes for this app under the given prefix
// prefix examples: "/", "/api", "/admin", etc.
func RegisterRoutes(app *core.Application, prefix string) {
	homeCtrl := NewHomeController(app)
	
	// Create a route group for this app
	group := app.Router.Group(prefix)
	
	// Register routes within the group
	group.Get("/", homeCtrl.Index)
	group.Get("/health", homeCtrl.HealthCheck)
}
`

const middlewareTemplate = `package main

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/bourbon/middleware"
)


func SetupMiddleware(app *core.Application) {
	// Register built-in middleware
	app.RegisterMiddleware("recovery", middleware.Recovery(app.Logger, app.ErrorStore))
	app.RegisterMiddleware("logger", middleware.Logger(app.Logger, app.ErrorStore))
	
	// CORS middleware - configure based on your needs
	corsOrigin := "*"
	if len(app.Config.Security.CorsOrigins) > 0 {
		corsOrigin = app.Config.Security.CorsOrigins[0]
	}
	app.RegisterMiddleware("cors", middleware.CORS(corsOrigin))
	
	// Register your custom middleware here
	// Example:
	// app.RegisterMiddleware("custom", MyCustomMiddleware())
	
	// Load middleware based on settings.toml configuration
	// Middleware are applied in the order listed in settings.toml
	for _, name := range app.Config.Middleware.Enabled {
		if err := app.UseMiddleware(name); err != nil {
			app.Logger.Warn("Failed to load middleware: " + name)
		}
	}
}
`
