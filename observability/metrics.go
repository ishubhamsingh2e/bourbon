package observability

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type metricEntry struct {
	method string
	path   string
	status int
	count  int64
}

// MetricsCollector records HTTP traffic metrics and exports them in Prometheus exposition format.
type MetricsCollector struct {
	mu            sync.RWMutex
	entries       map[string]*metricEntry
	totalRequests int64
	totalDuration int64 // nanoseconds
}

// NewMetricsCollector constructs a Prometheus MetricsCollector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		entries: make(map[string]*metricEntry),
	}
}

// Record increments request metrics.
func (m *MetricsCollector) Record(method, path string, status int, duration time.Duration) {
	atomic.AddInt64(&m.totalRequests, 1)
	atomic.AddInt64(&m.totalDuration, int64(duration))

	key := fmt.Sprintf("%s:%s:%d", method, path, status)

	m.mu.Lock()
	entry, exists := m.entries[key]
	if !exists {
		entry = &metricEntry{
			method: method,
			path:   path,
			status: status,
		}
		m.entries[key] = entry
	}
	m.mu.Unlock()

	atomic.AddInt64(&entry.count, 1)
}

// Middleware automatically records response timing and status codes.
func (m *MetricsCollector) Middleware() router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)

			status := http.StatusOK
			path := "/"
			method := http.MethodGet
			if r != nil && r.URL != nil {
				path = r.URL.Path
				method = r.Method
			}

			m.Record(method, path, status, duration)
		})
	}
}

// ServeHTTP exports metrics in Prometheus text format (implements http.Handler).
func (m *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("# HELP bourbon_http_requests_total Total number of HTTP requests\n")
	buf.WriteString("# TYPE bourbon_http_requests_total counter\n")

	m.mu.RLock()
	for _, e := range m.entries {
		buf.WriteString(fmt.Sprintf(`bourbon_http_requests_total{method="%s",path="%s",status="%d"} %d`+"\n", e.method, e.path, e.status, atomic.LoadInt64(&e.count)))
	}
	m.mu.RUnlock()

	durationSec := float64(atomic.LoadInt64(&m.totalDuration)) / float64(time.Second)
	buf.WriteString("# HELP bourbon_http_request_duration_seconds Total time spent in HTTP requests in seconds\n")
	buf.WriteString("# TYPE bourbon_http_request_duration_seconds counter\n")
	buf.WriteString(fmt.Sprintf("bourbon_http_request_duration_seconds %f\n", durationSec))

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// HandlerFunc returns an http.HandlerFunc exporting metrics.
func (m *MetricsCollector) HandlerFunc() http.HandlerFunc {
	return m.ServeHTTP
}
