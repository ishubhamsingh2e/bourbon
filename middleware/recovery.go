// Package middleware provides production-grade HTTP middleware components for logging, panic recovery, CORS, tracing, and timeout.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// Recovery returns a middleware that recovers from panics, logs stack traces, and sends a 500 RFC 7807 response.
func Recovery(logger ...*slog.Logger) router.MiddlewareFunc {
	var l *slog.Logger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					l.Error("panic recovered in HTTP handler",
						slog.Any("error", rec),
						slog.String("path", r.URL.Path),
						slog.String("method", r.Method),
						slog.String("stack", stack),
					)

					w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, `{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"An unexpected server error occurred."}`)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
