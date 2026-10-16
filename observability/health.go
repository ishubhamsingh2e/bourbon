// Package observability provides health probes (/healthz, /readyz) and Prometheus metrics collectors.
package observability

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
)

// CheckFunc defines a health check evaluator callback.
type CheckFunc func(ctx context.Context) error

// HealthChecker manages liveness and readiness health check probes.
type HealthChecker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// NewHealthChecker constructs a HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
	}
}

// AddCheck registers a named dependency readiness check.
func (h *HealthChecker) AddCheck(name string, fn CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = fn
}

// Liveness returns a Bourbon handler func responding with 200 OK if the process is alive.
func (h *HealthChecker) Liveness() core.HandlerFunc {
	return func(c *core.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "up",
		})
	}
}

// Readiness returns a Bourbon handler func running all registered dependency checks.
func (h *HealthChecker) Readiness() core.HandlerFunc {
	return func(c *core.Context) error {
		h.mu.RLock()
		checksCopy := make(map[string]CheckFunc, len(h.checks))
		for k, v := range h.checks {
			checksCopy[k] = v
		}
		h.mu.RUnlock()

		ctx, cancel := context.WithTimeout(c.Ctx, 5*time.Second)
		defer cancel()

		results := make(map[string]string)
		allHealthy := true

		for name, check := range checksCopy {
			if err := check(ctx); err != nil {
				results[name] = "down: " + err.Error()
				allHealthy = false
			} else {
				results[name] = "up"
			}
		}

		status := "ready"
		statusCode := http.StatusOK
		if !allHealthy {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, map[string]any{
			"status": status,
			"checks": results,
		})
	}
}
