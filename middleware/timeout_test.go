package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddlewareSuccess(t *testing.T) {
	toMw := Timeout(100 * time.Millisecond)

	handler := toMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Fast response"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %v, want 200", w.Code)
	}
}

func TestTimeoutMiddlewareExceeded(t *testing.T) {
	toMw := Timeout(20 * time.Millisecond)

	handler := toMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Slow response"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("w.Code = %v, want 504", w.Code)
	}
}
