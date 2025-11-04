package user

import (
	"context"
	"time"
)

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type Service interface {
	Register(ctx context.Context, email, name, password string) (*User, error)
	Login(ctx context.Context, email string, password string) (*User, string, error)
}
