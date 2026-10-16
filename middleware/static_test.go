package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestStaticMiddleware(t *testing.T) {
	mockFS := fstest.MapFS{
		"style.css": &fstest.MapFile{
			Data:    []byte("body { background: #000; }"),
			ModTime: time.Unix(1700000000, 0),
		},
		"index.html": &fstest.MapFile{
			Data:    []byte("<h1>Homepage</h1>"),
			ModTime: time.Unix(1700000000, 0),
		},
	}

	staticMw := Static(mockFS, StaticOptions{
		Prefix:       "/assets",
		EnableETag:   true,
		CacheControl: "public, max-age=3600",
	})

	handler := staticMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))

	// 1. Serving static asset
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/assets/style.css", nil)

	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /assets/style.css, got: %d", w1.Code)
	}
	if !strings.Contains(w1.Body.String(), "body { background: #000; }") {
		t.Errorf("content mismatch: %s", w1.Body.String())
	}
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Errorf("expected ETag header")
	}

	// 2. If-None-Match ETag conditional GET returns 304 Not Modified
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/assets/style.css", nil)
	req2.Header.Set("If-None-Match", etag)

	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotModified {
		t.Errorf("expected 304 Not Modified, got: %d", w2.Code)
	}

	// 3. Fallthrough for missing asset passes to next handler
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)

	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusNotFound {
		t.Errorf("expected fallthrough 404 for missing asset, got: %d", w3.Code)
	}
}
