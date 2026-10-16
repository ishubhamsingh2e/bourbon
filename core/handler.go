package core

import (
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

// Wrap converts a Bourbon HandlerFunc into a standard http.HandlerFunc.
func Wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r)
		if err := h(c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func init() {
	router.RegisterHandlerAdapter(func(h any) (http.Handler, bool) {
		if fn, ok := h.(HandlerFunc); ok {
			return Wrap(fn), true
		}
		if fn, ok := h.(func(*Context) error); ok {
			return Wrap(fn), true
		}
		return nil, false
	})
}
