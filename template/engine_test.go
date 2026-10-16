package template

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

func TestTemplateEngineRendering(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<h1>Hello {{.Name | upper}}</h1>`),
		},
		"about.html": &fstest.MapFile{
			Data: []byte(`<p>{{customFunc .Message}}</p>`),
		},
	}

	engine := New(mockFS, "*.html", false)
	engine.AddFunc("customFunc", func(s string) string {
		return "Custom: " + s
	})

	// 1. Render index.html
	var buf1 bytes.Buffer
	err := engine.Render(&buf1, "index.html", map[string]any{"Name": "bourbon"})
	if err != nil {
		t.Fatalf("Render index.html failed: %v", err)
	}

	if !strings.Contains(buf1.String(), "<h1>Hello BOURBON</h1>") {
		t.Errorf("rendered output mismatch: %s", buf1.String())
	}

	// 2. Render about.html with customFunc
	var buf2 bytes.Buffer
	err = engine.Render(&buf2, "about.html", map[string]any{"Message": "Framework"})
	if err != nil {
		t.Fatalf("Render about.html failed: %v", err)
	}

	if !strings.Contains(buf2.String(), "<p>Custom: Framework</p>") {
		t.Errorf("rendered output mismatch: %s", buf2.String())
	}

	// 3. View not found
	var buf3 bytes.Buffer
	err = engine.Render(&buf3, "missing.html", nil)
	if err == nil {
		t.Errorf("expected error for missing view, got nil")
	}
}
