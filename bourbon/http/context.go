package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type H map[string]interface{}

type Context struct {
	Writer          http.ResponseWriter
	Request         *http.Request
	Params          map[string]string
	store           map[string]interface{}
	TemplateEngine  *TemplateEngine
	asyncDispatcher AsyncDispatcher // For dispatching async jobs
}

// AsyncDispatcher is an interface for dispatching async jobs
type AsyncDispatcher interface {
	Dispatch(ctx context.Context, jobID, handler string, payload map[string]interface{}) error
	GetResult(ctx context.Context, jobID string) (interface{}, error)
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		store:   make(map[string]interface{}),
	}
}

func (c *Context) JSON(status int, data interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) String(status int, text string) error {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(text))
	return err
}

func (c *Context) HTML(status int, html string) error {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(html))
	return err
}

func (c *Context) Redirect(code int, url string) error {
	http.Redirect(c.Writer, c.Request, url, code)
	return nil
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Query(key string, defaultValue ...string) string {
	value := c.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) Body(v interface{}) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(v)
}

func (c *Context) Bind(v interface{}) error {
	return c.Body(v)
}

func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

func (c *Context) GetString(key string) string {
	if val, ok := c.store[key].(string); ok {
		return val
	}
	return ""
}

func (c *Context) Accepts(contentType string) bool {
	accept := c.Request.Header.Get("Accept")
	return strings.Contains(accept, contentType)
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Method() string {
	return c.Request.Method
}

func (c *Context) Path() string {
	return c.Request.URL.Path
}

func (c *Context) ClientIP() string {
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.Request.RemoteAddr
}

func (c *Context) Render(templateName string, data interface{}) error {
	if c.TemplateEngine == nil {
		return c.HTML(http.StatusInternalServerError, "Template engine not configured")
	}

	html, err := c.TemplateEngine.Render(templateName, data)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, html)
}

func (c *Context) RenderWithStatus(status int, templateName string, data interface{}) error {
	if c.TemplateEngine == nil {
		return c.HTML(http.StatusInternalServerError, "Template engine not configured")
	}

	html, err := c.TemplateEngine.Render(templateName, data)
	if err != nil {
		return err
	}

	return c.HTML(status, html)
}

func (c *Context) Validate(v interface{}) map[string]string {
	return nil
}

// DispatchAsync dispatches an async job and returns job ID
func (c *Context) DispatchAsync(handler string, payload map[string]interface{}) (string, error) {
	if c.asyncDispatcher == nil {
		return "", ErrAsyncNotConfigured
	}

	jobID := generateJobID()
	err := c.asyncDispatcher.Dispatch(c.Request.Context(), jobID, handler, payload)
	if err != nil {
		return "", err
	}

	return jobID, nil
}

// DispatchAsyncJSON dispatches an async job and returns JSON response
func (c *Context) DispatchAsyncJSON(status int, handler string, payload map[string]interface{}) error {
	jobID, err := c.DispatchAsync(handler, payload)
	if err != nil {
		return c.JSON(500, H{"error": err.Error()})
	}

	return c.JSON(status, H{
		"job_id":  jobID,
		"status":  "queued",
		"message": "Job dispatched successfully",
	})
}

// GetAsyncResult retrieves the result of an async job
func (c *Context) GetAsyncResult(jobID string) (interface{}, error) {
	if c.asyncDispatcher == nil {
		return nil, ErrAsyncNotConfigured
	}

	return c.asyncDispatcher.GetResult(c.Request.Context(), jobID)
}

// SetAsyncDispatcher sets the async dispatcher (called by middleware)
func (c *Context) SetAsyncDispatcher(dispatcher AsyncDispatcher) {
	c.asyncDispatcher = dispatcher
}

// Helper to generate unique job IDs
func generateJobID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// ErrAsyncNotConfigured is returned when async is not configured
var ErrAsyncNotConfigured = &AsyncError{Message: "async dispatcher not configured"}

type AsyncError struct {
	Message string
}

func (e *AsyncError) Error() string {
	return e.Message
}
