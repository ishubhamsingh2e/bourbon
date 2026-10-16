package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/router"
)

type sampleBody struct {
	Name string `json:"name" form:"name"`
	Age  int    `json:"age" form:"age"`
}

func (s *sampleBody) Validate() error {
	return nil
}

func TestContextJSONAndString(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	c := core.NewContext(w, r)

	err := c.JSON(http.StatusOK, map[string]string{"msg": "pong"})
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"msg":"pong"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestContextBinding(t *testing.T) {
	payload := `{"name":"Bourbon","age":2}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(payload))
	r.Header.Set("Content-Type", "application/json")
	c := core.NewContext(w, r)

	var body sampleBody
	if err := c.BindJSON(&body); err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}
	if body.Name != "Bourbon" || body.Age != 2 {
		t.Fatalf("unexpected bound data: %+v", body)
	}
}

func TestNativeRouterHandlerFunc(t *testing.T) {
	r := router.New()

	// Direct registration of func(c *core.Context) error without Wrap()
	r.Get("/greet/{name}", func(c *core.Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, map[string]string{
			"greeting": "hello " + name,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/greet/alice", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if res["greeting"] != "hello alice" {
		t.Fatalf("expected 'hello alice', got %q", res["greeting"])
	}
}

func TestContextStoreAndSubsystems(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := core.NewContext(w, r)

	c.Set("foo", "bar")
	val, ok := c.Get("foo")
	if !ok || val != "bar" {
		t.Fatalf("expected 'bar', got %v", val)
	}

	if c.DB() != nil {
		t.Fatal("expected nil DB")
	}
	if c.Cache() != nil {
		t.Fatal("expected nil Cache")
	}
	if c.Session() != nil {
		t.Fatal("expected nil Session")
	}
	if c.Auth() != nil {
		t.Fatal("expected nil Auth")
	}
	if _, err := c.RequireAuth(); err == nil {
		t.Fatal("expected error for unauthenticated RequireAuth")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	c := core.NewContext(w, r)

	select {
	case <-c.Ctx.Done():
	default:
		t.Fatal("expected context to be cancelled")
	}
}
