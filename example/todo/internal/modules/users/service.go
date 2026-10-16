package users

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, u *User) error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.ID == "" {
		u.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return s.repo.Create(ctx, u)
}
