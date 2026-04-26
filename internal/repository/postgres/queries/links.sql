-- name: GetLinkByOriginalURL :one
SELECT original_url, short_code
FROM links
WHERE original_url = $1;

-- name: GetLinkByShortCode :one
SELECT original_url, short_code
FROM links
WHERE short_code = $1;

-- name: CreateLink :one
INSERT INTO links (original_url, short_code)
VALUES ($1, $2)
RETURNING original_url, short_code;