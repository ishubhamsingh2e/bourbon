package auth

import (
	"context"
	"net/http"
	"strings"
)

// TokenGuard authenticates requests by validating Bearer tokens extracted from HTTP headers.
type TokenGuard struct {
	provider TokenProvider
	header   string
	prefix   string
}

// NewTokenGuard constructs an initialized TokenGuard.
func NewTokenGuard(provider TokenProvider, header ...string) *TokenGuard {
	hdr := "Authorization"
	if len(header) > 0 && header[0] != "" {
		hdr = header[0]
	}
	return &TokenGuard{
		provider: provider,
		header:   hdr,
		prefix:   "Bearer ",
	}
}

// Authenticate extracts the bearer token from the header and validates it using the TokenProvider.
func (g *TokenGuard) Authenticate(ctx context.Context, r *http.Request) (Principal, error) {
	val := r.Header.Get(g.header)
	if val == "" {
		return nil, ErrUnauthenticated
	}

	if !strings.HasPrefix(val, g.prefix) {
		return nil, ErrUnauthenticated
	}

	token := strings.TrimPrefix(val, g.prefix)
	return g.provider.Validate(token)
}
