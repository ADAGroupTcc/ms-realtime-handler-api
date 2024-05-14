package postgresql

import (
	"context"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/user"
	pperr "github.com/PicPay/lib-go-pperr"
)

// Store represents a postgresql store
type Store struct{}

// New create a new store
func New() *Store {
	return &Store{}
}

// Get return an user
func (s *Store) Get(ctx context.Context, id int) (*user.User, error) {
	if id != 1 {
		return nil, pperr.New("not found", pperr.ENOTFOUND)
	}
	return &user.User{
		ID:       1,
		Name:     "Fake User",
		Email:    "fake@email.com",
		Password: "ssdsdsdsw",
	}, nil
}

// Create insert an user
func (s *Store) Create(ctx context.Context, u *user.User) (int, error) {
	return 1, nil
}
