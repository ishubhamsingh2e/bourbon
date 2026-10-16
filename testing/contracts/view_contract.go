package contracts

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/view"
)

// VerifyViewRendererContract executes a standard compliance test suite against any view.Renderer implementation.
func VerifyViewRendererContract(t *testing.T, renderer view.Renderer, viewName string, data any, expectedSubstring string) {
	t.Run("Render_Success", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Render(&buf, viewName, data)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, expectedSubstring) {
			t.Errorf("rendered output does not contain expected substring %q, got: %s", expectedSubstring, output)
		}
	})

	t.Run("Render_Missing_View_Error", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Render(&buf, "non_existent_view_12345.html", nil)
		if err == nil {
			t.Errorf("expected error for non-existent view, got nil")
		}
	})
}
