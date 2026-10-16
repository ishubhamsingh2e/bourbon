package router

import (
	"net/http"
	"strings"
)

// RouteGroup implements scoped sub-routing with path prefix compounding and isolated middleware.
type RouteGroup struct {
	parent      Router
	prefix      string
	middlewares []MiddlewareFunc
}

// NewGroup creates a new RouteGroup wrapping the parent router.
func NewGroup(parent Router, prefix string, middlewares ...MiddlewareFunc) *RouteGroup {
	cleanPrefix := "/" + strings.Trim(prefix, "/")
	if cleanPrefix == "/" {
		cleanPrefix = ""
	}
	return &RouteGroup{
		parent:      parent,
		prefix:      cleanPrefix,
		middlewares: middlewares,
	}
}

// Use appends middleware functions to this route group.
func (g *RouteGroup) Use(mw ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, mw...)
}

// With returns a new RouteGroup with additional scoped middlewares.
func (g *RouteGroup) With(mw ...MiddlewareFunc) Router {
	subMws := make([]MiddlewareFunc, 0, len(g.middlewares)+len(mw))
	subMws = append(subMws, g.middlewares...)
	subMws = append(subMws, mw...)
	return NewGroup(g.parent, g.prefix, subMws...)
}

func (g *RouteGroup) fullPattern(pattern string) string {
	clean := "/" + strings.Trim(pattern, "/")
	if clean == "/" && g.prefix != "" {
		return g.prefix
	}
	return g.prefix + clean
}

func (g *RouteGroup) wrap(h http.Handler) http.Handler {
	wrapped := h
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		wrapped = g.middlewares[i](wrapped)
	}
	return wrapped
}

// Method registers a handler for a specific HTTP method with group middlewares and prefix.
func (g *RouteGroup) Method(method, pattern string, handler any) {
	g.parent.Method(method, g.fullPattern(pattern), g.wrap(AdaptHandler(handler)))
}

// MethodFunc registers an HTTP method handler.
func (g *RouteGroup) MethodFunc(method, pattern string, h any) {
	g.Method(method, pattern, h)
}

// Handle registers a handler for all methods.
func (g *RouteGroup) Handle(pattern string, handler any) {
	g.parent.Handle(g.fullPattern(pattern), g.wrap(AdaptHandler(handler)))
}

// HandleFunc registers a handler for all methods.
func (g *RouteGroup) HandleFunc(pattern string, h any) {
	g.Handle(pattern, h)
}

// Get registers a GET route on this group.
func (g *RouteGroup) Get(pattern string, h any) {
	g.Method(http.MethodGet, pattern, h)
}

// Post registers a POST route on this group.
func (g *RouteGroup) Post(pattern string, h any) {
	g.Method(http.MethodPost, pattern, h)
}

// Put registers a PUT route on this group.
func (g *RouteGroup) Put(pattern string, h any) {
	g.Method(http.MethodPut, pattern, h)
}

// Patch registers a PATCH route on this group.
func (g *RouteGroup) Patch(pattern string, h any) {
	g.Method(http.MethodPatch, pattern, h)
}

// Delete registers a DELETE route on this group.
func (g *RouteGroup) Delete(pattern string, h any) {
	g.Method(http.MethodDelete, pattern, h)
}

// Options registers an OPTIONS route on this group.
func (g *RouteGroup) Options(pattern string, h any) {
	g.Method(http.MethodOptions, pattern, h)
}

// Head registers a HEAD route on this group.
func (g *RouteGroup) Head(pattern string, h any) {
	g.Method(http.MethodHead, pattern, h)
}

// Connect registers a CONNECT route on this group.
func (g *RouteGroup) Connect(pattern string, h any) {
	g.Method(http.MethodConnect, pattern, h)
}

// Trace registers a TRACE route on this group.
func (g *RouteGroup) Trace(pattern string, h any) {
	g.Method(http.MethodTrace, pattern, h)
}

// Group creates an inline nested route group under this group.
func (g *RouteGroup) Group(fn func(r Router)) Router {
	sub := NewGroup(g, "")
	if fn != nil {
		fn(sub)
	}
	return sub
}

// Route creates a nested route group with a path prefix.
func (g *RouteGroup) Route(pattern string, fn func(r Router)) Router {
	sub := NewGroup(g, pattern)
	if fn != nil {
		fn(sub)
	}
	return sub
}

// Mount mounts an http.Handler onto the group prefix.
func (g *RouteGroup) Mount(pattern string, handler http.Handler) {
	g.parent.Mount(g.fullPattern(pattern), handler)
}

// NotFound sets custom 404 handler on parent router.
func (g *RouteGroup) NotFound(h any) {
	g.parent.NotFound(h)
}

// MethodNotAllowed sets custom 405 handler on parent router.
func (g *RouteGroup) MethodNotAllowed(h any) {
	g.parent.MethodNotAllowed(h)
}

// ServeHTTP implements http.Handler interface.
func (g *RouteGroup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.parent.ServeHTTP(w, r)
}

var _ Router = (*RouteGroup)(nil)
