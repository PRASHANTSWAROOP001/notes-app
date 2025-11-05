package notes

import (
	"encoding/json"
	"net/http"

	"github.com/PRASHANTSWAROOP001/notes-app/internal/middleware"
)

// NoteHandler is the top-level HTTP handler for the Notes feature.
// It does NOT directly depend on the database or repository.
// Instead, it depends on the NotesService interface —
// meaning it only needs to know how to call business logic,
// not how data is stored or fetched.
//
// This is Dependency Injection at the final layer:
// The NotesService (already bound to a repository internally)
// is "injected" into the handler when it's created.
//
// Why? → To keep the handler focused on HTTP (parsing JSON, returning responses)
// and leave business logic to the service layer.
type NoteHandler struct {
	service NotesService
}

// NewNotehandler is a public constructor function that returns a pointer to NoteHandler.
//
// Dependency Injection happens here:
// We pass in a NotesService (which itself has a repository inside),
// so this handler can access all service methods like CreateNote or GetUserNotes.
//
// The handler itself has no dependencies of its own (no DB, no repo),
// it simply acts as a bridge between HTTP requests and service calls.
//
// The flow looks like:
// HTTP Request → Auth Middleware → NoteHandler → NotesService → NotesRepository → Database
func NewNotehandler(svc NotesService) *NoteHandler {
	return &NoteHandler{service: svc}
}


func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Public  bool   `json:"public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	note := &Note{
		AuthorID: userId,
		Title:    req.Title,
		Content:  req.Content,
		Public:   req.Public,
	}

	createdNote, err := h.service.CreateNote(r.Context(), note)

	if err != nil {
		http.Error(w, "could not create the note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdNote)

}

func (h *NoteHandler) GetUserNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	notesData, err := h.service.GetUserNotes(r.Context(), userId)

	if err != nil {
		http.Error(w, "error while fetching users data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(notesData)

}
