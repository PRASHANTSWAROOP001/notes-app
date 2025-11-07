package notes

import (
	"encoding/json"
	"fmt"
	"log"
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

func (h *NoteHandler) GetUserNotes(w http.ResponseWriter, r *http.Request) {
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

func (h *NoteHandler) GetUserNoteById(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "unauthorized request", http.StatusUnauthorized)
		return
	}

	noteId := r.URL.Query().Get("id")

	if noteId == " " {
		http.Error(w, "missing note id param", http.StatusBadRequest)
		return
	}

	note, err := h.service.GetUserNote(r.Context(), noteId, userId)

	if err != nil {
		http.Error(w, "error while getting note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(note)

}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	noteID := r.URL.Query().Get("id")

	if noteID == " " {
		http.Error(w, "missing note id", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteNote(r.Context(), noteID, userId)

	if err != nil {
		http.Error(w, fmt.Sprintf("error while deleting: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "note deleted successfully",
	})

}

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Public  bool   `json:"public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	note := &Note{
		ID:       req.ID,
		Title:    req.Title,
		Content:  req.Content,
		Public:   req.Public,
		AuthorID: userId,
	}

	noteSummary, err := h.service.UpdateNote(r.Context(), note)
	if err != nil {
		http.Error(w, fmt.Sprintf("error %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(noteSummary)

}

func (h *NoteHandler) ShareWithEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	err := h.service.ShareNoteViaEmail(r.Context(), req.ID, userId, req.Email)

	if err != nil {
		http.Error(w, fmt.Sprintf("error while adding email %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "note shared successfully",
	})

}

func (h *NoteHandler) RemoveEmailShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok || userId == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	email := r.URL.Query().Get("email")

	noteid := r.URL.Query().Get("id")

	if email == "" || noteid == "" {
		http.Error(w, "missing params in query", http.StatusBadRequest)
		return
	}

	err := h.service.RevokeEmailAccess(r.Context(), noteid, userId, email)

	if err != nil {
		http.Error(w, fmt.Sprintf("error while adding email %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "note access removed successfully",
	})

}

func (h *NoteHandler) GetPublicAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := r.URL.Query().Get("q")

	userID, _ := middleware.GetUserID(r.Context())
	userEmail, _ := middleware.GetEmail(r.Context())
    log.Print(slug)
	log.Print(userID)
	log.Print(userEmail)

	var (
		note *Note
		err  error
	)

	if userID == "" {
		note, err = h.service.GetPublicNote(r.Context(), slug, nil, nil)
	} else {
		
		note, err = h.service.GetPublicNote(r.Context(), slug, &userID, &userEmail)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// ✅ Now you can return the note
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}
