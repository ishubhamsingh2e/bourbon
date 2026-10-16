package router

import (
	"net/http"
	"sync"
)

// Engine is the native Radix Trie router for Bourbon applications with Chi-style semantics.
type Engine struct {
	mu               sync.RWMutex
	tree             *Tree
	middlewares      []MiddlewareFunc
	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc
}

// New initializes the native Bourbon Router engine.
func New() *Engine {
	return &Engine{
		tree:        NewTree(),
		middlewares: make([]MiddlewareFunc, 0),
		notFound: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "404 page not found", http.StatusNotFound)
		},
		methodNotAllowed: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		},
	}
}

// NewRouter initializes an empty Bourbon Router engine (alias for New).
func NewRouter() *Engine {
	return New()
}

// Use appends global middlewares to the router engine.
func (e *Engine) Use(mw ...MiddlewareFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middlewares = append(e.middlewares, mw...)
}

// With creates an inline sub-router with additional middlewares attached.
func (e *Engine) With(mw ...MiddlewareFunc) Router {
	return NewGroup(e, "", mw...)
}

// Group creates an inline route group with its own scoped middlewares.
func (e *Engine) Group(fn func(r Router)) Router {
	g := NewGroup(e, "")
	if fn != nil {
		fn(g)
	}
	return g
}

// Route creates a sub-router scoped to pattern.
func (e *Engine) Route(pattern string, fn func(r Router)) Router {
	g := NewGroup(e, pattern)
	if fn != nil {
		fn(g)
	}
	return g
}

// Mount mounts an http.Handler onto the given path prefix.
func (e *Engine) Mount(pattern string, handler http.Handler) {
	e.tree.InsertMount(pattern, handler)
}

// NotFound sets a custom 404 handler.
func (e *Engine) NotFound(h any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	adapted := AdaptHandler(h)
	e.notFound = adapted.ServeHTTP
}

// MethodNotAllowed sets a custom 405 handler.
func (e *Engine) MethodNotAllowed(h any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	adapted := AdaptHandler(h)
	e.methodNotAllowed = adapted.ServeHTTP
}

// Method registers a handler for a specific HTTP method.
func (e *Engine) Method(method, pattern string, handler any) {
	e.tree.Insert(method, pattern, AdaptHandler(handler))
}

// MethodFunc registers a handler for a specific HTTP method.
func (e *Engine) MethodFunc(method, pattern string, h any) {
	e.Method(method, pattern, h)
}

// Handle registers a handler for all standard HTTP methods.
func (e *Engine) Handle(pattern string, handler any) {
	adapted := AdaptHandler(handler)
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions, http.MethodTrace, http.MethodConnect,
	}
	for _, m := range methods {
		e.tree.Insert(m, pattern, adapted)
	}
}

// HandleFunc registers a handler for all standard HTTP methods.
func (e *Engine) HandleFunc(pattern string, h any) {
	e.Handle(pattern, h)
}

// Connect registers a CONNECT route.
func (e *Engine) Connect(pattern string, h any) {
	e.Method(http.MethodConnect, pattern, h)
}

// Delete registers a DELETE route.
func (e *Engine) Delete(pattern string, h any) {
	e.Method(http.MethodDelete, pattern, h)
}

// Get registers a GET route.
func (e *Engine) Get(pattern string, h any) {
	e.Method(http.MethodGet, pattern, h)
}

// Head registers a HEAD route.
func (e *Engine) Head(pattern string, h any) {
	e.Method(http.MethodHead, pattern, h)
}

// Options registers an OPTIONS route.
func (e *Engine) Options(pattern string, h any) {
	e.Method(http.MethodOptions, pattern, h)
}

// Patch registers a PATCH route.
func (e *Engine) Patch(pattern string, h any) {
	e.Method(http.MethodPatch, pattern, h)
}

// Post registers a POST route.
func (e *Engine) Post(pattern string, h any) {
	e.Method(http.MethodPost, pattern, h)
}

// Put registers a PUT route.
func (e *Engine) Put(pattern string, h any) {
	e.Method(http.MethodPut, pattern, h)
}

// Trace registers a TRACE route.
func (e *Engine) Trace(pattern string, h any) {
	e.Method(http.MethodTrace, pattern, h)
}

// ServeHTTP handles incoming HTTP requests using the Radix Trie wrapped by global middlewares.
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.mu.RLock()
	mws := e.middlewares
	e.mu.RUnlock()

	var handler http.Handler = http.HandlerFunc(e.dispatch)
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}

	handler.ServeHTTP(w, req)
}

func (e *Engine) dispatch(w http.ResponseWriter, req *http.Request) {
	rctx := RouteContext(req.Context())
	if rctx == nil {
		rctx = NewRouteContext()
		req = req.WithContext(WithRouteContext(req.Context(), rctx))
	}

	path := req.URL.Path
	res := e.tree.Search(req.Method, path)

	if res.mountServer != nil {
		for k, v := range res.params {
			rctx.URLParams.Add(k, v)
		}
		res.mountServer.ServeHTTP(w, req)
		return
	}

	if res.entry != nil {
		for k, v := range res.params {
			rctx.URLParams.Add(k, v)
		}
		rctx.RoutePattern = res.entry.pattern
		rctx.RoutePath = path
		res.entry.handler.ServeHTTP(w, req)
		return
	}

	if res.hasRoute {
		e.methodNotAllowed(w, req)
		return
	}

	e.notFound(w, req)
}

var _ Router = (*Engine)(nil)
