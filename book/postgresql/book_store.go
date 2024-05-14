package postgresql

import (
	"context"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	pperr "github.com/PicPay/lib-go-pperr"
)

// Store represents a postgresql store
type Store struct{}

// New create a new store
func New() *Store {
	return &Store{}
}

// Get return a book
func (s *Store) Get(ctx context.Context, id int) (*book.Book, error) {
	if id != 1 {
		return nil, pperr.New("not found", pperr.ENOTFOUND)
	}
	return &book.Book{
		ID:       1,
		Title:    "Fake Book",
		Author:   "Fake Author",
		Pages:    100,
		Quantity: 50,
	}, nil
}

// Create insert a book
func (s *Store) Create(ctx context.Context, e *book.Book) (int, error) {
	return 1, nil
}
