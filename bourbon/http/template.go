package http

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type TemplateEngine struct {
	templates  *template.Template
	directory  string
	extension  string
	autoReload bool
	funcs      template.FuncMap
	mu         sync.RWMutex
}

func NewTemplateEngine(directory, extension string, autoReload bool) *TemplateEngine {
	engine := &TemplateEngine{
		directory:  directory,
		extension:  extension,
		autoReload: autoReload,
		funcs:      template.FuncMap{},
	}
	return engine
}

func (e *TemplateEngine) AddFunc(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
}

func (e *TemplateEngine) AddFuncs(funcs template.FuncMap) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for name, fn := range funcs {
		e.funcs[name] = fn
	}
}

func (e *TemplateEngine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, err := os.Stat(e.directory); os.IsNotExist(err) {
		return fmt.Errorf("template directory does not exist: %s", e.directory)
	}

	tmpl := template.New("").Funcs(e.funcs)

	err := filepath.WalkDir(e.directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == e.extension {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read template %s: %w", path, err)
			}

			relPath, err := filepath.Rel(e.directory, path)
			if err != nil {
				return err
			}

			name := filepath.ToSlash(relPath)

			_, err = tmpl.New(name).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", name, err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	e.templates = tmpl
	return nil
}

func (e *TemplateEngine) Render(name string, data interface{}) (string, error) {
	if e.autoReload {
		if err := e.Load(); err != nil {
			return "", err
		}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.templates == nil {
		return "", fmt.Errorf("templates not loaded, call Load() first")
	}

	tmpl := e.templates.Lookup(name)
	if tmpl == nil {
		return "", fmt.Errorf("template not found: %s", name)
	}

	var buf []byte
	writer := &bufferWriter{buf: &buf}

	if err := tmpl.Execute(writer, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return string(buf), nil
}

type bufferWriter struct {
	buf *[]byte
}

func (w *bufferWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
