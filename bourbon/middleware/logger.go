package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/bourbon/logging"
	"go.uber.org/zap"
)

// Logger middleware logs incoming HTTP requests with method, path, status code, duration, and client IP
func Logger(logger *logging.Logger, errorStore *logging.ErrorStore) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Human-readable console output for development
			statusColor := getStatusColor(wrapped.statusCode)
			methodColor := getMethodColor(r.Method)
			
			fmt.Printf("%s %s%-6s\x1b[0m | %s%3d\x1b[0m | %10s | %s\n",
				time.Now().Format("15:04:05"),
				methodColor,
				r.Method,
				statusColor,
				wrapped.statusCode,
				duration.Round(time.Millisecond),
				r.URL.Path,
			)

			// Store server errors (5xx) in database
			if wrapped.statusCode >= 500 && errorStore != nil {
				// Only log errors to structured logger (for file/database)
				logger.HTTP(r.Method, r.URL.Path, wrapped.statusCode, duration,
					zap.String("ip", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
				)
				
				errorLog := &logging.ErrorLog{
					Timestamp: start,
					Level:     "error",
					Message:   fmt.Sprintf("HTTP %d: %s %s", wrapped.statusCode, r.Method, r.URL.Path),
					Method:    r.Method,
					Path:      r.URL.Path,
					Status:    wrapped.statusCode,
					IP:        r.RemoteAddr,
					UserAgent: r.UserAgent(),
				}
				_ = errorStore.Store(errorLog)
			}
		})
	}
}

// getStatusColor returns ANSI color code based on HTTP status
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\x1b[32m" // Green
	case status >= 300 && status < 400:
		return "\x1b[36m" // Cyan
	case status >= 400 && status < 500:
		return "\x1b[33m" // Yellow
	case status >= 500:
		return "\x1b[31m" // Red
	default:
		return "\x1b[37m" // White
	}
}

// getMethodColor returns ANSI color code based on HTTP method
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\x1b[34m" // Blue
	case "POST":
		return "\x1b[32m" // Green
	case "PUT":
		return "\x1b[33m" // Yellow
	case "DELETE":
		return "\x1b[31m" // Red
	case "PATCH":
		return "\x1b[35m" // Magenta
	default:
		return "\x1b[37m" // White
	}
}
