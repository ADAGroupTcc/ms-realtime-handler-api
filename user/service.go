package user

import (
	"context"
	"fmt"
)

// Service user usecase
type Service struct {
	repo Repository
}

// NewService create new service
func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

// Create create an user
func (s *Service) Create(ctx context.Context, name, email, password string) (*User, error) {
	u := &User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	id, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("Error creating user %w", err)
	}
	u.ID = id
	return u, nil
}

// Get get an user
func (s *Service) Get(ctx context.Context, id int) (*User, error) {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Error reading user %w", err)
	}
	return b, nil
}
