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

	slug := slugifyWithID(n.Title, n.ID)
	n.Slug = &slug

	_, err = r.db.Exec(ctx, `UPDATE notes SET slug=$1 WHERE id=$2`, n.Slug, n.ID)
	if err != nil {
		return nil, err
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

func (r *postgresNotesRepository) GetNoteByID(ctx context.Context, noteID, authorID string) (*Note, error) {
	query := `
		SELECT id, author_id, title, content, public, slug, created_at, updated_at
		FROM notes
		WHERE author_id = $1 AND id = $2
	`

	var n Note

	err := r.db.QueryRow(ctx, query, authorID, noteID).Scan(
		&n.ID,
		&n.AuthorID,
		&n.Title,
		&n.Content,
		&n.Public,
		&n.Slug,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not find the requested note: %w", err)
	}

	return &n, nil
}

func (r *postgresNotesRepository) DeleteNote(ctx context.Context, noteID, autourID string) error {
	query := `DELETE FROM notes
			WHERE id= $1 AND author_id = $2`

	cmdtag, err := r.db.Exec(ctx, query, noteID, autourID)

	if err != nil {
		return fmt.Errorf("error while deleting note %w", err)
	}

	if cmdtag.RowsAffected() == 0 {
		return fmt.Errorf("error could not found any note to delete")
	}

	return nil
}

func (r *postgresNotesRepository) UpdateNote(ctx context.Context, n *Note) (*NoteSummary, error) {
	query := `
		UPDATE notes
		SET title = $3,
		    content = $4,
		    public = $5,
		    slug = $6,
		    updated_at = NOW()
		WHERE author_id = $1 AND id = $2
		RETURNING id, title, slug, public, created_at, author_id;
	`

	newSlug := slugifyWithID(n.Title, n.ID)

	var summary NoteSummary
	err := r.db.QueryRow(ctx, query,
		n.AuthorID, // 🧠 now required to match logged-in user
		n.ID,
		n.Title,
		n.Content,
		n.Public,
		newSlug,
	).Scan(
		&summary.ID,
		&summary.Title,
		&summary.Slug,
		&summary.Public,
		&summary.CreatedAt,
		&summary.AuthorID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return &summary, nil
}

func (r *postgresNotesRepository) AddEmailShare(ctx context.Context, noteID, ownerId, emailId string) error {
	query := `
INSERT INTO note_shares(note_id, email)
SELECT n.id, $3
FROM notes n
WHERE n.id = $1 AND n.author_id = $2;
`

	cmdTag, err := r.db.Exec(ctx, query, noteID, ownerId, emailId)
	if err != nil {
		return fmt.Errorf("failed to share note: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("unauthorized or invalid note")
	}

	return nil
}

func (r *postgresNotesRepository) RemoveEmailShare(ctx context.Context, noteID, ownerID, emailID string) error {

	query := `
	DELETE FROM note_shares
	WHERE note_id = $1 AND email = $2
	  AND EXISTS(
	      SELECT 1 FROM notes 
	      WHERE id = $1 AND author_id = $3
	  )
	`

	cmdTag, err := r.db.Exec(ctx, query, noteID, emailID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to remove share: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("unauthorized or share not found")
	}

	return nil
}

func (r *postgresNotesRepository) GetNoteBySlug(
	ctx context.Context,
	slug string,
	userID *string,
	userEmail *string) (*Note, error) {

	var query string
	var args []interface{}

	if userID == nil {
		// -------------------------------
		// Public user: can only see public notes
		// -------------------------------
		query = `
            SELECT id, title, content, author_id, public, slug, created_at, updated_at
            FROM notes
            WHERE slug = $1
              AND public = TRUE
            LIMIT 1
        `
		args = []interface{}{slug}

	} else {
		// -------------------------------
		// Logged-in user:
		// owner OR shared OR public
		// -------------------------------
		query = `
            SELECT n.id, n.title, n.content, n.author_id, n.public, n.slug,
                   n.created_at, n.updated_at
            FROM notes n
            LEFT JOIN note_shares ns ON n.id = ns.note_id
            WHERE n.slug = $1
              AND (
                    n.public = TRUE OR
                    n.author_id = $2 OR
                    ns.email = $3
                  )
            LIMIT 1
        `

		args = []interface{}{slug, *userID, *userEmail}
	}

	var note Note
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.AuthorID,
		&note.Public,
		&note.Slug,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("note not found or access denied: %w", err)
	}

	return &note, nil
}
