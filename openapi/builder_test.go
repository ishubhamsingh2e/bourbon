package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/core"
)

func TestOpenAPIBuilderAndHandlers(t *testing.T) {
	builder := NewBuilder("Bourbon API", "v1.0.0", "Production API Specification")

	builder.AddRoute(http.MethodGet, "/users", "List Users", "Returns paginated list of users", map[int]string{
		200: "User list",
		401: "Unauthenticated",
	})

	builder.AddRoute(http.MethodPost, "/users", "Create User", "Creates a new user", map[int]string{
		201: "User created",
		422: "Validation error",
	})

	// 1. JSON Export
	jsonBytes, err := builder.BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON failed: %v", err)
	}

	var spec Spec
	if err := json.Unmarshal(jsonBytes, &spec); err != nil {
		t.Fatalf("Unmarshal OpenAPI failed: %v", err)
	}

	if spec.Info.Title != "Bourbon API" {
		t.Errorf("title mismatch: %s", spec.Info.Title)
	}
	if _, exists := spec.Paths["/users"]["get"]; !exists {
		t.Errorf("missing GET /users in spec")
	}
	if _, exists := spec.Paths["/users"]["post"]; !exists {
		t.Errorf("missing POST /users in spec")
	}

	// 2. Swagger UI Handler
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	c := core.NewContext(w, req)

	_ = builder.SwaggerUIHandler("/openapi.json")(c)
	if w.Code != http.StatusOK {
		t.Errorf("Swagger UI code = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "SwaggerUIBundle") {
		t.Errorf("Swagger UI HTML mismatch: %s", w.Body.String())
	}
}
