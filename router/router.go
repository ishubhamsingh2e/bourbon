// Package router provides the native high-performance Radix Trie routing engine with Chi-inspired ergonomics for Bourbon.
package router

import (
	"net/http"
)

// MiddlewareFunc wraps an http.Handler to provide pre/post request interception.
type MiddlewareFunc func(http.Handler) http.Handler

// Router defines the primary routing contract for Bourbon applications, featuring Chi-identical routing and middleware mechanics.
type Router interface {
	// Middleware & Scoping
	Use(middlewares ...MiddlewareFunc)
	With(middlewares ...MiddlewareFunc) Router

	// Route Hierarchy & Mounting
	Group(fn func(r Router)) Router
	Route(pattern string, fn func(r Router)) Router
	Mount(pattern string, handler http.Handler)

	// Primary Route Registrations
	Handle(pattern string, handler any)
	HandleFunc(pattern string, h any)
	Method(method, pattern string, handler any)
	MethodFunc(method, pattern string, h any)

	// HTTP Method Shortcuts
	Connect(pattern string, h any)
	Delete(pattern string, h any)
	Get(pattern string, h any)
	Head(pattern string, h any)
	Options(pattern string, h any)
	Patch(pattern string, h any)
	Post(pattern string, h any)
	Put(pattern string, h any)
	Trace(pattern string, h any)

	// Custom Fallback Handlers
	NotFound(h any)
	MethodNotAllowed(h any)

	// HTTP Handler Interface
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
