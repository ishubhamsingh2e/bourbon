package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineRenderGoFormatting(t *testing.T) {
	eng := New()

	tmpl := `
package {{.Package}}

type {{.Name | pascal}} struct {
	ID string ` + "`json:\"id\"`" + `
}

func New{{.Name | pascal}}(id string) *{{.Name | pascal}} {
	return &{{.Name | pascal}}{ID: id}
}
`

	data := map[string]string{
		"Package": "user",
		"Name":    "user_account",
	}

	formatted, err := eng.RenderGo(tmpl, data)
	if err != nil {
		t.Fatalf("RenderGo failed: %v", err)
	}

	code := string(formatted)
	if !strings.Contains(code, "type UserAccount struct {") {
		t.Errorf("rendered code mismatch: %s", code)
	}

	// Test WriteFile
	tempDir, _ := os.MkdirTemp("", "bourbon_gen_*")
	defer os.RemoveAll(tempDir)

	target := filepath.Join(tempDir, "user", "account.go")
	err = eng.WriteFile(target, formatted, false)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Should error if exists and overwrite=false
	err = eng.WriteFile(target, formatted, false)
	if err == nil {
		t.Errorf("expected error when writing over existing file without overwrite")
	}
}
