package main

import (
	"context"
	"flag"
	"log"

	"datingapp/internal/config"
	"datingapp/internal/seeder"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	count := flag.Int("count", 1000, "number of users to seed")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DBConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	s := seeder.NewSeeder(db)
	if err := s.SeedUsers(context.Background(), *count); err != nil {
		log.Fatalf("failed to seed users: %v", err)
	}
}
