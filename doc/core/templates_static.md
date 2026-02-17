# Templates and Static Files

Bourbon makes it easy to serve HTML templates and static assets like CSS, JavaScript, and images.

## Templates

Bourbon uses Go's `html/template` package with some enhancements, including auto-reloading during development.

### Configuration

Templates are configured in `settings.toml`:

```toml
[templates]
directory = "templates"
extension = ".html"
auto_reload = true
```

- `directory`: The root directory for templates.
- `extension`: The file extension to look for (e.g., `.html`, `.tmpl`).
- `auto_reload`: If true, templates are reloaded on every request (useful for development).

### Rendering Templates

In your controller or handler, use `c.Render()`:

```go
func Index(c *http.Context) error {
    data := map[string]interface{}{
        "Title": "Home Page",
        "User":  currentUser,
    }
    return c.Render("index.html", data)
}
```

### Template Inheritance

Bourbon supports template inheritance using `{{define "name"}}` and `{{template "name" .}}`.

**layout.html:**

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <nav>...</nav>
    <main>
        {{template "content" .}}
    </main>
    <footer>...</footer>
</body>
</html>
```

**index.html:**

```html
{{define "content"}}
    <h1>Welcome {{.User.Name}}</h1>
    <p>This is the home page.</p>
{{end}}
```

To render this, you would render `layout.html` but need to ensure `index.html` is parsed. Bourbon handles this by loading all templates in the directory. You might need to structure your templates correctly or use partials.

### Custom Functions

You can add custom functions to your templates in `main.go`:

```go
app.AddTemplateFunc("formatDate", func(t time.Time) string {
    return t.Format("2006-01-02")
})
```

Usage in template:

```html
<p>Published on: {{formatDate .CreatedAt}}</p>
```

## Static Files

Static files are served directly by the application.

### Configuration

```toml
[static]
directory = "static"
url_prefix = "/static"
```

- `directory`: The local directory containing static files.
- `url_prefix`: The URL path prefix to serve files from.

### Usage

Place your files in the `static/` directory:

```
static/
├── css/
│   └── style.css
├── js/
│   └── app.js
└── images/
    └── logo.png
```

Access them in your HTML:

```html
<link rel="stylesheet" href="/static/css/style.css">
<script src="/static/js/app.js"></script>
<img src="/static/images/logo.png" alt="Logo">
```

### Serving Multiple Directories

You can serve multiple static directories programmatically:

```go
app.Static("/assets", "./assets")
app.Static("/public", "./public")
```
