# Routing & Middleware in Bourbon

Bourbon delivers zero-allocation HTTP routing built upon Go 1.22+ standard library `http.ServeMux` conventions, extended with parameterized segments, group scoping, API versioning, and REST resource shortcuts.

---

## 1. Defining Routes

Register handlers using standard HTTP verb methods:

```go
package routes

import (
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func Register(r router.Router) {
	// Native net/http handler
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Native Bourbon *core.Context handlers (no Wrap() required)
	r.Get("/users", ListUsers)
	r.Post("/users", CreateUser)
	r.Get("/users/{id}", GetUserByID)
	r.Put("/users/{id}", UpdateUser)
	r.Delete("/users/{id}", DeleteUser)
}

func GetUserByID(c *core.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"user_id": id})
}
```

---

## 2. Route Groups & API Versioning

### Prefix Grouping
```go
api := r.Group(func(sub router.Router) {
	sub.Get("/metrics", GetMetrics)
})
```

### Versioning Shorthand
Bourbon includes first-class API versioning:

```go
router.Version(r, "v1", func(v1 router.Router) {
	v1.Get("/articles", ListArticlesV1)
})

router.Version(r, "v2", func(v2 router.Router) {
	v2.Get("/articles", ListArticlesV2)
})
```

---

## 3. REST Resource Controller Routing

Define standard RESTful resources using `r.Resource()`:

```go
// Automatically binds:
// GET    /posts          -> controller.Index
// POST   /posts          -> controller.Store
// GET    /posts/{id}     -> controller.Show
// PUT    /posts/{id}     -> controller.Update
// DELETE /posts/{id}     -> controller.Destroy
r.Resource("/posts", postController)
```

---

## 4. Built-in Middleware Stack

Bourbon provides production-ready middlewares located in `github.com/ishubhamsingh2e/bourbon/middleware`:

```go
package main

import (
	"time"

	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
	"github.com/ishubhamsingh2e/bourbon/session/memstore"
)

func ApplyMiddlewares(r router.Router) {
	// 1. Panic Recovery with structured stack trace logging
	r.Use(middleware.Recovery())

	// 2. Structured request logging via log/slog
	r.Use(middleware.Logger())

	// 3. Request ID injection (propagated via X-Request-ID header)
	r.Use(middleware.RequestID())

	// 4. Cross-Origin Resource Sharing (CORS)
	r.Use(middleware.CORS())

	// 5. Context Timeout deadline
	r.Use(middleware.Timeout(15 * time.Second))

	// 6. Double-Submit Cookie CSRF defense
	r.Use(middleware.CSRF(middleware.CSRFConfig{
		CookieName: "bourbon_csrf",
	}))

	// 7. Stateful HTTP session hydration
	sessionStore := memstore.New()
	r.Use(middleware.Session(sessionStore))
}
```

---

## 5. Custom Middleware

Bourbon middlewares use the standard Go `func(http.Handler) http.Handler` signature:

```go
func CustomHeaderMiddleware(headerName, headerVal string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerName, headerVal)
			next.ServeHTTP(w, r)
		})
	}
}
```
