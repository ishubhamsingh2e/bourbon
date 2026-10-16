package memstore

import (
	"context"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/session"
)

func TestMemStorePersistence(t *testing.T) {
	ctx := context.Background()
	store := New(0)
	defer store.Close()

	sess := session.New("sess_abc")
	sess.Set("theme", "dark")
	sess.Flash("info", "Welcome!")

	if err := store.Save(ctx, sess, 10*time.Minute); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load(ctx, "sess_abc")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	theme, ok := loaded.Get("theme")
	if !ok || theme != "dark" {
		t.Errorf("loaded theme = %v, %v, want dark, true", theme, ok)
	}

	flash, ok := loaded.Flashed("info")
	if !ok || flash != "Welcome!" {
		t.Errorf("loaded flash info = %v, %v, want Welcome!, true", flash, ok)
	}

	// Destroy
	sess.Destroy()
	if err := store.Save(ctx, sess, 10*time.Minute); err != nil {
		t.Fatalf("Save destroyed session failed: %v", err)
	}

	reloaded, _ := store.Load(ctx, "sess_abc")
	if _, exists := reloaded.Get("theme"); exists {
		t.Errorf("destroyed session still had values in store")
	}
}

func TestMemStoreTTLExpiration(t *testing.T) {
	ctx := context.Background()
	store := New(0)
	defer store.Close()

	sess := session.New("short_lived")
	sess.Set("foo", "bar")
	_ = store.Save(ctx, sess, 20*time.Millisecond)

	time.Sleep(50 * time.Millisecond)

	loaded, _ := store.Load(ctx, "short_lived")
	if _, exists := loaded.Get("foo"); exists {
		t.Errorf("expired session should not load values")
	}
}

func TestMemStoreProviderLifecycle(t *testing.T) {
	provider := NewProvider(time.Minute)
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering memstore provider: %v", err)
	}

	var store session.Store
	if err := application.Container().Resolve(&store); err != nil {
		t.Fatalf("failed resolving session.Store: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting memstore provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down memstore provider: %v", err)
	}
}
