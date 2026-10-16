package contracts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/cache"
)

// VerifyCacheContract executes a standard compliance test suite against any cache.Cache implementation.
func VerifyCacheContract(t *testing.T, c cache.Cache) {
	ctx := context.Background()

	t.Run("Set_Get_Delete", func(t *testing.T) {
		key := "contract_cache_key"
		val := "contract_cache_val"

		if err := c.Set(ctx, key, val, 10*time.Minute); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		has, err := c.Has(ctx, key)
		if err != nil || !has {
			t.Fatalf("Has failed: %v, has=%v", err, has)
		}

		var retrieved string
		if err := c.Get(ctx, key, &retrieved); err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if retrieved != val {
			t.Errorf("retrieved %v, want %v", retrieved, val)
		}

		if err := c.Delete(ctx, key); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		err = c.Get(ctx, key, &retrieved)
		if !errors.Is(err, cache.ErrCacheMiss) {
			t.Errorf("expected ErrCacheMiss, got: %v", err)
		}
	})

	t.Run("Remember", func(t *testing.T) {
		key := "contract_remember_key"
		callCount := 0
		cb := func() (any, error) {
			callCount++
			return "computed_data", nil
		}

		var res1, res2 string
		_ = c.Remember(ctx, key, time.Minute, &res1, cb)
		_ = c.Remember(ctx, key, time.Minute, &res2, cb)

		if callCount != 1 {
			t.Errorf("expected closure to execute once, executed %d times", callCount)
		}
		if res1 != "computed_data" || res2 != "computed_data" {
			t.Errorf("unexpected results: res1=%v, res2=%v", res1, res2)
		}
	})

	t.Run("Increment_Decrement", func(t *testing.T) {
		key := "contract_counter"
		_ = c.Delete(ctx, key)

		v1, err := c.Increment(ctx, key, 10)
		if err != nil || v1 != 10 {
			t.Fatalf("Increment failed: %v, v1=%d", err, v1)
		}

		v2, err := c.Decrement(ctx, key, 4)
		if err != nil || v2 != 6 {
			t.Fatalf("Decrement failed: %v, v2=%d", err, v2)
		}
	})

	t.Run("Flush", func(t *testing.T) {
		_ = c.Set(ctx, "k1", "v1", time.Minute)
		_ = c.Set(ctx, "k2", "v2", time.Minute)

		if err := c.Flush(ctx); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		has, _ := c.Has(ctx, "k1")
		if has {
			t.Errorf("key k1 should have been flushed")
		}
	})
}
