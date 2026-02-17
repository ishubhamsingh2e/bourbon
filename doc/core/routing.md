# Routing

Bourbon provides a powerful routing engine that supports RESTful methods, route grouping, parameters, and middleware.

## Basic Routing

Routes are defined on the `app.Router` instance. The router supports standard HTTP methods:

```go
func RegisterRoutes(app *core.App) {
    app.Router.Get("/users", listUsers)
    app.Router.Post("/users", createUser)
    app.Router.Put("/users/:id", updateUser)
    app.Router.Delete("/users/:id", deleteUser)
}
```

## Route Handlers

A handler function signature is `func(*http.Context) error`.

```go
func listUsers(c *http.Context) error {
    return c.JSON(200, map[string]string{"message": "List users"})
}
```

## Route Parameters

You can define named parameters in your route path using `:paramName`. Access them using `c.Param("paramName")`.

```go
app.Router.Get("/users/:id", func(c *http.Context) error {
    id := c.Param("id")
    return c.String(200, "User ID: " + id)
})
```

## Route Groups

Grouping routes allows you to share a common prefix and middleware.

```go
api := app.Router.Group("/api/v1")
{
    api.Get("/users", listUsers)
    api.Get("/posts", listPosts)
}

// Admin group with authentication middleware
admin := app.Router.Group("/admin", authMiddleware)
{
    admin.Get("/dashboard", adminDashboard)
}
```

### Path Normalization

Route groups automatically normalize paths to prevent double slashes and ensure clean URLs:

```go
// These all result in the same clean path: "/"
root := app.Router.Group("/")
root.Get("/", homeHandler)  // Registered as GET /

// Nested groups work correctly
api := app.Router.Group("/api")
v1 := api.Group("/v1")      // Clean path: /api/v1
v1.Get("/users", listUsers) // Clean path: /api/v1/users
```

The router uses `path.Clean()` internally to remove redundant slashes and ensure patterns are valid.

## Static Files

Serve static files using `app.Static` or configured via `settings.toml`.

```go
// In code
app.Static("/static", "./static")

// In settings.toml
[static]
directory = "static"
url_prefix = "/static"
```

The application will serve files from the `./static` directory at the `/static` URL prefix.
