package todos

import (
	"github.com/ishubhamsingh2e/bourbon/router"
)

// RegisterRoutes registers Todo application endpoints on the router.
func RegisterRoutes(r router.Router) {
	ctrl := NewTodoController(nil)

	r.Route("/todos", func(g router.Router) {
		g.Get("/", ctrl.Index)
		g.Post("/", ctrl.Create)
		g.Get("/{id}", ctrl.Show)
	})
}
