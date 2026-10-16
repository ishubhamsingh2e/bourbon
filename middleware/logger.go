package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	bytesOut   int
}

func (rw *responseWriterInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesOut += n
	return n, err
}

// Logger emits structured log events for incoming HTTP requests.
func Logger(logger ...*slog.Logger) router.MiddlewareFunc {
	var l *slog.Logger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			interceptor := &responseWriterInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(interceptor, r)

			latency := time.Since(start)

			l.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", interceptor.statusCode),
				slog.Duration("latency", latency),
				slog.Int("bytes", interceptor.bytesOut),
			)
		})
	}
}
