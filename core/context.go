// Package core defines the minimal application shell interfaces, DI container contracts, and lifecycle provider standards.
package core

import (
	stdctx "context"
	"net/http"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/auth"
	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/session"
	"github.com/ishubhamsingh2e/bourbon/view"
)

const (
	contextKeyDB        = "__bourbon_db__"
	contextKeyCache     = "__bourbon_cache__"
	contextKeySession   = "__bourbon_session__"
	contextKeyAuth      = "__bourbon_auth__"
	contextKeyValidator = "__bourbon_validator__"
	contextKeyView      = "__bourbon_view__"
)

// Context wraps http.ResponseWriter and *http.Request with ergonomic helpers.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Ctx      stdctx.Context

	mu     sync.RWMutex
	params map[string]string
	store  map[string]any
}

// NewContext initializes a new Context wrapping the given ResponseWriter and Request.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Ctx:      r.Context(),
		params:   make(map[string]string),
		store:    make(map[string]any),
	}
}

// Set stores a key-value pair scoped to this request context.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Get retrieves a key-value pair from the request-scoped store.
func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

// DB returns the bound database pool instance or nil if not registered.
func (c *Context) DB() database.DB {
	if val, ok := c.Get(contextKeyDB); ok {
		if db, ok := val.(database.DB); ok {
			return db
		}
	}
	return nil
}

// SetDB binds a database pool to this request context.
func (c *Context) SetDB(db database.DB) {
	c.Set(contextKeyDB, db)
}

// Cache returns the bound cache instance or nil if not registered.
func (c *Context) Cache() cache.Cache {
	if val, ok := c.Get(contextKeyCache); ok {
		if ca, ok := val.(cache.Cache); ok {
			return ca
		}
	}
	return nil
}

// SetCache binds a cache instance to this request context.
func (c *Context) SetCache(ca cache.Cache) {
	c.Set(contextKeyCache, ca)
}

// Session returns the active stateful session or nil if middleware is not attached.
func (c *Context) Session() *session.Session {
	if val, ok := c.Get(contextKeySession); ok {
		if sess, ok := val.(*session.Session); ok {
			return sess
		}
	}
	return nil
}

// SetSession binds an active session to this request context.
func (c *Context) SetSession(sess *session.Session) {
	c.Set(contextKeySession, sess)
}

// Auth returns the authenticated Principal or nil if not authenticated.
func (c *Context) Auth() auth.Principal {
	if val, ok := c.Get(contextKeyAuth); ok {
		if p, ok := val.(auth.Principal); ok {
			return p
		}
	}
	return nil
}

// SetAuth binds an authenticated Principal to this request context.
func (c *Context) SetAuth(p auth.Principal) {
	c.Set(contextKeyAuth, p)
}

// RequireAuth returns the authenticated Principal, or auth.ErrUnauthenticated if absent.
func (c *Context) RequireAuth() (auth.Principal, error) {
	p := c.Auth()
	if p == nil {
		return nil, auth.ErrUnauthenticated
	}
	return p, nil
}

// RequireCan verifies if the authenticated principal is allowed to perform action on resource.
func (c *Context) RequireCan(enforcer auth.Enforcer, action string, resource ...any) error {
	p := c.Auth()
	if p == nil {
		return auth.ErrUnauthenticated
	}
	var res any
	if len(resource) > 0 {
		res = resource[0]
	}
	can, err := enforcer.Can(c.Ctx, p, action, res)
	if err != nil {
		return err
	}
	if !can {
		return auth.ErrUnauthorized
	}
	return nil
}

// ViewRenderer returns the registered view renderer or nil if not bound.
func (c *Context) ViewRenderer() view.Renderer {
	if val, ok := c.Get(contextKeyView); ok {
		if r, ok := val.(view.Renderer); ok {
			return r
		}
	}
	return nil
}

// SetViewRenderer binds a view renderer to this request context.
func (c *Context) SetViewRenderer(r view.Renderer) {
	c.Set(contextKeyView, r)
}
