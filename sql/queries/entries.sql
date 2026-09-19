-- name: CreateEntry :one
INSERT INTO entries (website_url, username, encryped_password)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE website_url = ? AND username = ?;

-- name: ListEntries :many
SELECT id, website_url, username FROM entries
ORDER BY website_url;