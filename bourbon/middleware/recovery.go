package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ishubhamsingh2e/bourbon/bourbon/logging"
	"go.uber.org/zap"
)

// Recovery middleware recovers from panics in the request handling chain and logs the error with stack trace
func Recovery(logger *logging.Logger, errorStore *logging.ErrorStore) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := string(debug.Stack())

					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.String("stack", stack),
					)

					// Store panic in database
					if errorStore != nil {
						errorLog := &logging.ErrorLog{
							Timestamp: time.Now(),
							Level:     "panic",
							Message:   fmt.Sprintf("Panic: %v", err),
							Method:    r.Method,
							Path:      r.URL.Path,
							Status:    http.StatusInternalServerError,
							IP:        r.RemoteAddr,
							UserAgent: r.UserAgent(),
							Stack:     stack,
						}
						_ = errorStore.Store(errorLog)
					}

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
