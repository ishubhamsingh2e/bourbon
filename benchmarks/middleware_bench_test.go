package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func BenchmarkMiddlewarePipelineOverhead(b *testing.B) {
	r := router.New()

	// Stack common middlewares
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(5 * time.Second))

	r.Get("/benchmark", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}
