package system

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
	"todo/config"
	"todo/internal/platform"
)

// RegisterRoutes mounts all system and diagnostics endpoints.
func RegisterRoutes(r router.Router, cfg *config.Settings, plat *platform.Container) {
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "healthy",
			"service": cfg.AppName,
			"env":     cfg.Environment,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	if plat.Metrics != nil {
		r.Get("/metrics", plat.Metrics.ServeHTTP)
	}
}
