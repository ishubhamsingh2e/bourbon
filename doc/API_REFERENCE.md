# API Reference

Quick reference for commonly used Bourbon framework APIs.

## Application (core.Application)

### Initialization
```go
app := core.NewApplication(config)
```

### Router Methods
```go
app.Router.Get(pattern, handler)
app.Router.Post(pattern, handler)
app.Router.Put(pattern, handler)
app.Router.Patch(pattern, handler)
app.Router.Delete(pattern, handler)
```

### Route Groups
```go
group := app.Router.Group(prefix, middleware...)
group.Get(pattern, handler)
```

### Middleware
```go
// Register named middleware
app.RegisterMiddleware(name, middleware)

// Apply middleware
app.UseMiddleware(name)

// Apply multiple at once
app.Use(middleware1, middleware2)
```

### Static Files
```go
app.Static(urlPrefix, directory)
```

---

## Context (http.Context)

### Request Data
```go
// URL Parameters
id := c.Param("id")

// Query Parameters
page := c.Query("page")
pageWithDefault := c.QueryDefault("page", "1")

// Form Data
name := c.FormValue("name")

// Headers
authHeader := c.GetHeader("Authorization")
```

### Data Binding
```go
// JSON
var user User
c.BindJSON(&user)

// Form
var form LoginForm
c.BindForm(&form)

// Query string
var filters Filters
c.BindQuery(&filters)
```

### Responses
```go
// JSON response
c.JSON(200, data)

// String response
c.String(200, "Hello, World!")

// HTML rendering
c.Render("template.html", data)
c.RenderWithStatus(404, "error.html", data)

// Redirect
c.Redirect(302, "/login")
```

### Async Jobs
```go
// Dispatch job
jobID, err := c.DispatchAsync(handler, payload)

// Quick JSON response
c.DispatchAsyncJSON(handler, payload)

// Get result
result, err := c.GetAsyncResult(jobID)
```

### Context Storage
```go
// Set value
c.Set("key", value)

// Get value
value := c.Get("key")

// Get with type assertion
user := c.Get("user").(*User)
```

---

## Models (database/models)

### BaseModel
```go
type User struct {
    models.BaseModel  // Includes ID, CreatedAt, UpdatedAt, DeletedAt
    Email    string   `gorm:"uniqueIndex"`
    Username string   `gorm:"not null"`
}
```

### Common GORM Operations
```go
// Create
app.DB.Create(&user)

// Find by ID
var user User
app.DB.First(&user, id)

// Find all
var users []User
app.DB.Find(&users)

// Update
app.DB.Model(&user).Update("email", "new@email.com")

// Delete (soft delete with BaseModel)
app.DB.Delete(&user, id)
```

---

## Middleware

### Built-in Middleware

#### Logger
```go
middleware.Logger(logger, errorStore)
```
Logs HTTP requests with colored output, status codes, and timing.

#### Recovery
```go
middleware.Recovery(logger, errorStore)
```
Recovers from panics, logs stack traces, and stores errors in database.

#### CORS
```go
middleware.CORS(origin)
```
Handles Cross-Origin Resource Sharing with configurable origins.

### Custom Middleware
```go
func MyMiddleware() http.MiddlewareFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(c *http.Context) error {
            // Before request
            
            err := next(c)
            
            // After request
            
            return err
        }
    }
}
```

---

## Configuration (settings.toml)

### App Configuration
```toml
[app]
name = "myapp"
env = "development"  # development, staging, production
debug = true
secret_key = "your-secret-key"
timezone = "UTC"
```

### Server Configuration
```toml
[server]
host = "localhost"
port = 8000
read_timeout = 30    # seconds
write_timeout = 30   # seconds
```

### Database Configuration
```toml
[database]
driver = "sqlite"    # sqlite, postgres, mysql
path = "storage/app.db"  # SQLite path

# For PostgreSQL/MySQL
# host = "localhost"
# port = 5432
# database = "myapp"
# user = "postgres"
# password = "secret"
# ssl_mode = "disable"
```

### Middleware Configuration
```toml
[middleware]
enabled = ["recovery", "logger", "cors"]
```

### Template Configuration
```toml
[templates]
directory = "templates"
extension = ".html"
auto_reload = true  # Reload templates on each request (development)
```

### Static Files Configuration
```toml
[static]
directory = "static"
url_prefix = "/static"
```

### Logging Configuration
```toml
[logging]
level = "info"  # debug, info, warn, error
file_logging = true
storage_path = "storage/logs"
rotation = "daily"  # hourly, daily, weekly
store_errors_db = true  # Store 5xx errors in database
```

### Security Configuration
```toml
[security]
cors_origins = ["http://localhost:3000"]
csrf_enabled = false
```

---

## CLI Commands

### Project Creation
```bash
bourbon new myproject
bourbon create:app posts
```

### Migrations
```bash
go run . make:migration --app=posts --name=add_author
go run . migrate
go run . migrate:status
go run . migrate:rollback
go run . migrate:rollback --to=20260101120000
```

### Development
```bash
go run .              # Start server
go run main.go        # Alternative (CLI commands only)
```

---

## Error Handling

### Returning Errors
```go
func myHandler(c *http.Context) error {
    if err != nil {
        return c.JSON(400, http.H{"error": err.Error()})
    }
    return c.JSON(200, http.H{"success": true})
}
```

### Error Storage
```go
// Get recent errors
errors := app.ErrorStore.GetRecent(10)

// Get by status
errors := app.ErrorStore.GetByStatus(500)

// Get server errors
errors := app.ErrorStore.GetServerErrors()

// Clean old errors
app.ErrorStore.Clean(24 * time.Hour)  // Delete errors older than 24 hours
```

---

## Template Functions

### Adding Custom Functions
```go
app.TemplateEngine.AddFunc("toUpper", strings.ToUpper)

app.TemplateEngine.AddFuncs(template.FuncMap{
    "formatDate": formatDateFunc,
    "truncate":   truncateFunc,
})
```

### Using in Templates
```html
<h1>{{ .Title | toUpper }}</h1>
<p>{{ .Content | truncate 100 }}</p>
```

---

## Testing Patterns

### Handler Testing
```go
func TestMyHandler(t *testing.T) {
    app := core.NewApplication(testConfig)
    
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    
    c := http.NewContext(w, req, app)
    
    err := myHandler(c)
    assert.NoError(t, err)
    assert.Equal(t, 200, w.Code)
}
```

---

## Best Practices

1. **Always use BaseModel** for timestamps and soft deletes
2. **Validate input** before database operations
3. **Use middleware** for cross-cutting concerns (auth, logging, etc.)
4. **Keep handlers thin** - move business logic to separate functions
5. **Use async jobs** for time-consuming operations
6. **Enable auto-reload** in development for faster iteration
7. **Store errors in DB** for production debugging
8. **Use route groups** for API versioning and shared middleware

---

See individual documentation files for more detailed information on each component.
