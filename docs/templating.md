# Server-Rendered HTML Templating in Bourbon

Bourbon provides high-speed, standard-library-backed HTML templating using Go's native `html/template` package, extended with layout composition, partials, custom helper functions, and dev-mode hot reloading without server restart.

---

## 1. Setting Up the Template Engine

Bourbon's `template.Engine` implements the `view.Renderer` interface and can load templates from any standard `fs.FS` or `embed.FS`:

```go
package main

import (
	"embed"
	"net/http"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/router"
	"github.com/ishubhamsingh2e/bourbon/template"
)

//go:embed templates/*
var templateFS embed.FS

func SetupTemplates(r router.Router) {
	// Initialize engine (fsys, pattern, devMode)
	engine := template.New(templateFS, "templates/*.html", true)

	// Register custom helper functions
	engine.AddFunc("titleCase", func(s string) string {
		return strings.Title(s)
	})

	r.Get("/welcome", func(c *core.Context) error {
		c.SetViewRenderer(engine)
		return c.Render(http.StatusOK, "welcome.html", map[string]any{
			"Title": "Welcome to Bourbon",
			"User":  c.Auth(),
		})
	})
}
```

---

## 2. Layouts and Partials Composition

Organize your templates with reusable base layouts and partials:

```html
<!-- templates/layout.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/css/app.css">
</head>
<body>
    {{template "nav" .}}
    <main class="container">
        {{template "content" .}}
    </main>
    {{template "footer" .}}
</body>
</html>
```

```html
<!-- templates/welcome.html -->
{{define "content"}}
<h1>{{.Title}}</h1>
<p>Hello from Bourbon's native template engine!</p>
{{end}}
```

---

## 3. Built-in Template Helpers

Bourbon provides safe string and formatting functions pre-registered in every template engine:

- `safeHTML(s)`: Prevents escaping for trusted raw HTML strings.
- `safeJS(s)`: Emits trusted inline JavaScript.
- `safeCSS(s)`: Emits trusted inline CSS.
- `lower(s)`: Converts strings to lowercase.
- `upper(s)`: Converts strings to uppercase.

---

## 4. Development-Mode Hot Reloading

When `devMode` is set to `true`, the `template.Engine` checks for template file modifications and recompiles templates on every incoming request, eliminating the need to restart the Go process during UI iterations.
