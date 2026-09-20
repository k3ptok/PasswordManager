-- +goose Up
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value BLOB NOT NULL
);

-- +goose down
DROP TABLE config;