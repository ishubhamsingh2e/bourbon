package middleware

import (
	"fmt"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/router"
)

// RequireAuth ensures that the request is authenticated via the specified guards.
func RequireAuth(manager *auth.Manager, guardNames ...string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := manager.Authenticate(r.Context(), r, guardNames...)
			if err != nil || principal == nil {
				w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, `{"type":"https://bourbon.dev/errors/unauthenticated","title":"Unauthenticated","status":401,"detail":"Authentication credentials were not provided or are invalid.","instance":"%s"}`, r.URL.Path)
				return
			}

			ctx := auth.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole ensures that the authenticated principal possesses all the specified roles.
func RequireRole(roles ...string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := auth.PrincipalFromContext(r.Context())
			if p == nil {
				forbiddenResponse(w, r, "Authentication required to check roles.")
				return
			}

			for _, role := range roles {
				if !p.HasRole(role) {
					forbiddenResponse(w, r, "Missing required role: "+role)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole ensures that the authenticated principal possesses at least one of the specified roles.
func RequireAnyRole(roles ...string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := auth.PrincipalFromContext(r.Context())
			if p == nil || !p.HasAnyRole(roles...) {
				forbiddenResponse(w, r, "Principal does not have any of the required roles.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission ensures that the principal holds all specified permissions (evaluated via RBAC registry).
func RequirePermission(rbac *auth.RBACRegistry, permissions ...string) router.MiddlewareFunc {
	return RequireAllPermissions(rbac, permissions...)
}

// RequireAllPermissions ensures that the principal holds all specified permissions.
func RequireAllPermissions(rbac *auth.RBACRegistry, permissions ...string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := auth.PrincipalFromContext(r.Context())
			if p == nil || !rbac.CheckAllPermissions(p, permissions...) {
				forbiddenResponse(w, r, "Access denied. Missing required permissions.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission ensures that the principal holds at least one of the specified permissions.
func RequireAnyPermission(rbac *auth.RBACRegistry, permissions ...string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := auth.PrincipalFromContext(r.Context())
			if p == nil || !rbac.CheckAnyPermission(p, permissions...) {
				forbiddenResponse(w, r, "Access denied. None of the required permissions were met.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireCan evaluates authorization policies on the authenticated principal.
func RequireCan(enforcer auth.Enforcer, action string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := auth.PrincipalFromContext(r.Context())
			if p == nil {
				forbiddenResponse(w, r, "Authentication required.")
				return
			}

			can, err := enforcer.Can(r.Context(), p, action, nil)
			if err != nil || !can {
				forbiddenResponse(w, r, "Action forbidden by authorization policy.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func forbiddenResponse(w http.ResponseWriter, r *http.Request, detail string) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = fmt.Fprintf(w, `{"type":"https://bourbon.dev/errors/unauthorized","title":"Forbidden","status":403,"detail":"%s","instance":"%s"}`, detail, r.URL.Path)
}
