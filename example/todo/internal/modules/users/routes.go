package users

import (
	"github.com/ishubhamsingh2e/bourbon/router"
)

// RegisterRoutes mounts all user feature endpoints on the provided router.
func RegisterRoutes(r router.Router, service *Service) {
	ctrl := NewController(service)

	r.Route("/users", func(g router.Router) {
		g.Get("/", ctrl.Index)
		g.Post("/", ctrl.Create)
		g.Get("/{id}", ctrl.Show)
	})
}
