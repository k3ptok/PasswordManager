-- name: CreateEntry :one
INSERT INTO entries (website_url, username, encrypted_password)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE website_url = ? AND username = ?;

-- name: ListEntries :many
SELECT website_url, username FROM entries
ORDER BY website_url;

-- name: UpdateEntry :exec
UPDATE entries
SET encrypted_password = ?
WHERE website_url = ? AND username = ?;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE website_url = ? AND username = ?;

-- name: SearchEntries :many
SELECT website_url, username FROM entries
WHERE website_url LIKE ?
ORDER BY website_url;
-- name: ImportEntry :exec
INSERT INTO entries (website_url, username, encrypted_password)
VALUES (?, ?, ?)
ON CONFLICT(website_url, username) DO NOTHING;


-- name: SetConfig :exec
INSERT INTO config (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;

-- name: GetConfig :one
SELECT value FROM config WHERE key = ?;