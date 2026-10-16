package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/session"
	"github.com/ishubhamsingh2e/bourbon/session/memstore"
)

func TestSessionMiddlewareFlow(t *testing.T) {
	store := memstore.New(0)
	defer store.Close()

	sessMw := Session(store)

	// Step 1: Initial request sets session values and flash message
	handler1 := sessMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := session.FromContext(r.Context())
		if s == nil {
			t.Fatalf("session was not injected into context")
		}
		s.Set("auth_user", "bob")
		s.Flash("msg", "Welcome Back!")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Session Initialized"))
	}))

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/login", nil)

	handler1.ServeHTTP(w1, req1)

	setCookieHeader := w1.Header().Get("Set-Cookie")
	if !strings.Contains(setCookieHeader, "bourbon_session=") {
		t.Fatalf("expected Set-Cookie header with bourbon_session: %v", setCookieHeader)
	}

	// Extract cookie value for subsequent request
	cookieParts := strings.Split(setCookieHeader, ";")
	cookieVal := strings.TrimPrefix(cookieParts[0], "bourbon_session=")

	// Step 2: Next request reads session values and consumes flash message
	var readUser any
	var readFlash any
	handler2 := sessMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := session.FromContext(r.Context())
		readUser, _ = s.Get("auth_user")
		readFlash, _ = s.Flashed("msg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Dashboard"))
	}))

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req2.AddCookie(&http.Cookie{Name: "bourbon_session", Value: cookieVal})

	handler2.ServeHTTP(w2, req2)

	if readUser != "bob" {
		t.Errorf("readUser = %v, want bob", readUser)
	}
	if readFlash != "Welcome Back!" {
		t.Errorf("readFlash = %v, want Welcome Back!", readFlash)
	}
}
