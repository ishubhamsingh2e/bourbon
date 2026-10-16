package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	recMw := Recovery(logger)

	panickyHandler := recMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database connection collapsed")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	panickyHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %v, want 500", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("response body does not contain expected error: %v", w.Body.String())
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic recovered in HTTP handler") || !strings.Contains(logOutput, "database connection collapsed") {
		t.Errorf("structured log does not contain panic info: %v", logOutput)
	}
}
