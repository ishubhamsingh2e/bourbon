package testutil

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/controller"
	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
)

type UserPayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *UserPayload) Validate() error {
	if u.Name == "" {
		return controller.NewValidationError("/api/v1/users", []controller.InvalidParam{
			{Field: "name", Tag: "required", Message: "Name cannot be empty"},
		})
	}
	return nil
}

func TestEndToEndIntegration(t *testing.T) {
	r := router.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	})

	router.Version(r, "v1", func(v router.Router) {
		v.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			var payload UserPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(controller.NewProblemDetail(http.StatusBadRequest, "Bad Request", err.Error()))
				return
			}
			if err := payload.Validate(); err != nil {
				if prob, ok := err.(*controller.ProblemDetail); ok {
					w.Header().Set("Content-Type", "application/problem+json")
					w.WriteHeader(prob.Status)
					_ = json.NewEncoder(w).Encode(prob)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    101,
				"name":  payload.Name,
				"email": payload.Email,
			})
		})
	})

	testApp := NewTestApp(t, app.WithRouter(r))
	client := testApp.Client()

	// 1. Test GET /ping
	client.Get("/ping").
		AssertStatus(http.StatusOK).
		AssertHeaderNotEmpty("X-Request-ID").
		AssertJSONPath("status", "pong")

	// 2. Test POST /api/v1/users Success
	client.PostJSON("/api/v1/users", UserPayload{Name: "Alice", Email: "alice@example.com"}).
		AssertStatus(http.StatusCreated).
		AssertJSONPath("id", 101).
		AssertJSONPath("name", "Alice")

	// 3. Test POST /api/v1/users Validation Failure
	client.PostJSON("/api/v1/users", UserPayload{Name: "", Email: "bad@example.com"}).
		AssertStatus(http.StatusUnprocessableEntity).
		AssertJSONPath("title", "Validation Failed")
}
