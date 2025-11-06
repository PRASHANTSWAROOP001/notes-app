package notes

import (
	"context"
	"time"
)

type Note struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Public     bool      `json:"public"`
	Slug       *string   `json:"slug,omitempty"`
	SharedWith []string  `json:"shared_with,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// this is for recieving values for admin and how much notes it created in list.
type NoteSummary struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Title     string    `json:"title"`
	Public    bool      `json:"public"`
	Slug      *string   `json:"slug,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// this is to be used by repository like must be implemented function handling database.
type NotesRepository interface {
	CreateNote(ctx context.Context, n *Note) (*Note, error)
	//UpdateNote(ctx context.Context, n *Note) (*Note, error)
	DeleteNote(ctx context.Context, noteId, authorId string) error

	GetNoteByID(ctx context.Context, noteID, authorID string) (*Note, error)
	GetNotesByAuthor(ctx context.Context, authorID string) ([]*NoteSummary, error)

	//GetNoteBySlug(ctx context.Context, slug string) (*Note, error)
	//AddEmailShare(ctx context.Context, noteId, emailId string) error
	//RemoveEmailShare(ctx context.Context, noteId, emailId string) error

}

// this is to be implemented by services will be used via repos and handler.
type NotesService interface {
	CreateNote(ctx context.Context, note *Note) (*Note, error)
	//UpdateNote(ctx context.Context, note *Note, userID string) (*Note, error)
	DeleteNote(ctx context.Context, noteID, userID string) error
	GetUserNotes(ctx context.Context, userID string) ([]*NoteSummary, error)
	GetUserNote(ctx context.Context, noteID, userID string) (*Note, error)
	//GetPublicNote(ctx context.Context, slug string) (*Note, error)
	//ShareNoteViaEmail(ctx context.Context, noteID, ownerID, email string) error
}
