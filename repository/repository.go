// Package repository defines standard generic repository abstractions for domain data access.
package repository

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/database"
)

// Repository defines standard CRUD and query contracts for domain entities.
type Repository[T any, ID any] interface {
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
	WithTx(tx database.Tx) Repository[T, ID]
}
