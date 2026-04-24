package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encanto/api"
	"encanto/config"
	"encanto/workers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := api.OpenDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\nSet DATABASE_URL env var or ensure postgres is running on localhost:5432/encanto", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	store := api.NewPGStore(db)
	server := api.NewServer(store)

	workerPool := workers.New(store, server.WhatsApp)
	workerPool.SetPollInterval(cfg.WorkerPollInterval)

	// Load existing WhatsApp sessions
	if err := server.WhatsApp.LoadSessions(); err != nil {
		log.Printf("WARNING: failed to load WhatsApp sessions: %v", err)
	}

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

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	rootCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		workerPool.Start(workerCtx)
	}()

	serverErrCh := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	log.Printf("🚀 Server listening on port %s", port)

	select {
	case <-rootCtx.Done():
	case err := <-serverErrCh:
		if err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("WARNING: HTTP shutdown error: %v", err)
	}
	cancel()

	workerPool.Stop()
	workerCancel()
	<-workerDone
}
