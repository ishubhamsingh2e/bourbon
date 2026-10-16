package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	logMw := Logger(logger)

	handler := logMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Created item"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/items", nil)

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("w.Code = %v, want 201", w.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, `"method":"POST"`) || !strings.Contains(logOutput, `"path":"/items"`) || !strings.Contains(logOutput, `"status":201`) {
		t.Errorf("structured log missing request attributes: %v", logOutput)
	}
}
