-- URL Shortener SQL Queries

-- name: CreateURL :one
INSERT INTO urls (short_code, original_url, expires_at)
VALUES ($1, $2, $3)
RETURNING id, short_code, original_url, created_at, expires_at, click_count;

-- name: GetURLByShortCode :one
SELECT id, short_code, original_url, created_at, expires_at, click_count
FROM urls
WHERE short_code = $1
AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);

-- name: IncrementClickCount :exec
UPDATE urls
SET click_count = click_count + 1
WHERE short_code = $1;

-- name: DeleteExpiredURLs :exec
DELETE FROM urls
WHERE expires_at IS NOT NULL
AND expires_at <= CURRENT_TIMESTAMP;

-- name: GetURLStats :one
SELECT id, short_code, original_url, created_at, expires_at, click_count
FROM urls
WHERE short_code = $1;