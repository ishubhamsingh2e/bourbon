package app

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/config"
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/router"
)

type MockProvider struct {
	mu         sync.Mutex
	registered bool
	booted     bool
	shutdown   bool
	bootOrder  *[]string
	name       string
}

func (m *MockProvider) Register(c core.Container) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registered = true
	return nil
}

func (m *MockProvider) Boot(ctx context.Context, a core.Application) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.booted = true
	if m.bootOrder != nil {
		*m.bootOrder = append(*m.bootOrder, m.name)
	}
	return nil
}

func (m *MockProvider) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdown = true
	if m.bootOrder != nil {
		*m.bootOrder = append(*m.bootOrder, "shutdown:"+m.name)
	}
	return nil
}

func TestAppLifecycleAndOptions(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Name: "Custom App"},
		HTTP: config.HTTPConfig{
			Host: "127.0.0.1",
			Port: 9876,
		},
		Env: config.EnvTesting,
	}

	bootOrder := []string{}
	p1 := &MockProvider{name: "p1", bootOrder: &bootOrder}
	p2 := &MockProvider{name: "p2", bootOrder: &bootOrder}

	customRouter := router.New()
	customRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	app := New(
		WithConfig(cfg),
		WithProviders(p1, p2),
		WithRouter(customRouter),
	)

	if app.Config().App.Name != "Custom App" {
		t.Errorf("expected App.Name 'Custom App', got %v", app.Config().App.Name)
	}
	if app.Router() != customRouter {
		t.Errorf("router was not set properly")
	}

	// Verify container bindings
	var resolvedConfig *config.Config
	if err := app.Container().Resolve(&resolvedConfig); err != nil || resolvedConfig != cfg {
		t.Errorf("container failed resolving App config singleton")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Run in background and cancel after brief execution
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-errChan
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}

	if !p1.registered || !p1.booted || !p1.shutdown {
		t.Errorf("p1 lifecycle incomplete: reg=%v, boot=%v, shut=%v", p1.registered, p1.booted, p1.shutdown)
	}
	if !p2.registered || !p2.booted || !p2.shutdown {
		t.Errorf("p2 lifecycle incomplete: reg=%v, boot=%v, shut=%v", p2.registered, p2.booted, p2.shutdown)
	}

	expectedOrder := []string{"p1", "p2", "shutdown:p2", "shutdown:p1"}
	if len(bootOrder) != len(expectedOrder) {
		t.Fatalf("unexpected lifecycle order: got %v, want %v", bootOrder, expectedOrder)
	}
	for i := range bootOrder {
		if bootOrder[i] != expectedOrder[i] {
			t.Errorf("bootOrder[%d] = %v, want %v", i, bootOrder[i], expectedOrder[i])
		}
	}
}

func TestAppUseAndLogger(t *testing.T) {
	application := New()

	if application.Logger() == nil {
		t.Fatal("expected default logger to be non-nil")
	}

	p := &MockProvider{name: "custom"}
	ret := application.Use(p)
	if ret != application {
		t.Fatal("expected Use() to return Application instance")
	}

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	go func() {
		errChan <- application.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	if err := <-errChan; err != nil {
		t.Fatalf("application.Run failed: %v", err)
	}

	if !p.registered || !p.booted || !p.shutdown {
		t.Errorf("provider lifecycle incomplete: reg=%v, boot=%v, shut=%v", p.registered, p.booted, p.shutdown)
	}
}
