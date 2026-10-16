package otel

import (
	"fmt"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// Middleware returns a Bourbon router middleware that instruments HTTP requests with OpenTelemetry spans.
func Middleware(tracer *Tracer) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := "/"
			method := "GET"
			if r != nil && r.URL != nil {
				path = r.URL.Path
				method = r.Method
			}

			spanName := fmt.Sprintf("HTTP %s %s", method, path)
			ctx, span := tracer.Start(r.Context(), spanName)
			defer span.End()

			span.SetAttribute("http.method", method)
			span.SetAttribute("http.target", path)

			w.Header().Set("traceparent", span.TraceParentHeader())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
