# Middleware

Bourbon supports two types of middleware:

1.  **Global Middleware:** Applied to all requests using `http.Handler`.
2.  **Route/Group Middleware:** Applied to specific routes or groups using `HandlerFunc` wrappers.

## Global Middleware

Global middleware intercepts every request entering the application. It is defined as a standard `http.Handler` wrapper:

```go
func MyGlobalMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before request
        next.ServeHTTP(w, r)
        // After request
    })
}
```

Register global middleware in `main.go` using `app.RegisterMiddleware` and `app.UseMiddleware`.

```go
app.RegisterMiddleware("myMiddleware", MyGlobalMiddleware)
app.UseMiddleware("myMiddleware")
```

### Built-in Global Middleware

Bourbon comes with several built-in global middlewares:

- **Logger:** Logs requests and responses.
- **Recovery:** Recovers from panics and logs errors.
- **CORS:** Handles Cross-Origin Resource Sharing.

Enable them in `settings.toml`:

```toml
[middleware]
enabled = ["Logger", "Recovery", "CORS"]
```

## Route/Group Middleware

Route middleware wraps specific handlers and has access to the `Context`. It uses the signature `func(HandlerFunc) HandlerFunc`.

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(c *http.Context) error {
        // Check authentication
        if !isAuthenticated(c) {
            return c.JSON(401, map[string]string{"error": "Unauthorized"})
        }
        return next(c)
    }
}
```

Apply it to a group or route:

```go
admin := app.Router.Group("/admin", AuthMiddleware)
{
    admin.Get("/dashboard", dashboardHandler)
}
```

## Creating Custom Middleware

### Global Middleware (Standard)

Use this for cross-cutting concerns like logging, security headers, or request tracking that apply to the whole app.

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.New().String()
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r)
    })
}
```

### Route Middleware (Context-Aware)

Use this for logic that depends on route parameters or specific application state, like authentication or validation.

```go
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
    return func(c *http.Context) error {
        user := c.Get("user").(User)
        if !user.IsAdmin {
            return c.JSON(403, http.H{"error": "Forbidden"})
        }
        return next(c)
    }
}
```
