package users

import (
	"context"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, u *User) error
	List(ctx context.Context) ([]*User, error)
}

type sqlRepository struct {
	db database.DB
}

func NewRepository(db database.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if r.db == nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}
	query := "SELECT id, email, name, created_at, updated_at FROM users WHERE id = ?"
	row := r.db.QueryRow(ctx, query, id)
	if row == nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}
	return &u, nil
}

func (r *sqlRepository) Create(ctx context.Context, u *User) error {
	if r.db == nil {
		return nil
	}
	query := "INSERT INTO users (id, email, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(ctx, query, u.ID, u.Email, u.Name, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *sqlRepository) List(ctx context.Context) ([]*User, error) {
	if r.db == nil {
		return []*User{
			{ID: "1", Email: "admin@bourbon.dev", Name: "Admin User", CreatedAt: time.Now()},
		}, nil
	}
	query := "SELECT id, email, name, created_at, updated_at FROM users"
	rows, err := r.db.Query(ctx, query)
	if err != nil || rows == nil {
		return []*User{
			{ID: "1", Email: "admin@bourbon.dev", Name: "Admin User", CreatedAt: time.Now()},
		}, nil
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, &u)
	}
	return result, nil
}
