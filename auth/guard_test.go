package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/session"
	"github.com/ishubhamsingh2e/bourbon/session/memstore"
)

func TestGuardManagerAndGuards(t *testing.T) {
	ctx := context.Background()
	store := memstore.New(0)
	defer store.Close()

	// 1. Setup SessionGuard
	sess := session.New("sess_guard_test")
	sess.Set("auth_principal", "user_sess_99")
	_ = store.Save(ctx, sess, time.Hour)

	sessGuard := auth.NewSessionGuard(store, "bourbon_session")

	// 2. Setup TokenGuard
	tokenProvider := auth.NewHMACTokenProvider([]byte("guard_secret_123"))
	tokenGuard := auth.NewTokenGuard(tokenProvider)

	p := auth.NewPrincipal("user_token_88").WithRoles("api_client")
	validToken, _ := tokenProvider.Generate(p, time.Hour)

	// 3. Setup Manager
	manager := auth.NewManager("session")
	manager.Register("session", sessGuard)
	manager.Register("token", tokenGuard)

	// Test Session Request
	reqSess := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqSess.AddCookie(&http.Cookie{Name: "bourbon_session", Value: "sess_guard_test"})

	authSess, err := manager.Authenticate(ctx, reqSess, "session")
	if err != nil || authSess.ID() != "user_sess_99" {
		t.Fatalf("Session guard failed: %v, principal: %+v", err, authSess)
	}

	// Test Token Request
	reqToken := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	reqToken.Header.Set("Authorization", "Bearer "+validToken)

	authToken, err := manager.Authenticate(ctx, reqToken, "token")
	if err != nil || authToken.ID() != "user_token_88" {
		t.Fatalf("Token guard failed: %v, principal: %+v", err, authToken)
	}

	// Test Fallback / Multi-Guard Authenticate
	authFallback, err := manager.Authenticate(ctx, reqToken, "session", "token")
	if err != nil || authFallback.ID() != "user_token_88" {
		t.Fatalf("Multi-guard fallback failed: %v", err)
	}
}
