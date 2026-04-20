package main

import (
	"log"
	"net/http"
	"os"

	"encanto/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Connect to PostgreSQL
	db, err := api.OpenDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\nSet DATABASE_URL env var or ensure postgres is running on localhost:5432/encanto", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	store := api.NewPGStore(db)
	server := api.NewServer(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:4173",
			"http://127.0.0.1:4173",
			"http://localhost:4174",
			"http://127.0.0.1:4174",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Mount("/api", server.Router())
	r.Get("/ws", server.HandleWS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
