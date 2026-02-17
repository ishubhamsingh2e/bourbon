package http

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context) error

type Router struct {
	mux            *http.ServeMux
	routes         []Route
	middlewares    []MiddlewareFunc
	TemplateEngine *TemplateEngine
	staticHandlers map[string]http.Handler
}

type Route struct {
	Method  string
	Pattern string
	Handler HandlerFunc
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

func NewRouter() *Router {
	return &Router{
		mux:            http.NewServeMux(),
		routes:         make([]Route, 0),
		middlewares:    make([]MiddlewareFunc, 0),
		TemplateEngine: nil,
		staticHandlers: make(map[string]http.Handler),
	}
}

func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware...)
}

func (r *Router) Get(pattern string, handler HandlerFunc) {
	r.addRoute("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler HandlerFunc) {
	r.addRoute("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler HandlerFunc) {
	r.addRoute("PUT", pattern, handler)
}

func (r *Router) Patch(pattern string, handler HandlerFunc) {
	r.addRoute("PATCH", pattern, handler)
}

func (r *Router) Delete(pattern string, handler HandlerFunc) {
	r.addRoute("DELETE", pattern, handler)
}

func (r *Router) addRoute(method, pattern string, handler HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})

	key := fmt.Sprintf("%s %s", method, pattern)
	r.mux.HandleFunc(key, r.wrapHandler(method, pattern, handler))
}

func (r *Router) wrapHandler(method, pattern string, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := &Context{
			Writer:         w,
			Request:        req,
			Params:         extractParams(pattern, req.URL.Path),
			store:          make(map[string]interface{}),
			TemplateEngine: r.TemplateEngine,
		}

		finalHandler := handler
		for i := len(r.middlewares) - 1; i >= 0; i-- {
			finalHandler = r.middlewares[i](finalHandler)
		}

		if err := finalHandler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (r *Router) Static(prefix, root string) {
	fs := http.FileServer(http.Dir(root))
	handler := http.StripPrefix(prefix, fs)

	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	r.staticHandlers[prefix] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for prefix, handler := range r.staticHandlers {
		if strings.HasPrefix(req.URL.Path, prefix) {
			handler.ServeHTTP(w, req)
			return
		}
	}

	r.mux.ServeHTTP(w, req)
}

func (r *Router) GetRoutes() []Route {
	return r.routes
}

func extractParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	for i, part := range patternParts {
		if i >= len(pathParts) {
			break
		}
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = pathParts[i]
		}
	}

	return params
}

type Group struct {
	router      *Router
	prefix      string
	middlewares []MiddlewareFunc
}

func (r *Router) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	return &Group{
		router:      r,
		prefix:      prefix,
		middlewares: middleware,
	}
}

// cleanPath ensures the path is clean and doesn't have double slashes
func cleanPath(prefix, pattern string) string {
	combined := prefix + pattern
	// Use path.Clean to remove any double slashes and clean the path
	cleaned := path.Clean(combined)
	// path.Clean removes trailing slashes, but we might want to keep them for root
	// However, for pattern matching, we should use the cleaned version
	return cleaned
}

func (g *Group) Get(pattern string, handler HandlerFunc) {
	finalHandler := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		finalHandler = g.middlewares[i](finalHandler)
	}
	g.router.Get(cleanPath(g.prefix, pattern), finalHandler)
}

func (g *Group) Post(pattern string, handler HandlerFunc) {
	finalHandler := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		finalHandler = g.middlewares[i](finalHandler)
	}
	g.router.Post(cleanPath(g.prefix, pattern), finalHandler)
}

func (g *Group) Put(pattern string, handler HandlerFunc) {
	finalHandler := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		finalHandler = g.middlewares[i](finalHandler)
	}
	g.router.Put(cleanPath(g.prefix, pattern), finalHandler)
}

func (g *Group) Patch(pattern string, handler HandlerFunc) {
	finalHandler := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		finalHandler = g.middlewares[i](finalHandler)
	}
	g.router.Patch(cleanPath(g.prefix, pattern), finalHandler)
}

func (g *Group) Delete(pattern string, handler HandlerFunc) {
	finalHandler := handler
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		finalHandler = g.middlewares[i](finalHandler)
	}
	g.router.Delete(cleanPath(g.prefix, pattern), finalHandler)
}

func (r *Router) Resource(path string, controller interface{}) {
}
