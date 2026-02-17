# Requests and Responses

Bourbon uses a `Context` object to encapsulate the HTTP request and response. This object is passed to every route handler.

## Request Handling

### Path Parameters

Access named path parameters using `c.Param()`.

```go
// Route: /users/:id
id := c.Param("id")
```

### Query Parameters

Access query string parameters using `c.Query()`.

```go
// URL: /users?search=john
search := c.Query("search")
```

### Form Data

Access form values using `c.FormValue()`.

```go
name := c.FormValue("name")
```

### Request Body

Bind JSON or form data to a struct using `c.Bind()` or `c.Body()`.

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

var user User
if err := c.Bind(&user); err != nil {
    return c.JSON(400, map[string]string{"error": "Invalid request"})
}
```

### Request Context

You can store and retrieve values in the request context (useful for middleware).

```go
c.Set("currentUser", user)
currentUser := c.Get("currentUser").(User)
```

## Response Handling

### JSON Response

Send a JSON response using `c.JSON()`.

```go
c.JSON(200, map[string]interface{}{
    "message": "User created",
    "user_id": 123,
})
```

You can also use `http.H` (alias for `map[string]interface{}`) for convenience:

```go
c.JSON(200, http.H{"status": "ok"})
```

### String Response

Send a plain text response using `c.String()`.

```go
c.String(200, "Hello, World!")
```

### HTML Response

Send raw HTML using `c.HTML()`.

```go
c.HTML(200, "<h1>Hello</h1>")
```

### Redirect

Redirect the client to another URL.

```go
c.Redirect(302, "/login")
```

## Template Rendering

Render HTML templates using `c.Render()`.

```go
c.Render("profile.html", map[string]interface{}{
    "User": user,
})
```

Templates are loaded from the configured `templates` directory and can be organized into subdirectories. Bourbon supports template inheritance and partials.

## Async Jobs

Bourbon has a unique feature to dispatch asynchronous jobs directly from the context.

```go
// Dispatch a job
jobID, err := c.DispatchAsync("sendEmail", map[string]interface{}{
    "to": "user@example.com",
    "subject": "Welcome",
})

// Or dispatch and return JSON immediately
c.DispatchAsyncJSON(202, "processImage", payload)
```

This requires configuring an async dispatcher.
