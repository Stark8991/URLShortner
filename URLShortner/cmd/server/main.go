package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Stark8991/URLShortner/internal/config"
	"github.com/Stark8991/URLShortner/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type app struct {
	db      *pgxpool.Pool
	queries *database.Queries
}

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

	app := &app{
		db:      db,
		queries: database.New(db),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.healthHandler)

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

func (a *app) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := a.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "database unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("failed to write JSON response: ", err)
	}
}
