package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/session"
)

// SessionGuard authenticates requests by retrieving principal data from an active HTTP session.
type SessionGuard struct {
	store      session.Store
	cookieName string
	sessionKey string
}

// NewSessionGuard initializes a session authentication guard.
func NewSessionGuard(store session.Store, cookieName string, sessionKey ...string) *SessionGuard {
	key := "auth_principal"
	if len(sessionKey) > 0 && sessionKey[0] != "" {
		key = sessionKey[0]
	}
	if cookieName == "" {
		cookieName = "bourbon_session"
	}
	return &SessionGuard{
		store:      store,
		cookieName: cookieName,
		sessionKey: key,
	}
}

// Authenticate reads the session cookie and extracts the stored Principal.
func (g *SessionGuard) Authenticate(ctx context.Context, r *http.Request) (Principal, error) {
	cookie, err := r.Cookie(g.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, ErrUnauthenticated
	}

	sess, err := g.store.Load(ctx, cookie.Value)
	if err != nil || sess == nil {
		return nil, ErrUnauthenticated
	}

	raw, exists := sess.Get(g.sessionKey)
	if !exists || raw == nil {
		return nil, ErrUnauthenticated
	}

	switch v := raw.(type) {
	case string:
		// Attempt JSON unmarshal or treat as subject ID
		var p UserPrincipal
		if err := json.Unmarshal([]byte(v), &p); err == nil && p.Subject != "" {
			return &p, nil
		}
		return NewPrincipal(v), nil
	case map[string]any:
		data, err := json.Marshal(v)
		if err == nil {
			var p UserPrincipal
			if err := json.Unmarshal(data, &p); err == nil && p.Subject != "" {
				return &p, nil
			}
		}
	}

	return nil, ErrUnauthenticated
}
