package router

import (
	"net/http"
	"strings"
)

// ResourceController defines the standard CRUD interface for RESTful resource routing.
type ResourceController interface {
	Index(w http.ResponseWriter, r *http.Request)
	Show(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// Resource binds RESTful resource endpoints to the controller actions.
func Resource(r Router, path string, ctrl ResourceController) {
	basePath := "/" + strings.Trim(path, "/")
	itemPath := basePath + "/{id}"

	r.Get(basePath, ctrl.Index)
	r.Post(basePath, ctrl.Create)
	r.Get(itemPath, ctrl.Show)
	r.Put(itemPath, ctrl.Update)
	r.Delete(itemPath, ctrl.Delete)
}
