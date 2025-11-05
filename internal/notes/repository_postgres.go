package notes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)


// postgresNotesRepository is a private struct that implements the NotesRepository interface.
// It holds a database connection pool and provides methods that operate on it.
type postgresNotesRepository struct {
	db *pgxpool.Pool
}

// NewPostgresNotesRepository is a constructor that returns a NotesRepository implementation
// backed by PostgreSQL. This is where dependency injection happens — we inject the pgxpool
// dependency into our repository.
//
// Because this function returns a NotesRepository interface (not a concrete struct),
// the returned object must implement all methods defined in that interface.
//
// This pattern hides the underlying implementation (postgresNotesRepository)
// and safely exposes only the methods we choose — Go’s way of achieving encapsulation.
func NewPostgresNotesRepository(db *pgxpool.Pool) NotesRepository {
	return &postgresNotesRepository{db: db}
}


func (r *postgresNotesRepository) CreateNote(ctx context.Context, n *Note) (*Note, error) {
	query := `
	INSERT INTO notes(author_id, title, content, public, slug)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at
	`

	// Scan returned fields back into struct
	err := r.db.QueryRow(ctx, query,
		n.AuthorID,
		n.Title,
		n.Content,
		n.Public,
		n.Slug,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		fmt.Printf("error while creating note: %v", err)
		return nil, fmt.Errorf("error while creating note: %w", err)
	}

	return n, nil
}

func (r *postgresNotesRepository) GetNotesByAuthor(ctx context.Context, authorID string) ([]*NoteSummary, error) {
	query := `
	SELECT id, title, author_id, public, slug, created_at
	FROM notes
	WHERE author_id = $1
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []*NoteSummary

	for rows.Next() {
		var n NoteSummary
		err := rows.Scan(
			&n.ID,
			&n.Title,
			&n.AuthorID,
			&n.Public,
			&n.Slug,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}
		notes = append(notes, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return notes, nil
}
