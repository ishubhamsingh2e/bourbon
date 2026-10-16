package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ishubhamsingh2e/bourbon/config"
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/logger"
	"github.com/ishubhamsingh2e/bourbon/logger/slogadapter"
	"github.com/ishubhamsingh2e/bourbon/router"
)

// Option represents a functional configuration option for App.
type Option func(*App)

// App orchestrates application bootstrapping, DI container, routing, and lifecycle events.
type App struct {
	config    *config.Config
	container *Container
	router    router.Router
	logger    logger.Logger
	server    *http.Server
	providers []core.Provider
}

// New constructs an App instance with sensible defaults and applied functional options.
func New(opts ...Option) *App {
	a := &App{
		config:    config.LoadFromEnv(),
		container: NewContainer(),
		router:    router.New(),
		logger:    slogadapter.NewDefault(),
		providers: make([]core.Provider, 0),
	}

	for _, opt := range opts {
		opt(a)
	}

	// Bind core singletons into container
	a.container.Singleton(a.config)
	a.container.Singleton(a.router)
	a.container.Singleton(a.logger)

	return a
}

// WithConfig sets a custom Config for the application.
func WithConfig(cfg *config.Config) Option {
	return func(a *App) {
		if cfg != nil {
			a.config = cfg
		}
	}
}

// WithProviders registers service providers with the application.
func WithProviders(providers ...core.Provider) Option {
	return func(a *App) {
		a.providers = append(a.providers, providers...)
	}
}

// WithRouter sets a custom router adapter implementation.
func WithRouter(r router.Router) Option {
	return func(a *App) {
		if r != nil {
			a.router = r
		}
	}
}

// WithLogger sets a custom structured logger.
func WithLogger(l logger.Logger) Option {
	return func(a *App) {
		if l != nil {
			a.logger = l
		}
	}
}

// Router returns the application router.
func (a *App) Router() router.Router {
	return a.router
}

// Container returns the application DI container.
func (a *App) Container() core.Container {
	return a.container
}

// Logger returns the structured application logger.
func (a *App) Logger() logger.Logger {
	return a.logger
}

// Config returns the application configuration.
func (a *App) Config() *config.Config {
	return a.config
}

// Use registers one or more modular service providers into the application.
func (a *App) Use(providers ...core.Provider) core.Application {
	a.providers = append(a.providers, providers...)
	return a
}

// Run boots all service providers and starts the HTTP server with graceful shutdown.
func (a *App) Run(ctx context.Context) error {
	// 1. Register Providers
	for _, p := range a.providers {
		if err := p.Register(a.container); err != nil {
			return fmt.Errorf("app: provider register error: %w", err)
		}
	}

	// 2. Boot Providers
	for _, p := range a.providers {
		if err := p.Boot(ctx, a); err != nil {
			return fmt.Errorf("app: provider boot error: %w", err)
		}
	}

	// 3. Configure HTTP Server
	addr := fmt.Sprintf("%s:%d", a.config.HTTP.Host, a.config.HTTP.Port)
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a.router,
		ReadTimeout:  a.config.HTTP.ReadTimeout,
		WriteTimeout: a.config.HTTP.WriteTimeout,
		IdleTimeout:  a.config.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 4. Wait for Termination Signal or Context Cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-sigChan:
	case err := <-serverErr:
		return fmt.Errorf("app: http server failed to start: %w", err)
	}

	// 5. Graceful Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.Shutdown(shutdownCtx)
}

// Shutdown gracefully stops the HTTP server and executes provider shutdowns in reverse order.
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("app: http server shutdown error: %w", err))
		}
	}

	// Shutdown providers in reverse registration order
	for i := len(a.providers) - 1; i >= 0; i-- {
		if err := a.providers[i].Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("app: provider shutdown error: %w", err))
		}
	}

	return errors.Join(errs...)
}

var _ core.Application = (*App)(nil)
