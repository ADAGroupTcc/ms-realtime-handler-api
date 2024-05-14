package user

import "context"

// User represents an user
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

// Reader user reader
type Reader interface {
	Get(ctx context.Context, id int) (*User, error)
}

// Writer user writer
type Writer interface {
	Create(ctx context.Context, u *User) (int, error)
}

// Repository interface
type Repository interface {
	Reader
	Writer
}

// UseCase interface
type UseCase interface {
	Get(ctx context.Context, id int) (*User, error)
	Create(ctx context.Context, name, email, password string) (*User, error)
}
