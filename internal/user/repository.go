package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
INSERT INTO users(email, name, password)
VALUES($1, $2, $3)
RETURNING id, created_at
`

	err := r.db.QueryRow(ctx,
		query,
		user.Email,
		user.Name,
		user.Password,
	).Scan(&user.Id, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, password, created_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRow(ctx, query, email)

	var u User
	err := row.Scan(&u.Id, &u.Email, &u.Name, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &u, nil
}
