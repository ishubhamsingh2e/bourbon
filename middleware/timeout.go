package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type timeoutWriter struct {
	underlying  http.ResponseWriter
	mu          sync.Mutex
	header      http.Header
	timedOut    bool
	wroteHeader bool
	code        int
}

func newTimeoutWriter(w http.ResponseWriter) *timeoutWriter {
	return &timeoutWriter{
		underlying: w,
		header:     make(http.Header),
		code:       http.StatusOK,
	}
}

func (tw *timeoutWriter) Header() http.Header {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	return tw.header
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.wroteHeader = true
	tw.code = code
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.wroteHeader = true
	}
	// Copy headers to underlying
	for k, vv := range tw.header {
		for _, v := range vv {
			tw.underlying.Header().Add(k, v)
		}
	}
	tw.underlying.WriteHeader(tw.code)
	return tw.underlying.Write(b)
}

// Timeout returns a middleware that cancels the request context if execution exceeds the duration.
func Timeout(duration time.Duration) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			tw := newTimeoutWriter(w)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				tw.mu.Lock()
				if !tw.wroteHeader {
					for k, vv := range tw.header {
						for _, v := range vv {
							tw.underlying.Header().Add(k, v)
						}
					}
					tw.underlying.WriteHeader(tw.code)
					tw.wroteHeader = true
				}
				tw.mu.Unlock()

			case <-ctx.Done():
				tw.mu.Lock()
				tw.timedOut = true
				tw.underlying.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
				tw.underlying.WriteHeader(http.StatusGatewayTimeout)
				_, _ = tw.underlying.Write([]byte(`{"type":"about:blank","title":"Gateway Timeout","status":504,"detail":"Request execution timed out."}`))
				tw.wroteHeader = true
				tw.mu.Unlock()
			}
		})
	}
}
