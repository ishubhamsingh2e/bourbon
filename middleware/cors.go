package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// CORSOptions configures cross-origin resource sharing headers and behavior.
type CORSOptions struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSOptions provides sensible, secure CORS defaults.
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// CORS returns a middleware configuring CORS headers and handling preflight OPTIONS requests.
func CORS(options ...CORSOptions) router.MiddlewareFunc {
	opts := DefaultCORSOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	allowMethods := strings.Join(opts.AllowMethods, ", ")
	allowHeaders := strings.Join(opts.AllowHeaders, ", ")
	exposeHeaders := strings.Join(opts.ExposeHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check origin matching
			allowedOrigin := ""
			for _, o := range opts.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = o
					break
				}
			}

			if allowedOrigin != "" {
				if allowedOrigin == "*" && opts.AllowCredentials && origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				}

				if opts.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if exposeHeaders != "" {
					w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
				}
			}

			// Preflight request
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
				if allowHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				}
				if opts.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
