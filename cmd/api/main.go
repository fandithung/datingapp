package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"datingapp/internal/config"
	"datingapp/internal/server"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DBConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	srv := server.NewServer(db, *cfg)
	go func() {
		log.Printf("starting server on port %s", cfg.Port)
		if err := srv.Start(cfg.Port); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited properly")
}
