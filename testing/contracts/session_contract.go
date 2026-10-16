package contracts

import (
	"context"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/session"
)

// VerifySessionStoreContract executes a standard compliance test suite against any session.Store implementation.
func VerifySessionStoreContract(t *testing.T, store session.Store) {
	ctx := context.Background()

	t.Run("Save_Load_Destroy", func(t *testing.T) {
		sess := session.New("contract_sess_id")
		sess.Set("user_id", 999)
		sess.Flash("notification", "Account verified")

		if err := store.Save(ctx, sess, 10*time.Minute); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := store.Load(ctx, "contract_sess_id")
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		uid, ok := loaded.GetInt("user_id")
		if !ok || uid != 999 {
			t.Errorf("loaded user_id = %v, %v, want 999, true", uid, ok)
		}

		notif, ok := loaded.Flashed("notification")
		if !ok || notif != "Account verified" {
			t.Errorf("loaded flash = %v, %v, want Account verified, true", notif, ok)
		}

		if err := store.Destroy(ctx, "contract_sess_id"); err != nil {
			t.Fatalf("Destroy failed: %v", err)
		}

		reloaded, err := store.Load(ctx, "contract_sess_id")
		if err != nil {
			t.Fatalf("Load after Destroy failed: %v", err)
		}
		if _, exists := reloaded.Get("user_id"); exists {
			t.Errorf("destroyed session still had values")
		}
	})
}
