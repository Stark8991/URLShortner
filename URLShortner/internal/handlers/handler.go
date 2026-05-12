package handlers

import (
	"github.com/Stark8991/URLShortner/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db      *pgxpool.Pool
	queries *database.Queries
}

func New(db *pgxpool.Pool, queries *database.Queries) *Handler {
	return &Handler{
		db:      db,
		queries: queries,
	}
}
