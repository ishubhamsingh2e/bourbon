package observability

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/core"
)

func TestHealthCheckerLivenessAndReadiness(t *testing.T) {
	checker := NewHealthChecker()

	// 1. Liveness returns 200 OK
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	c1 := core.NewContext(w1, req1)

	_ = checker.Liveness()(c1)
	if w1.Code != http.StatusOK {
		t.Errorf("Liveness code = %d, want 200", w1.Code)
	}

	// 2. Readiness with passing check returns 200 OK
	checker.AddCheck("db", func(ctx context.Context) error {
		return nil
	})

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	c2 := core.NewContext(w2, req2)

	_ = checker.Readiness()(c2)
	if w2.Code != http.StatusOK {
		t.Errorf("Readiness code = %d, want 200", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), `"db":"up"`) {
		t.Errorf("Readiness body mismatch: %s", w2.Body.String())
	}

	// 3. Readiness with failing check returns 503
	checker.AddCheck("cache", func(ctx context.Context) error {
		return errors.New("redis connection refused")
	})

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	c3 := core.NewContext(w3, req3)

	_ = checker.Readiness()(c3)
	if w3.Code != http.StatusServiceUnavailable {
		t.Errorf("Readiness failing code = %d, want 503", w3.Code)
	}
	if !strings.Contains(w3.Body.String(), "redis connection refused") {
		t.Errorf("Readiness failing body mismatch: %s", w3.Body.String())
	}
}
