package book

import "context"

// Book represents a book
type Book struct {
	ID       int
	Title    string
	Author   string
	Pages    int
	Quantity int
}

// Reader book reader
type Reader interface {
	Get(ctx context.Context, id int) (*Book, error)
}

// Writer book writer
type Writer interface {
	Create(ctx context.Context, e *Book) (int, error)
}

// Repository interface
type Repository interface {
	Reader
	Writer
}

// UseCase interface
type UseCase interface {
	Get(ctx context.Context, id int) (*Book, error)
	Create(ctx context.Context, title, author string, pages, quantity int) (*Book, error)
}
