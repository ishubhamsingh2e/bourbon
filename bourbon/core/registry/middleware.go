package registry

import (
	"sync"

	"github.com/ishubhamsingh2e/bourbon/bourbon/middleware"
)

// MiddlewareFunc is an alias for the standard middleware function type
type MiddlewareFunc = middleware.Middleware

// MiddlewareRegistry holds registered middlewares with their names
type MiddlewareRegistry struct {
	middlewares map[string]MiddlewareFunc
	mu          sync.RWMutex
}

// NewMiddlewareRegistry creates a new middleware registry
func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{
		middlewares: make(map[string]MiddlewareFunc),
	}
}

// Register registers a named middleware function
func (mr *MiddlewareRegistry) Register(name string, middleware MiddlewareFunc) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.middlewares[name] = middleware
}

// Get retrieves a middleware by name
func (mr *MiddlewareRegistry) Get(name string) (MiddlewareFunc, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	middleware, exists := mr.middlewares[name]
	return middleware, exists
}

// Has checks if a middleware is registered
func (mr *MiddlewareRegistry) Has(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	_, exists := mr.middlewares[name]
	return exists
}

// Unregister removes a middleware from the registry
func (mr *MiddlewareRegistry) Unregister(name string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	delete(mr.middlewares, name)
}

// List returns all registered middleware names
func (mr *MiddlewareRegistry) List() []string {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	
	names := make([]string, 0, len(mr.middlewares))
	for name := range mr.middlewares {
		names = append(names, name)
	}
	return names
}

// Global middleware registry
var globalMiddlewareRegistry = NewMiddlewareRegistry()

// RegisterMiddleware registers a middleware globally
func RegisterMiddleware(name string, mw MiddlewareFunc) {
	globalMiddlewareRegistry.Register(name, mw)
}

// GetMiddleware retrieves a globally registered middleware
func GetMiddleware(name string) (MiddlewareFunc, bool) {
	return globalMiddlewareRegistry.Get(name)
}

// HasMiddleware checks if a middleware is registered globally
func HasMiddleware(name string) bool {
	return globalMiddlewareRegistry.Has(name)
}

// UnregisterMiddleware removes a middleware from global registry
func UnregisterMiddleware(name string) {
	globalMiddlewareRegistry.Unregister(name)
}

// ListMiddlewares returns all globally registered middleware names
func ListMiddlewares() []string {
	return globalMiddlewareRegistry.List()
}

