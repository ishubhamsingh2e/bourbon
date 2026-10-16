package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFMiddleware(t *testing.T) {
	csrfMw := CSRF()

	handler := csrfMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Form processed"))
	}))

	// 1. GET request generates cookie
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/form", nil)

	handler.ServeHTTP(w1, req1)

	setCookie := w1.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "bourbon_csrf=") {
		t.Fatalf("expected Set-Cookie with bourbon_csrf, got: %s", setCookie)
	}

	parts := strings.Split(setCookie, ";")
	tokenVal := strings.TrimPrefix(parts[0], "bourbon_csrf=")

	// 2. Unsafe POST without token fails (403)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req2.AddCookie(&http.Cookie{Name: "bourbon_csrf", Value: tokenVal})

	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for missing CSRF header, got: %d", w2.Code)
	}

	// 3. Unsafe POST with valid token header succeeds (200)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req3.AddCookie(&http.Cookie{Name: "bourbon_csrf", Value: tokenVal})
	req3.Header.Set("X-CSRF-Token", tokenVal)

	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w3.Code)
	}
}
