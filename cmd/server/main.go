package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/PRASHANTSWAROOP001/notes-app/internal/user"
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

	http.HandleFunc("/auth/register", h.Register)
	http.HandleFunc("/auth/login", h.Login)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("🚀 Running on :8080")
	http.ListenAndServe(":8080", nil)
}
