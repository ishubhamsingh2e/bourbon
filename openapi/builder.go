// Package openapi provides a code-first OpenAPI 3.0 specification builder and Swagger UI renderer.
package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/core"
)

// Operation describes an HTTP operation on an API path.
type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Response models an HTTP response definition in OpenAPI 3.0.
type Response struct {
	Description string `json:"description"`
}

// Spec represents the top-level OpenAPI 3.0 specification.
type Spec struct {
	OpenAPI string                          `json:"openapi"`
	Info    Info                            `json:"info"`
	Paths   map[string]map[string]Operation `json:"paths"`
}

// Info models OpenAPI info metadata.
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Builder constructs OpenAPI 3.0 documents.
type Builder struct {
	mu   sync.RWMutex
	spec Spec
}

// NewBuilder initializes a new OpenAPI specification builder.
func NewBuilder(title, version, description string) *Builder {
	return &Builder{
		spec: Spec{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       title,
				Version:     version,
				Description: description,
			},
			Paths: make(map[string]map[string]Operation),
		},
	}
}

// AddRoute registers an endpoint operation into the OpenAPI spec.
func (b *Builder) AddRoute(method, path, summary, description string, responses map[int]string) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	httpMethod := strings.ToLower(method)
	if _, exists := b.spec.Paths[path]; !exists {
		b.spec.Paths[path] = make(map[string]Operation)
	}

	respMap := make(map[string]Response)
	for code, desc := range responses {
		respMap[fmt.Sprintf("%d", code)] = Response{Description: desc}
	}
	if len(respMap) == 0 {
		respMap["200"] = Response{Description: "Successful operation"}
	}

	b.spec.Paths[path][httpMethod] = Operation{
		Summary:     summary,
		Description: description,
		Responses:   respMap,
	}

	return b
}

// BuildJSON exports the OpenAPI spec as formatted JSON.
func (b *Builder) BuildJSON() ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.MarshalIndent(b.spec, "", "  ")
}

// OpenAPIJSONHandler returns a Bourbon handler func serving the raw openapi.json document.
func (b *Builder) OpenAPIJSONHandler() core.HandlerFunc {
	return func(c *core.Context) error {
		jsonBytes, err := b.BuildJSON()
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/json; charset=utf-8", jsonBytes)
	}
}

// SwaggerUIHandler returns an interactive Swagger UI HTML page pointed at specURL.
func (b *Builder) SwaggerUIHandler(specURL string) core.HandlerFunc {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <title>%s - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '%s',
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>`, b.spec.Info.Title, specURL)

	return func(c *core.Context) error {
		return c.Blob(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}
