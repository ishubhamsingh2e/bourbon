package middleware

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// StaticOptions configures static file delivery and cache policies.
type StaticOptions struct {
	Prefix       string
	IndexFile    string
	CacheControl string
	EnableETag   bool
}

// DefaultStaticOptions provides production-ready caching defaults.
func DefaultStaticOptions() StaticOptions {
	return StaticOptions{
		Prefix:       "/static",
		IndexFile:    "index.html",
		CacheControl: "public, max-age=3600",
		EnableETag:   true,
	}
}

// Static returns a middleware that serves static assets from an fs.FS (disk or embed.FS).
func Static(fsys fs.FS, options ...StaticOptions) router.MiddlewareFunc {
	opts := DefaultStaticOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	prefix := "/" + strings.Trim(opts.Prefix, "/")
	httpFS := http.FS(fsys)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Only intercept requests matching prefix
			if !strings.HasPrefix(path, prefix) {
				next.ServeHTTP(w, r)
				return
			}

			relPath := strings.TrimPrefix(path, prefix)
			relPath = strings.TrimPrefix(relPath, "/")
			if relPath == "" {
				relPath = opts.IndexFile
			}

			file, err := fsys.Open(relPath)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil || stat.IsDir() {
				next.ServeHTTP(w, r)
				return
			}

			// ETag validation
			if opts.EnableETag {
				etag := fmt.Sprintf(`"%x-%x"`, stat.ModTime().Unix(), stat.Size())
				w.Header().Set("ETag", etag)

				if match := r.Header.Get("If-None-Match"); match == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			if opts.CacheControl != "" {
				w.Header().Set("Cache-Control", opts.CacheControl)
			}

			// Serve static file directly
			http.StripPrefix(prefix, http.FileServer(httpFS)).ServeHTTP(w, r)
		})
	}
}
