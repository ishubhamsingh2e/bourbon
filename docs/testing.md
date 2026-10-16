# Testing & Contract Verification in Bourbon

Bourbon provides a rich, standard-library-aligned testing harness that eliminates boilerplate in HTTP route tests, enables fast unit tests with in-memory database fakes, and guarantees adapter compliance through contract test suites.

---

## 1. HTTP Route Testing with `testutil`

Bourbon's `testutil.TestClient` provides fluent methods for sending HTTP requests against `http.Handler` or running servers, returning chainable assertions via `TestResponse`:

```go
package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/router"
	"github.com/ishubhamsingh2e/bourbon/testing/testutil"
)

func TestHealthAndUsers(t *testing.T) {
	r := router.New()
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	})

	server := httptest.NewServer(r)
	defer server.Close()

	client := testutil.NewClient(t, server.URL)

	// Chainable assertions on status, headers, and JSON content
	client.Get("/healthz", nil).
		AssertStatus(http.StatusOK).
		AssertHeader("Content-Type", "application/json").
		AssertBodyContains("ok").
		AssertJSONPath("status", "ok").
		AssertJSONPath("version", "1.0.0")
}
```

### Supported Assertions

- `AssertStatus(code)`: Verifies HTTP response status code.
- `AssertHeader(key, val)`: Validates specific header value.
- `AssertHeaderNotEmpty(key)`: Verifies presence of a non-empty header (e.g. `X-Request-ID`).
- `AssertBodyContains(substr)`: Checks string substring in response body.
- `AssertJSONPath(path, expected)`: Inspects dot-delimited JSON attributes (e.g., `user.profile.name`).

---

## 2. In-Memory Database Fake (`database/memdb`)

For unit testing repositories and services without running external database daemons (such as SQLite or Postgres), use Bourbon's in-memory fake database `database/memdb`:

```go
package repository_test

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database/memdb"
)

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	db := memdb.New()

	// Stage seed data directly in the in-memory fake
	db.Insert("users", map[string]any{
		"id":    "user_123",
		"email": "shubham@example.com",
		"name":  "Shubham Singh",
	})

	repo := NewUserRepository(db)
	user, err := repo.FindByID(ctx, "user_123")
	if err != nil {
		t.Fatalf("unexpected error finding user: %v", err)
	}

	if user.Email != "shubham@example.com" {
		t.Errorf("expected email shubham@example.com, got %s", user.Email)
	}
}
```

---

## 3. Adapter Contract Test Suites (`testing/contracts`)

Bourbon ships with reusable contract test suites under `testing/contracts/`. If you build or test a custom driver or storage adapter, run Bourbon's compliance suite to guarantee it satisfies framework contracts:

| Contract Test | Interface Tested |
| :--- | :--- |
| `RunDatabaseContract(t, db)` | `database.DB` connection, query, transaction rollback/commit |
| `RunCacheContract(t, cache)` | `cache.Cache` Get, Set, Delete, Increment, TTL expiration |
| `RunSessionStoreContract(t, store)` | `session.Store` Save, Load, Destroy, Touch |
| `RunStorageContract(t, storage)` | `storage.Storage` Put, Get, Delete, Exists, URL generation |
| `RunMailerContract(t, mailer)` | `mail.Mailer` Send, queue dispatch |
| `RunEventBusContract(t, bus)` | `events.Bus` Sync/async publication and subscription |

### Example Contract Verification

```go
package custom_redis_test

import (
	"testing"

	"github.com/ishubhamsingh2e/bourbon/testing/contracts"
	"my-app/custom_redis"
)

func TestCustomRedisCacheCompliance(t *testing.T) {
	client := custom_redis.NewClient("localhost:6379")
	contracts.RunCacheContract(t, client)
}
```
