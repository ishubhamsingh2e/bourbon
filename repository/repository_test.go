package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database"
)

type User struct {
	ID   string
	Name string
}

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]User
	tx    database.Tx
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]User),
	}
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return &u, nil
}

func (r *InMemoryUserRepository) FindAll(ctx context.Context) ([]User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]User, 0, len(r.users))
	for _, u := range r.users {
		res = append(res, u)
	}
	return res, nil
}

func (r *InMemoryUserRepository) Create(ctx context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = *u
	return nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; !ok {
		return database.ErrNotFound
	}
	r.users[u.ID] = *u
	return nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return database.ErrNotFound
	}
	delete(r.users, id)
	return nil
}

func (r *InMemoryUserRepository) WithTx(tx database.Tx) Repository[User, string] {
	return &InMemoryUserRepository{
		users: r.users,
		tx:    tx,
	}
}

func TestGenericRepositoryContract(t *testing.T) {
	ctx := context.Background()
	var repo Repository[User, string] = NewInMemoryUserRepository()

	// 1. Create
	user := &User{ID: "usr_1", Name: "Alice"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 2. FindByID
	found, err := repo.FindByID(ctx, "usr_1")
	if err != nil || found.Name != "Alice" {
		t.Fatalf("FindByID failed: %v, found: %+v", err, found)
	}

	// 3. Update
	user.Name = "Alice In Wonderland"
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := repo.FindByID(ctx, "usr_1")
	if updated.Name != "Alice In Wonderland" {
		t.Errorf("updated name = %v, want Alice In Wonderland", updated.Name)
	}

	// 4. FindAll
	all, err := repo.FindAll(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("FindAll failed: %v, len: %d", err, len(all))
	}

	// 5. Delete
	if err := repo.Delete(ctx, "usr_1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindByID(ctx, "usr_1")
	if err == nil {
		t.Fatalf("expected ErrNotFound after deletion, got nil")
	}

	// 6. WithTx
	txRepo := repo.WithTx(nil)
	if fmt.Sprintf("%T", txRepo) != "*repository.InMemoryUserRepository" {
		t.Errorf("WithTx returned unexpected type: %T", txRepo)
	}
}
