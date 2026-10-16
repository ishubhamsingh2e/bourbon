package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JSON serializes the given data struct into JSON and writes it with status code.
func (c *Context) JSON(code int, data any) error {
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(code)
	return json.NewEncoder(c.Response).Encode(data)
}

// String formats a string and writes it with the given status code.
func (c *Context) String(code int, format string, values ...any) error {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(code)
	if len(values) > 0 {
		_, err := fmt.Fprintf(c.Response, format, values...)
		return err
	}
	_, err := io.WriteString(c.Response, format)
	return err
}

// Blob writes raw bytes with the specified content type.
func (c *Context) Blob(code int, contentType string, b []byte) error {
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.WriteHeader(code)
	_, err := c.Response.Write(b)
	return err
}

// Stream streams data from an io.Reader to the client.
func (c *Context) Stream(code int, contentType string, r io.Reader) error {
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.WriteHeader(code)
	_, err := io.Copy(c.Response, r)
	return err
}

// Redirect performs an HTTP redirect to the specified URL.
func (c *Context) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusPermanentRedirect {
		return fmt.Errorf("core: invalid redirect status code: %d", code)
	}
	http.Redirect(c.Response, c.Request, url, code)
	return nil
}

// NoContent writes the status code with no response body.
func (c *Context) NoContent(code int) error {
	c.Response.WriteHeader(code)
	return nil
}

// Render renders a named view template with data using the registered view.Renderer.
func (c *Context) Render(code int, name string, data any) error {
	renderer := c.ViewRenderer()
	if renderer == nil {
		return fmt.Errorf("core: no view renderer registered")
	}

	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(code)
	return renderer.Render(c.Response, name, data)
}
