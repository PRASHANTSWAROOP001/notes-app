package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/PRASHANTSWAROOP001/notes-app/internal/middleware"
	"github.com/PRASHANTSWAROOP001/notes-app/internal/notes"
	"github.com/PRASHANTSWAROOP001/notes-app/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("DB connect failed:", err)
	}
	defer db.Close()

	repo := user.NewPostgresUserRepository(db)
	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	notesRepo := notes.NewPostgresNotesRepository(db)

	notesSvc := notes.NewNotesService(notesRepo)

	notesHandler := notes.NewNotehandler(notesSvc)

	http.HandleFunc("/auth/register", h.Register)
	http.HandleFunc("/auth/login", h.Login)

	http.Handle("/notes/create-note", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.CreateNote)))
	http.Handle("/notes/get-notes", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.GetUserNotes)))
	http.Handle("/notes/get-note", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.GetUserNoteById)))
	http.Handle("/notes/delete", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.DeleteNote)))
    http.Handle("/notes/update", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.UpdateNote)))
	http.Handle("/notes/share-slug", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.ShareWithEmail)))
	http.Handle("/notes", middleware.OptionalMiddleware(http.HandlerFunc(notesHandler.GetPublicAccess)))
	http.Handle("/notes/revoke-access", middleware.AuthMiddleware(http.HandlerFunc(notesHandler.RemoveEmailShare)))
	http.HandleFunc("/notes/public", notesHandler.GetPublicAccess)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("🚀 Running on :8080")
	http.ListenAndServe(":8080", nil)
}
