-- +goose Up
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    website_url TEXT NOT NULL,
    username TEXT NOT NULL,
    encrypted_password BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_service_user ON entries(website_url, username);

-- +goose Down
DROP TABLE entries;