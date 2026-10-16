// Package testutil provides lightweight in-memory testing harnesses, clients, and assertion helpers for Bourbon apps.
package testutil

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
)

// TestApp wraps an *app.App instance with an in-memory httptest.Server.
type TestApp struct {
	App    *app.App
	Server *httptest.Server
	t      *testing.T
}

// NewTestApp creates an in-memory test application harness.
func NewTestApp(t *testing.T, opts ...app.Option) *TestApp {
	a := app.New(opts...)

	// Start httptest Server wrapping router directly
	server := httptest.NewServer(a.Router())

	ta := &TestApp{
		App:    a,
		Server: server,
		t:      t,
	}

	if t != nil {
		t.Cleanup(func() {
			ta.Close()
		})
	}

	return ta
}

// Client returns a fluent HTTP test client targeted at this test application.
func (ta *TestApp) Client() *TestClient {
	return NewClient(ta.t, ta.Server.URL)
}

// Close terminates the test server and executes application shutdown.
func (ta *TestApp) Close() {
	if ta.Server != nil {
		ta.Server.Close()
	}
	_ = ta.App.Shutdown(context.Background())
}
