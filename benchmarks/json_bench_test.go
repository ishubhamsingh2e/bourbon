package benchmarks

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/controller"
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/validation"
)

type BenchmarkPayload struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Age      int      `json:"age"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

func BenchmarkJSONBindingAndResponse(b *testing.B) {
	payload := []byte(`{"id":"usr_100","email":"benchmark@example.com","name":"Gopher Benchmark","age":30,"roles":["admin","developer"],"is_active":true}`)

	b.ReportAllocs()

	for b.Loop() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/bench", bytes.NewReader(payload))
		c := core.NewContext(w, req)

		var p BenchmarkPayload
		if err := c.BindJSON(&p); err != nil {
			b.Fatal(err)
		}

		_ = c.JSON(http.StatusOK, p)
	}
}

func BenchmarkValidationExecution(b *testing.B) {
	validator := validation.New().
		Field("email", validation.Required(), validation.Email()).
		Field("name", validation.Required(), validation.Min(2)).
		Field("age", validation.Required(), validation.Min(18))

	data := map[string]any{
		"email": "test@example.com",
		"name":  "Alice",
		"age":   25,
	}

	ctx := context.Background()

	b.ReportAllocs()

	for b.Loop() {
		if err := validator.ValidateMap(ctx, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProblemDetailsSerialization(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		prob := controller.NewProblemDetails(
			"https://bourbon.dev/errors/validation-failed",
			"Validation Failed",
			http.StatusUnprocessableEntity,
			"Request body contains invalid arguments",
			"/api/v1/users",
		)
		prob.InvalidParams = []controller.InvalidParam{
			{Field: "email", Tag: "email", Message: "Invalid email format"},
			{Field: "age", Tag: "min", Message: "Must be at least 18"},
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
		c := core.NewContext(w, req)
		_ = c.JSON(http.StatusUnprocessableEntity, prob)
	}
}
