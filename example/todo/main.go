package main

import (
	"context"
	"log"
	"time"

	"todo/config"
	"todo/internal/modules/system"
	"todo/internal/modules/users"
	"todo/internal/platform"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/middleware"
)

func main() {
	cfg := config.Load()

	plat, err := platform.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed initializing platform: %v", err)
	}

	userRepo := users.NewRepository(plat.DB)
	userService := users.NewService(userRepo)

	application := app.New(
		app.WithConfig(cfg.AppConfig()),
	)

	r := application.Router()

	// 1. Global Middleware Stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))
	if plat.Metrics != nil {
		r.Use(plat.Metrics.Middleware())
	}

	// 2. Feature Routes
	system.RegisterRoutes(r, cfg, plat)
	users.RegisterRoutes(r, userService)

	log.Printf("Starting %s on :%d", cfg.AppName, cfg.Port)
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}
