package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
)

func TestAuthAndRBACMiddleware(t *testing.T) {
	tokenProvider := auth.NewHMACTokenProvider([]byte("secret_jwt_key"))
	tokenGuard := auth.NewTokenGuard(tokenProvider)

	manager := auth.NewManager("token")
	manager.Register("token", tokenGuard)

	rbac := auth.NewRBACRegistry()
	rbac.DefineRole("admin", "users:*", "billing:*")
	rbac.DefineRole("editor", "posts:*")

	// Principal with editor role + direct custom permission "reports:download"
	editorPrincipal := auth.NewPrincipal("usr_editor").
		WithRoles("editor").
		WithPermissions("reports:download")

	token, _ := tokenProvider.Generate(editorPrincipal, time.Hour)

	authMw := RequireAuth(manager, "token")
	roleMw := RequireRole("editor")
	adminRoleMw := RequireRole("admin")
	permMw := RequirePermission(rbac, "posts:create")
	directPermMw := RequirePermission(rbac, "reports:download")
	forbiddenPermMw := RequirePermission(rbac, "billing:refund")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Authorized"))
	})

	// 1. Success Flow (Auth + Role + Wildcard Permission + Direct Permission)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	chain := authMw(roleMw(permMw(directPermMw(handler))))
	chain.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}

	// 2. Failure: missing admin role (403)
	wAdmin := httptest.NewRecorder()
	reqAdmin := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+token)

	authMw(adminRoleMw(handler)).ServeHTTP(wAdmin, reqAdmin)
	if wAdmin.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for missing admin role, got: %d", wAdmin.Code)
	}

	// 3. Failure: missing permission (403)
	wPerm := httptest.NewRecorder()
	reqPerm := httptest.NewRequest(http.MethodPost, "/billing/refund", nil)
	reqPerm.Header.Set("Authorization", "Bearer "+token)

	authMw(forbiddenPermMw(handler)).ServeHTTP(wPerm, reqPerm)
	if wPerm.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for missing billing:refund, got: %d", wPerm.Code)
	}

	// 4. Failure: unauthenticated request (401)
	wUnauth := httptest.NewRecorder()
	reqUnauth := httptest.NewRequest(http.MethodGet, "/posts", nil)

	authMw(handler).ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got: %d", wUnauth.Code)
	}
}
