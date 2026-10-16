# Database, Models & Migrations in Bourbon

Bourbon decouples data access through the `database.DB` port, supporting connection pooling, read-replica routing, unit-of-work transactions, and embedded database migrations.

---

## 1. Database Port & Connection

Bourbon applications connect to databases through official adapters or standard `*sql.DB` wrappers:

```go
package main

import (
	"context"
	"log"

	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
	"github.com/ishubhamsingh2e/bourbon/database"
)

func ConnectDB() database.DB {
	db, err := sqlite.Open(sqlite.Config{
		DSN:       "storage/app.db",
		EnableWAL: true,
	})
	if err != nil {
		log.Fatalf("failed connecting to SQLite: %v", err)
	}
	return db
}
```

Supported drivers:
- **SQLite**: `github.com/ishubhamsingh2e/bourbon/contrib/sqlite` (pure Go via `modernc.org/sqlite`)
- **PostgreSQL**: `github.com/ishubhamsingh2e/bourbon/contrib/postgres` (via `pgx/v5`)
- **MySQL**: `github.com/ishubhamsingh2e/bourbon/contrib/mysql` (via `go-sql-driver/mysql`)
- **Test Fake**: `github.com/ishubhamsingh2e/bourbon/database/memdb` (zero daemon, in-memory)

---

## 2. Unit-of-Work Transactions (`database.WithTx`)

Bourbon guarantees atomic transactions with automatic rollback on error or panic:

```go
package accounts

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/database"
)

func TransferFunds(ctx context.Context, db database.DB, fromID, toID string, amount int64) error {
	return database.WithTx(ctx, db, func(tx database.Tx) error {
		_, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
		if err != nil {
			return err // Triggers automated rollback
		}

		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
		return err // Non-nil rolls back; nil commits atomically
	})
}
```

---

## 3. Generic Repository Pattern

Bourbon defines the `repository.Repository[T, ID]` generic contract for entity persistence:

```go
package users

import (
	"context"
	"database/sql"

	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/model"
	"github.com/ishubhamsingh2e/bourbon/repository"
)

type User struct {
	model.BaseModel
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type UserRepository struct {
	db database.DB
}

func NewRepository(db database.DB) repository.Repository[User, string] {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRow(ctx, "SELECT id, email, full_name, created_at, updated_at FROM users WHERE id = ?", id)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]User, error) {
	// Query implementation...
	return nil, nil
}

func (r *UserRepository) Create(ctx context.Context, u *User) error {
	// Insert implementation...
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *User) error {
	// Update implementation...
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

func (r *UserRepository) WithTx(tx database.Tx) repository.Repository[User, string] {
	return &UserRepository{db: tx}
}
```

---

## 4. Database Migrations

Bourbon provides a zero-magic, transparent migration system. Migrations are executed through an explicit, readable Go runner in your project at `cmd/migrate/main.go`, which connects using your configured `config.Settings` and `cfg.DBDriver()`.

### Running Migrations via Bourbon CLI

```bash
# Apply all pending database migrations (in db/migrations/)
bourbon migrate
# or
bourbon migrate up

# Rollback the last applied migration batch
bourbon migrate rollback

# Inspect migration status table
bourbon migrate status

# Scaffold a new SQL migration
bourbon migrate create create_users_table
# or
bourbon make migration create_users_table
```

### Direct Go Toolchain Execution (Zero Magic)

Because `cmd/migrate/main.go` is standard Go code in your repository, you can execute, debug, or deploy migrations in production/CI without needing the Bourbon CLI binary installed:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate rollback
go run ./cmd/migrate status
```

### Migration File Format

Migration files are stored in `db/migrations/`:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    full_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
```
