# Authentication & RBAC Authorization in Bourbon

Bourbon features a multi-guard authentication engine, standard crypto primitives, and a comprehensive Role-Based Access Control (RBAC) authorization system with wildcard permission matching.

---

## 1. Multi-Guard Authentication

Bourbon decouples credentials checking from identity storage using the `auth.Guard` port:

- **`auth.SessionGuard`**: Stateful cookie-based authentication for web apps.
- **`auth.TokenGuard`**: Stateless Bearer token authentication for APIs.

```go
package security

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/session"
)

func SetupAuth(sessionStore session.Store) *auth.Manager {
	manager := auth.NewManager()

	// 1. Register Session Guard
	sessionGuard := auth.NewSessionGuard(sessionStore, "user_id", func(ctx context.Context, id string) (auth.Principal, error) {
		return auth.NewPrincipal(id).WithRoles("member"), nil
	})
	manager.RegisterGuard("web", sessionGuard)

	// 2. Register Stateless Token Guard (HMAC / JWT)
	tokenProvider := auth.NewHMACProvider([]byte("32-byte-secret-key-12345678901234"))
	tokenGuard := auth.NewTokenGuard(tokenProvider, func(ctx context.Context, claims auth.Claims) (auth.Principal, error) {
		return auth.NewPrincipal(claims.Subject).
			WithRoles(claims.Roles...).
			WithPermissions(claims.Permissions...), nil
	})
	manager.RegisterGuard("api", tokenGuard)

	return manager
}
```

---

## 2. Granular RBAC & Wildcard Permissions

Bourbon's `auth.RBACRegistry` supports multiple roles per principal, role-to-role inheritance, and wildcard matching (`users:*`, `reports:export:*`):

```go
package security

import "github.com/ishubhamsingh2e/bourbon/auth"

func SetupRBAC() *auth.RBACRegistry {
	rbac := auth.NewRBACRegistry()

	// Define role permissions with wildcards
	rbac.RegisterRole("viewer", []string{"posts:read", "comments:read"})
	rbac.RegisterRole("editor", []string{"posts:*", "comments:*"}, "viewer") // inherits viewer
	rbac.RegisterRole("admin", []string{"*"}, "editor")                      // full wildcard access

	return rbac
}
```

---

## 3. Protecting Routes with Middleware

Bourbon provides ready-to-use route protection middlewares in `middleware/`:

```go
package routes

import (
	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func RegisterProtectedRoutes(r router.Router, authManager *auth.Manager, rbac *auth.RBACRegistry) {
	// Require valid authentication via "api" guard
	api := r.Group("/api", middleware.RequireAuth(authManager, "api"))
	{
		// Require specific role
		admin := api.Group("/admin", middleware.RequireRole("admin"))
		admin.Get("/stats", GetSystemStats)

		// Require granular permission (checks role inheritance and wildcards)
		api.Delete("/posts/{id}", DeletePost, middleware.RequirePermission(rbac, "posts:delete"))
	}
}
```

---

## 4. Accessing Principals in Handlers

Use `c.Auth()` on `*context.Context` to access the authenticated principal:

```go
func GetCurrentUserProfile(c *context.Context) error {
	principal := c.Auth()
	if principal == nil {
		return c.Unauthorized("not authenticated")
	}

	return c.JSON(200, map[string]any{
		"user_id":     principal.ID(),
		"roles":       principal.Roles(),
		"permissions": principal.Permissions(),
	})
}
```

---

## 5. Password Hashing

Bourbon includes standard BCrypt password hashing backed by `golang.org/x/crypto/bcrypt`:

```go
hasher := auth.NewBcryptHasher()

// Hash password
hash, err := hasher.Hash("securePassword123")

// Compare password
valid := hasher.Compare(hash, "securePassword123") // true
```
