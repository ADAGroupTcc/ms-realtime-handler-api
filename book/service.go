package book

import (
	"context"
	"fmt"
)

// Service book usecase
type Service struct {
	repo Repository
}

// NewService create new service
func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

// Create create a book
func (s *Service) Create(ctx context.Context, title, author string, pages, quantity int) (*Book, error) {
	b := &Book{
		Title:    title,
		Author:   author,
		Pages:    pages,
		Quantity: quantity,
	}
	id, err := s.repo.Create(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("Error creating book %w", err)
	}
	b.ID = id
	return b, nil
}

// Get get a book
func (s *Service) Get(ctx context.Context, id int) (*Book, error) {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Error reading book %w", err)
	}
	return b, nil
}
