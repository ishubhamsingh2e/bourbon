package otel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenTelemetryMiddleware(t *testing.T) {
	tracer := NewTracer("bourbon-api")
	mw := Middleware(tracer)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	handler.ServeHTTP(w, req)

	traceHeader := w.Header().Get("traceparent")
	if !strings.HasPrefix(traceHeader, "00-") {
		t.Errorf("missing or invalid traceparent header: %s", traceHeader)
	}

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}
	span := tracer.spans[0]
	if span.Name != "HTTP GET /orders/42" {
		t.Errorf("span name mismatch: %s", span.Name)
	}
	if span.Attributes["http.method"] != "GET" {
		t.Errorf("attribute mismatch")
	}
}
