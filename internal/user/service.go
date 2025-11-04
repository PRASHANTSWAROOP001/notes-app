package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail = errors.New("invalid email format provided")
	ErrWeakPassword = errors.New("password length is less than 8 chars")
	ErrEmailExists  = errors.New("email provided is already in use")
	ErrInvalidLogin = errors.New("wrong email/password combination provided")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type service struct {
	repo UserRepository
}

func NewService(r UserRepository) Service {
	return &service{repo: r}
}

func (s *service) Register(ctx context.Context, email, name, password string) (*User, error) {
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	existing, _ := s.repo.GetUserByEmail(ctx, email)

	if existing != nil {
		return nil, ErrEmailExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &User{
		Email:     email,
		Name:      name,
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*User, string, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || u == nil {
		return nil, "", ErrInvalidLogin
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, "", ErrInvalidLogin
	}

	u.Password = ""

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.Id,
		"email":   u.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := []byte(os.Getenv("JWT_SECRET"))

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	return u, tokenString, nil
}
