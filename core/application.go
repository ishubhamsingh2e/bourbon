// Package core defines the minimal application shell interfaces, DI container contracts, and lifecycle provider standards.
package core

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/logger"
	"github.com/ishubhamsingh2e/bourbon/router"
)

// Application defines the minimal application shell required to execute the framework.
type Application interface {
	// Container returns the DI container for service registration and resolution.
	Container() Container

	// Router returns the HTTP routing engine.
	Router() router.Router

	// Logger returns the structured logging port.
	Logger() logger.Logger

	// Use registers one or more modular service providers into the application.
	Use(providers ...Provider) Application

	// Run boots all registered providers and starts the HTTP server with graceful shutdown.
	Run(ctx context.Context) error

	// Shutdown terminates the application and releases resources.
	Shutdown(ctx context.Context) error
}

// Container defines the dependency injection container contract.
type Container interface {
	// Singleton binds an instantiated service or a single-eval factory constructor.
	Singleton(service any)

	// Transient registers a constructor function for transient instantiation on each Resolve.
	Transient(service any)

	// Resolve populates the target pointer with an instance of the requested type.
	Resolve(target any) error
}

// Provider defines the 3-phase lifecycle contract for any subsystem or module.
type Provider interface {
	// Register binds interfaces, configurations, and constructor factories into the container.
	Register(c Container) error

	// Boot executes active initialization (database pings, route binding) after all providers register.
	Boot(ctx context.Context, app Application) error

	// Shutdown releases resources during graceful process termination.
	Shutdown(ctx context.Context) error
}

// HandlerFunc defines a Bourbon HTTP handler receiving *Context natively.
type HandlerFunc func(c *Context) error
