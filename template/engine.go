// Package template provides a server-rendered html/template engine with layouts and partials.
package template

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"sync"
)

// Engine implements view.Renderer using standard library html/template.
type Engine struct {
	mu       sync.RWMutex
	fsys     fs.FS
	pattern  string
	funcMap  template.FuncMap
	devMode  bool
	compiled *template.Template
}

// New constructs an Engine with default helper functions and compiles templates from the provided filesystem.
func New(fsys fs.FS, pattern string, devMode bool) *Engine {
	if pattern == "" {
		pattern = "*.html"
	}

	e := &Engine{
		fsys:     fsys,
		pattern:  pattern,
		devMode:  devMode,
		funcMap:  make(template.FuncMap),
		compiled: template.New("root"),
	}

	// Register standard template helper functions
	e.funcMap["safeHTML"] = func(s string) template.HTML { return template.HTML(s) }
	e.funcMap["safeJS"] = func(s string) template.JS { return template.JS(s) }
	e.funcMap["safeCSS"] = func(s string) template.CSS { return template.CSS(s) }
	e.funcMap["lower"] = strings.ToLower
	e.funcMap["upper"] = strings.ToUpper

	_ = e.Reload()
	return e
}

// AddFunc registers a custom template function.
func (e *Engine) AddFunc(name string, fn any) *Engine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcMap[name] = fn
	_ = e.reloadInternal()
	return e
}

// AddFuncs registers multiple custom template functions.
func (e *Engine) AddFuncs(funcMap template.FuncMap) *Engine {
	e.mu.Lock()
	defer e.mu.Unlock()
	for k, v := range funcMap {
		e.funcMap[k] = v
	}
	_ = e.reloadInternal()
	return e
}

// Reload recompiles all templates from the filesystem.
func (e *Engine) Reload() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.reloadInternal()
}

func (e *Engine) reloadInternal() error {
	if e.fsys == nil {
		return nil
	}

	tmpl := template.New("bourbon_root").Funcs(e.funcMap)
	parsed, err := tmpl.ParseFS(e.fsys, e.pattern)
	if err != nil {
		return fmt.Errorf("template: failed parsing template pattern '%s': %w", e.pattern, err)
	}

	e.compiled = parsed
	return nil
}

// Render executes the named template with data into writer w.
func (e *Engine) Render(w io.Writer, name string, data any) error {
	if e.devMode {
		_ = e.Reload()
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.compiled == nil {
		return fmt.Errorf("template: no templates compiled")
	}

	tmpl := e.compiled.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("template: view '%s' not found", name)
	}

	return tmpl.Execute(w, data)
}
