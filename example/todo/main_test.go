package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"todo/config"
	"todo/internal/modules/system"
	"todo/internal/modules/users"
	"todo/internal/platform"
)

func setupTestApp() *app.App {
	cfg := config.Load()
	plat, _ := platform.NewContainer(cfg)
	userRepo := users.NewRepository(plat.DB)
	userService := users.NewService(userRepo)

	application := app.New(app.WithConfig(cfg.AppConfig()))
	r := application.Router()

	system.RegisterRoutes(r, cfg, plat)
	users.RegisterRoutes(r, userService)
	return application
}

func TestHealthEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "healthy") {
		t.Errorf("expected body to contain 'healthy', got: %s", w.Body.String())
	}
}

func TestUsersEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "users") {
		t.Errorf("expected body to contain 'users', got: %s", w.Body.String())
	}
}
