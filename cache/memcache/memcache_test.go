package memcache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/cache"
)

type Profile struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

func TestMemCacheBasicGetSet(t *testing.T) {
	ctx := context.Background()
	mc := New(0) // disable periodic cleaner for synchronous test
	defer mc.Close()

	prof := Profile{Username: "dev_user", Role: "admin"}
	if err := mc.Set(ctx, "user:1", prof, 10*time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	has, err := mc.Has(ctx, "user:1")
	if err != nil || !has {
		t.Fatalf("Has failed: %v, has=%v", err, has)
	}

	var fetched Profile
	if err := mc.Get(ctx, "user:1", &fetched); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if fetched.Username != "dev_user" || fetched.Role != "admin" {
		t.Errorf("fetched profile incorrect: %+v", fetched)
	}

	if err := mc.Delete(ctx, "user:1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	err = mc.Get(ctx, "user:1", &fetched)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got: %v", err)
	}
}

func TestMemCacheTTLExpiration(t *testing.T) {
	ctx := context.Background()
	mc := New(0)
	defer mc.Close()

	if err := mc.Set(ctx, "temp_key", "temporary_val", 20*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	var str string
	err := mc.Get(ctx, "temp_key", &str)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss on expired item, got: %v", err)
	}
}

func TestMemCacheRemember(t *testing.T) {
	ctx := context.Background()
	mc := New(0)
	defer mc.Close()

	callCount := 0
	compute := func() (any, error) {
		callCount++
		return "computed_data", nil
	}

	var res1, res2 string
	if err := mc.Remember(ctx, "comp_key", time.Minute, &res1, compute); err != nil {
		t.Fatalf("Remember 1 failed: %v", err)
	}
	if err := mc.Remember(ctx, "comp_key", time.Minute, &res2, compute); err != nil {
		t.Fatalf("Remember 2 failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected compute to be called exactly once, called %d times", callCount)
	}
	if res1 != "computed_data" || res2 != "computed_data" {
		t.Errorf("results mismatch: res1=%v, res2=%v", res1, res2)
	}
}

func TestMemCacheAtomicIncrement(t *testing.T) {
	ctx := context.Background()
	mc := New(0)
	defer mc.Close()

	v1, err := mc.Increment(ctx, "counter", 5)
	if err != nil || v1 != 5 {
		t.Fatalf("Increment 1 failed: %v, v1=%d", err, v1)
	}

	v2, err := mc.Increment(ctx, "counter", 10)
	if err != nil || v2 != 15 {
		t.Fatalf("Increment 2 failed: %v, v2=%d", err, v2)
	}

	v3, err := mc.Decrement(ctx, "counter", 3)
	if err != nil || v3 != 12 {
		t.Fatalf("Decrement failed: %v, v3=%d", err, v3)
	}
}

func TestMemCacheConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	mc := New(0)
	defer mc.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "concurrent_key"
			_, _ = mc.Increment(ctx, key, 1)
			var val int64
			_ = mc.Get(ctx, key, &val)
		}(i)
	}
	wg.Wait()
}

func TestMemCacheProviderLifecycle(t *testing.T) {
	provider := NewProvider(Config{CleanupInterval: time.Minute})
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering memcache provider: %v", err)
	}

	var ca cache.Cache
	if err := application.Container().Resolve(&ca); err != nil {
		t.Fatalf("failed resolving cache.Cache: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting memcache provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down memcache provider: %v", err)
	}
}
