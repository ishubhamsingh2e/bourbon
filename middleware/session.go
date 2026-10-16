package middleware

import (
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
	"github.com/ishubhamsingh2e/bourbon/session"
)

// SessionOptions configures session cookie attributes and lifespan.
type SessionOptions struct {
	CookieName   string
	CookiePath   string
	CookieDomain string
	MaxAge       time.Duration
	Secure       bool
	HttpOnly     bool
	SameSite     http.SameSite
}

// DefaultSessionOptions provides secure defaults for HTTP cookie sessions.
func DefaultSessionOptions() SessionOptions {
	return SessionOptions{
		CookieName: "bourbon_session",
		CookiePath: "/",
		MaxAge:     24 * time.Hour,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   http.SameSiteLaxMode,
	}
}

// Session returns a middleware that manages stateful session cookies and persistence.
func Session(store session.Store, options ...SessionOptions) router.MiddlewareFunc {
	opts := DefaultSessionOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionID string
			if cookie, err := r.Cookie(opts.CookieName); err == nil && cookie.Value != "" {
				sessionID = cookie.Value
			}

			sess, err := store.Load(r.Context(), sessionID)
			if err != nil || sess == nil {
				sess = session.New("")
			}

			ctx := session.WithSession(r.Context(), sess)
			rWithSession := r.WithContext(ctx)

			next.ServeHTTP(w, rWithSession)

			// After handler execution, persist or destroy session
			if sess.IsDestroyed() {
				_ = store.Destroy(r.Context(), sess.ID())
				http.SetCookie(w, &http.Cookie{
					Name:     opts.CookieName,
					Value:    "",
					Path:     opts.CookiePath,
					Domain:   opts.CookieDomain,
					MaxAge:   -1,
					Secure:   opts.Secure,
					HttpOnly: opts.HttpOnly,
					SameSite: opts.SameSite,
				})
			} else if sess.IsModified() || sessionID != sess.ID() {
				_ = store.Save(r.Context(), sess, opts.MaxAge)
				http.SetCookie(w, &http.Cookie{
					Name:     opts.CookieName,
					Value:    sess.ID(),
					Path:     opts.CookiePath,
					Domain:   opts.CookieDomain,
					MaxAge:   int(opts.MaxAge.Seconds()),
					Secure:   opts.Secure,
					HttpOnly: opts.HttpOnly,
					SameSite: opts.SameSite,
				})
			}
		})
	}
}
