package notes

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// service is a private struct that implements the NotesService interface.
// It represents the business logic layer for notes and depends on a NotesRepository
// to perform data persistence operations (CreateNote, GetNotesByAuthor, etc).
//
// In short: service = business logic + repository access.
type service struct {
	repo NotesRepository
}

// NewNotesService is a public constructor function that returns a NotesService implementation.
// This is where dependency injection happens — we "inject" a NotesRepository (usually backed by a DB)
// into the service layer, allowing the service to call repository methods without knowing *how*
// or *where* the data is stored.
//
// So the chain looks like:
// Handler (HTTP) → Service (business logic) → Repository (DB access)
//
// Returning the interface (NotesService) instead of the concrete type (service)
// hides implementation details and allows easy swapping or mocking in tests.
func NewNotesService(r NotesRepository) NotesService {
	return &service{repo: r}
}

func (s *service) CreateNote(ctx context.Context, n *Note) (*Note, error) {

	if n.AuthorID == "" {
		return nil, fmt.Errorf("missing authour id")
	}

	if n.Content == "" {
		return nil, fmt.Errorf("missing content")
	}

	if n.Title == "" {
		return nil, fmt.Errorf("missing title")
	}

	now := time.Now()

	n.CreatedAt = now
	n.UpdatedAt = now

	if n.Public {
		slug := slugify(n.Title)
		n.Slug = &slug
	}

	createdNote, err := s.repo.CreateNote(ctx, n)

	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return createdNote, nil

}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func (s *service) GetUserNotes(ctx context.Context, userID string) ([]*NoteSummary, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	notes, err := s.repo.GetNotesByAuthor(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notes: %w", err)
	}

	return notes, nil
}

func (s *service) GetUserNote(ctx context.Context, noteID, userID string) (*Note, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	note, err := s.repo.GetNoteByID(ctx, noteID, userID)

	if err != nil {
		return nil, fmt.Errorf("error while fetching the note by id %w", err)
	}

	return note, nil
}

func (s *service) DeleteNote(ctx context.Context, noteID, userID string) error {
	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	err := s.repo.DeleteNote(ctx, noteID, userID)

	if err != nil {
		return fmt.Errorf("error while deleting the note by id %w", err)
	}

	return nil
}
