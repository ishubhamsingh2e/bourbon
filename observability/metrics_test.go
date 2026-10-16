package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsCollectorRecordingAndExpose(t *testing.T) {
	collector := NewMetricsCollector()

	collector.Record(http.MethodGet, "/users", 200, 15*time.Millisecond)
	collector.Record(http.MethodPost, "/users", 201, 25*time.Millisecond)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	collector.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `bourbon_http_requests_total{method="GET",path="/users",status="200"} 1`) {
		t.Errorf("missing GET /users metric: %s", body)
	}
	if !strings.Contains(body, `bourbon_http_requests_total{method="POST",path="/users",status="201"} 1`) {
		t.Errorf("missing POST /users metric: %s", body)
	}
	if !strings.Contains(body, "bourbon_http_request_duration_seconds") {
		t.Errorf("missing duration metric: %s", body)
	}
}
