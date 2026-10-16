package router

import (
	"fmt"
	"net/http"
	"sync"
)

// HandlerAdapter converts a custom handler type into a standard http.Handler.
type HandlerAdapter func(any) (http.Handler, bool)

var (
	adaptersMu sync.RWMutex
	adapters   []HandlerAdapter
)

// RegisterHandlerAdapter registers an adapter to convert custom handler signatures to http.Handler.
func RegisterHandlerAdapter(adapter HandlerAdapter) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	adapters = append(adapters, adapter)
}

// AdaptHandler converts any supported handler type into an http.Handler.
func AdaptHandler(h any) http.Handler {
	if h == nil {
		panic("router: handler cannot be nil")
	}
	if handler, ok := h.(http.Handler); ok {
		return handler
	}
	if fn, ok := h.(func(http.ResponseWriter, *http.Request)); ok {
		return http.HandlerFunc(fn)
	}

	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	for _, adapter := range adapters {
		if handler, ok := adapter(h); ok {
			return handler
		}
	}

	panic(fmt.Sprintf("router: unsupported handler type: %T", h))
}

// Chain compiles a sequence of middleware handlers into a single reusable middleware.
func Chain(middlewares ...MiddlewareFunc) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		wrapped := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			wrapped = middlewares[i](wrapped)
		}
		return wrapped
	}
}

// AdaptMiddleware wraps a standard middleware function into MiddlewareFunc.
func AdaptMiddleware(mw func(http.Handler) http.Handler) MiddlewareFunc {
	return mw
}
