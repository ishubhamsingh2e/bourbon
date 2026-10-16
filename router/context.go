package router

import (
	"context"
	"net/http"
)

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "router context value " + k.name
}

var routeCtxKey = &contextKey{"RouteContext"}

// RouteParams holds URL parameter keys and values extracted from path patterns.
type RouteParams struct {
	Keys   []string
	Values []string
}

// Add appends a parameter key-value pair.
func (p *RouteParams) Add(key, value string) {
	p.Keys = append(p.Keys, key)
	p.Values = append(p.Values, value)
}

// Get returns the parameter value matching key, searching in reverse order.
func (p *RouteParams) Get(key string) string {
	for i := len(p.Keys) - 1; i >= 0; i-- {
		if p.Keys[i] == key {
			return p.Values[i]
		}
	}
	return ""
}

// Context is the routing context attached to http.Request during routing dispatch.
type Context struct {
	URLParams    RouteParams
	RoutePattern string
	RoutePath    string
}

// URLParam returns the URL parameter for key from this context.
func (c *Context) URLParam(key string) string {
	return c.URLParams.Get(key)
}

// Reset clears the context for reuse.
func (c *Context) Reset() {
	c.URLParams.Keys = c.URLParams.Keys[:0]
	c.URLParams.Values = c.URLParams.Values[:0]
	c.RoutePattern = ""
	c.RoutePath = ""
}

// NewRouteContext returns a freshly initialized routing Context.
func NewRouteContext() *Context {
	return &Context{
		URLParams: RouteParams{
			Keys:   make([]string, 0, 8),
			Values: make([]string, 0, 8),
		},
	}
}

// WithRouteContext returns a copy of parent context with the route Context embedded.
func WithRouteContext(parent context.Context, rctx *Context) context.Context {
	return context.WithValue(parent, routeCtxKey, rctx)
}

// RouteContext returns the route Context stored in ctx, or nil if none exists.
func RouteContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	val, _ := ctx.Value(routeCtxKey).(*Context)
	return val
}

// URLParam extracts a URL parameter by key from the request's context.
func URLParam(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	return URLParamFromCtx(r.Context(), key)
}

// URLParamFromCtx extracts a URL parameter by key from the context.
func URLParamFromCtx(ctx context.Context, key string) string {
	if rctx := RouteContext(ctx); rctx != nil {
		return rctx.URLParam(key)
	}
	return ""
}
