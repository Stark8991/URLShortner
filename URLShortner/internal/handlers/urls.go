package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Stark8991/URLShortner/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const shortCodeLength = 8

var shortCodeAlphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type createShortURLRequest struct {
	OriginalURL string  `json:"original_url"`
	ExpiresAt   *string `json:"expires_at"`
}

type urlResponse struct {
	ID          int32   `json:"id"`
	ShortCode   string  `json:"short_code"`
	OriginalURL string  `json:"original_url"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	ClickCount  int32   `json:"click_count"`
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req createShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	if !isValidURL(req.OriginalURL) {
		writeError(w, http.StatusBadRequest, "original_url must be a valid http or https URL")
		return
	}

	expiresAt, err := parseOptionalTime(req.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "expires_at must be a valid RFC3339 timestamp")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	createdURL, err := h.createShortURL(ctx, req.OriginalURL, expiresAt)
	if err != nil {
		log.Println("failed to create short URL: ", err)
		writeError(w, http.StatusInternalServerError, "failed to create short URL")
		return
	}

	writeJSON(w, http.StatusCreated, toURLResponse(createdURL))
}

func (h *Handler) createShortURL(ctx context.Context, originalURL string, expiresAt pgtype.Timestamptz) (database.Url, error) {
	var lastErr error

	for range 5 {
		shortCode, err := generateShortCode(shortCodeLength)
		if err != nil {
			return database.Url{}, err
		}

		createdURL, err := h.queries.CreateURL(ctx, database.CreateURLParams{
			ShortCode:   shortCode,
			OriginalUrl: originalURL,
			ExpiresAt:   expiresAt,
		})
		if err == nil {
			return createdURL, nil
		}

		if !isUniqueViolation(err) {
			return database.Url{}, err
		}

		lastErr = err
	}

	return database.Url{}, lastErr
}

func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i, b := range randomBytes {
		code[i] = shortCodeAlphabet[int(b)%len(shortCodeAlphabet)]
	}

	return string(code), nil
}

func parseOptionalTime(value *string) (pgtype.Timestamptz, error) {
	if value == nil || *value == "" {
		return pgtype.Timestamptz{Valid: false}, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return pgtype.Timestamptz{}, err
	}

	return pgtype.Timestamptz{
		Time:  parsedTime,
		Valid: true,
	}, nil
}

func isValidURL(value string) bool {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toURLResponse(value database.Url) urlResponse {
	createdAt := value.CreatedAt.Time.Format(time.RFC3339)

	var expiresAt *string
	if value.ExpiresAt.Valid {
		formattedExpiresAt := value.ExpiresAt.Time.Format(time.RFC3339)
		expiresAt = &formattedExpiresAt
	}

	return urlResponse{
		ID:          value.ID,
		ShortCode:   value.ShortCode,
		OriginalURL: value.OriginalUrl,
		CreatedAt:   createdAt,
		ExpiresAt:   expiresAt,
		ClickCount:  value.ClickCount.Int32,
	}
}
