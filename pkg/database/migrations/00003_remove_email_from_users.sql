-- +goose Up
-- +goose StatementBegin
-- SQLite doesn't support dropping columns with constraints, so recreate the table
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table
INSERT INTO users_new (id, username, password_hash, created_at, updated_at)
SELECT id, username, password_hash, created_at, updated_at FROM users;

-- Drop old table and indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE users;

-- Rename new table
ALTER TABLE users_new RENAME TO users;

-- Recreate username index
CREATE INDEX idx_users_username ON users (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate table with email column
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data back (email will be empty string)
INSERT INTO users_new (id, username, email, password_hash, created_at, updated_at)
SELECT id, username, '', password_hash, created_at, updated_at FROM users;

-- Drop old table and index
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE users;

-- Rename new table
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email ON users (email);
-- +goose StatementEnd
