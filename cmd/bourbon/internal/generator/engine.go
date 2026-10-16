// Package generator provides template execution and gofmt formatting for code generation.
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Engine executes text templates and formats Go source code.
type Engine struct {
	funcMap template.FuncMap
}

// New constructs a code generation Engine with helper functions.
func New() *Engine {
	return &Engine{
		funcMap: template.FuncMap{
			"lower":  strings.ToLower,
			"upper":  strings.ToUpper,
			"title":  strings.Title,
			"kebab":  toKebabCase,
			"snake":  toSnakeCase,
			"camel":  toCamelCase,
			"pascal": toPascalCase,
			"plural": pluralize,
		},
	}
}

// Render executes a template string with data.
func (e *Engine) Render(tmplStr string, data any) ([]byte, error) {
	tmpl, err := template.New("gen").Funcs(e.funcMap).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("generator: failed parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("generator: failed executing template: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderString applies a named string transform (e.g. "snake", "pascal", "plural").
func (e *Engine) RenderString(transform, s string) string {
	switch transform {
	case "snake":
		return toSnakeCase(s)
	case "pascal":
		return toPascalCase(s)
	case "camel":
		return toCamelCase(s)
	case "kebab":
		return toKebabCase(s)
	case "plural":
		return pluralize(s)
	default:
		return s
	}
}

// RenderGo executes a template and formats the result with standard gofmt.
func (e *Engine) RenderGo(tmplStr string, data any) ([]byte, error) {
	raw, err := e.Render(tmplStr, data)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(raw)
	if err != nil {
		return nil, fmt.Errorf("generator: gofmt failed on generated output:\n%s\nerror: %w", string(raw), err)
	}

	return formatted, nil
}

// WriteFile writes content to destination, creating parent directories as necessary.
func (e *Engine) WriteFile(destPath string, content []byte, overwrite bool) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("generator: failed creating directory '%s': %w", dir, err)
	}

	if !overwrite {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("generator: file '%s' already exists", destPath)
		}
	}

	return os.WriteFile(destPath, content, 0644)
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(strings.ToLower(toSnakeCase(s)), "_", "-")
}

func toSnakeCase(s string) string {
	var res strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res.WriteRune('_')
		}
		res.WriteRune(r)
	}
	return strings.ToLower(res.String())
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") && !strings.HasSuffix(s, "ay") && !strings.HasSuffix(s, "ey") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}
