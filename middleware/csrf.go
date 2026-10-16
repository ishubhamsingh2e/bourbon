package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// CSRFOptions configures the double-submit cookie CSRF defense middleware.
type CSRFOptions struct {
	CookieName   string
	HeaderName   string
	FormField    string
	CookiePath   string
	CookieDomain string
	MaxAge       time.Duration
	Secure       bool
	HttpOnly     bool
	SameSite     http.SameSite
}

// DefaultCSRFOptions provides secure defaults for CSRF protection.
func DefaultCSRFOptions() CSRFOptions {
	return CSRFOptions{
		CookieName: "bourbon_csrf",
		HeaderName: "X-CSRF-Token",
		FormField:  "_csrf",
		CookiePath: "/",
		MaxAge:     24 * time.Hour,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   http.SameSiteLaxMode,
	}
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CSRF returns a middleware implementing Double-Submit Cookie CSRF protection.
func CSRF(options ...CSRFOptions) router.MiddlewareFunc {
	opts := DefaultCSRFOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cookieToken string
			if cookie, err := r.Cookie(opts.CookieName); err == nil && cookie.Value != "" {
				cookieToken = cookie.Value
			}

			// If no cookie exists, generate a new token and attach it
			if cookieToken == "" {
				cookieToken = generateCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name:     opts.CookieName,
					Value:    cookieToken,
					Path:     opts.CookiePath,
					Domain:   opts.CookieDomain,
					MaxAge:   int(opts.MaxAge.Seconds()),
					Secure:   opts.Secure,
					HttpOnly: opts.HttpOnly,
					SameSite: opts.SameSite,
				})
			}

			// Exempt safe HTTP methods
			method := r.Method
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodTrace {
				next.ServeHTTP(w, r)
				return
			}

			// Extract submitted token from headers or form body
			submittedToken := r.Header.Get(opts.HeaderName)
			if submittedToken == "" {
				submittedToken = r.FormValue(opts.FormField)
			}

			if submittedToken == "" || subtle.ConstantTimeCompare([]byte(submittedToken), []byte(cookieToken)) != 1 {
				w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				_, _ = fmt.Fprintf(w, `{"type":"https://bourbon.dev/errors/csrf-token-mismatch","title":"CSRF Token Mismatch","status":403,"detail":"The request failed CSRF protection check.","instance":"%s"}`, r.URL.Path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
