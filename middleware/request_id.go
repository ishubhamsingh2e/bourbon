package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

const (
	// HeaderXRequestID is the HTTP header key used for request correlation.
	HeaderXRequestID = "X-Request-ID"
)

type reqIDContextKey struct{}

var requestIDKey = reqIDContextKey{}

// GetReqID returns a request ID from the given context if one exists.
func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}

// RequestID extracts or generates an X-Request-ID header and injects it into context.
func RequestID() router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(HeaderXRequestID)
			if reqID == "" {
				reqID = generateRequestID()
			}

			w.Header().Set(HeaderXRequestID, reqID)
			ctx := context.WithValue(r.Context(), requestIDKey, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
