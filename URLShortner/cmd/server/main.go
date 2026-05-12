package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Stark8991/URLShortner/internal/config"
	"github.com/Stark8991/URLShortner/internal/database"
	"github.com/Stark8991/URLShortner/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DB_URL)
	if err != nil {
		log.Fatal("failed to create database pool: ", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	handler := handlers.New(db, database.New(db))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /shorten", handler.CreateShortURL)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server listening on port", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("server failed: ", err)
	}
}
