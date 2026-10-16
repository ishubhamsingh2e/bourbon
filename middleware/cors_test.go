package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewarePreflight(t *testing.T) {
	corsMw := CORS()

	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Should not be called"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("preflight status = %v, want 204", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("origin header = %v, want https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("credentials header = %v, want true", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestCORSMiddlewareRegularRequest(t *testing.T) {
	corsMw := CORS()

	handler := corsMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want 200", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Errorf("origin header = %v, want https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
