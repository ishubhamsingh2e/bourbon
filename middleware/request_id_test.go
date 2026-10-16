package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddlewareGenerated(t *testing.T) {
	reqIDMw := RequestID()

	var capturedID string
	handler := reqIDMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetReqID(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	handler.ServeHTTP(w, req)

	resHeader := w.Header().Get(HeaderXRequestID)
	if resHeader == "" {
		t.Fatalf("expected X-Request-ID header, got empty")
	}

	if capturedID != resHeader {
		t.Errorf("context value %v != response header %v", capturedID, resHeader)
	}
}

func TestRequestIDMiddlewarePropagated(t *testing.T) {
	reqIDMw := RequestID()

	handler := reqIDMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderXRequestID, "custom-id-12345")

	handler.ServeHTTP(w, req)

	if w.Header().Get(HeaderXRequestID) != "custom-id-12345" {
		t.Errorf("expected propagated X-Request-ID 'custom-id-12345', got %v", w.Header().Get(HeaderXRequestID))
	}
}
