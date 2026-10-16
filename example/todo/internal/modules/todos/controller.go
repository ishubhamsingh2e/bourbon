package todos

import (
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type TodoController struct {
	service *TodoService
}

func NewTodoController(s *TodoService) *TodoController {
	if s == nil {
		s = NewTodoService(nil)
	}
	return &TodoController{service: s}
}

func (ctrl *TodoController) Index(w http.ResponseWriter, r *http.Request) {
	items, err := ctrl.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": items,
	})
}

func (ctrl *TodoController) Show(w http.ResponseWriter, r *http.Request) {
	id := router.URLParam(r, "id")
	item, err := ctrl.service.GetByID(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Todo not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (ctrl *TodoController) Create(w http.ResponseWriter, r *http.Request) {
	var item Todo
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ctrl.service.Create(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}
