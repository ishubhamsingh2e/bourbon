package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/contrib/sqlite"
	"github.com/ishubhamsingh2e/bourbon/database"
	_ "modernc.org/sqlite"
)

type TestUser struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	Email     string     `db:"email"`
	Age       int        `db:"age"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func setupTestDB(t *testing.T) database.DB {
	t.Helper()
	db, err := sqlite.Open(sqlite.Config{
		DSN: "file::memory:?cache=shared",
	})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}

	ctx := context.Background()
	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		age INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP NULL
	);`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestQueryAllAndQueryOne(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)

	// Insert test rows
	_, err := database.Exec(ctx, db, "INSERT INTO users (id, name, email, age, created_at) VALUES (?, ?, ?, ?, ?)",
		"u1", "Alice", "alice@example.com", 30, now,
	)
	if err != nil {
		t.Fatalf("failed to insert u1: %v", err)
	}
	_, err = database.Exec(ctx, db, "INSERT INTO users (id, name, email, age, created_at) VALUES (?, ?, ?, ?, ?)",
		"u2", "Bob", "bob@example.com", 25, now,
	)
	if err != nil {
		t.Fatalf("failed to insert u2: %v", err)
	}

	// QueryAll with positional parameters
	users, err := database.QueryAll[TestUser](ctx, db, "SELECT * FROM users WHERE age >= ? ORDER BY name ASC", 20)
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[1].Name != "Bob" {
		t.Errorf("unexpected user names: %+v", users)
	}

	// QueryOne found
	user, err := database.QueryOne[TestUser](ctx, db, "SELECT * FROM users WHERE id = ?", "u1")
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %s", user.Email)
	}

	// QueryOne not found -> ErrNotFound
	missingUser, err := database.QueryOne[TestUser](ctx, db, "SELECT * FROM users WHERE id = ?", "non-existent")
	if !errors.Is(err, database.ErrNotFound) {
		t.Errorf("expected database.ErrNotFound, got: %v (user: %+v)", err, missingUser)
	}
}

func TestRawQueryWithNamedParameters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)

	_, err := database.Exec(ctx, db, "INSERT INTO users (id, name, email, age, created_at) VALUES (?, ?, ?, ?, ?)",
		"u3", "Charlie", "charlie@example.com", 35, now,
	)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	var user TestUser
	err = database.Raw(db, "SELECT * FROM users WHERE email = :email AND age = :age", database.Params{
		"email": "charlie@example.com",
		"age":   35,
	}).ScanOne(ctx, &user)
	if err != nil {
		t.Fatalf("ScanOne with named params failed: %v", err)
	}
	if user.Name != "Charlie" {
		t.Errorf("expected Charlie, got %s", user.Name)
	}

	var users []TestUser
	err = database.Raw(db, "SELECT * FROM users WHERE age > :min_age", database.Params{
		"min_age": 30,
	}).ScanAll(ctx, &users)
	if err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Charlie" {
		t.Errorf("expected Charlie in users, got %+v", users)
	}
}
