package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encanto/api"
	"encanto/cache"
	"encanto/config"
	"encanto/core"
	"encanto/data"
	"encanto/shared"
)

func main() {
	command := flag.String("cmd", "serve", "one of: serve, migrate, seed, reset-seed")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := data.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	catalog, err := core.NewPermissionCatalog()
	if err != nil {
		log.Fatalf("load permission catalog: %v", err)
	}
	rawCatalog, err := shared.LoadPermissionCatalog()
	if err != nil {
		log.Fatalf("load raw permission catalog: %v", err)
	}

	if err := data.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	switch *command {
	case "migrate":
		log.Printf("migrations applied successfully")
		return
	case "seed":
		if err := data.SeedDevData(ctx, pool, rawCatalog); err != nil {
			log.Fatalf("seed database: %v", err)
		}
		log.Printf("seed data applied successfully")
		return
	case "reset-seed":
		if err := data.ResetAndSeed(ctx, pool, rawCatalog); err != nil {
			log.Fatalf("reset seed database: %v", err)
		}
		log.Printf("seed data reset successfully")
		return
	case "serve":
	default:
		log.Fatalf("unknown command %q", *command)
	}

	if cfg.AutoSeed {
		if err := data.SeedDevData(ctx, pool, rawCatalog); err != nil {
			log.Fatalf("seed database: %v", err)
		}
	}

	redisClient, err := cache.New(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	store := data.NewStore(pool)
	sessionManager := core.NewSessionManager(cfg, redisClient)
	accessService := core.NewAccessService(store, catalog)
	chatService := core.NewChatService(store, accessService)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: api.Router(api.Dependencies{
			Config:         cfg,
			Store:          store,
			SessionManager: sessionManager,
			AccessService:  accessService,
			ChatService:    chatService,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Encanto backend listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve http: %v", err)
		}
	}()

	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	fmt.Println("server stopped")
}
