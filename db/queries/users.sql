-- name: CreateUser :one
INSERT INTO users (username, password_hash)
VALUES (?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ?
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ?
LIMIT 1;

-- name: GetUserBySessionToken :one
SELECT u.* FROM users u
INNER JOIN refresh_tokens rt ON u.id = rt.user_id
WHERE rt.token = ?
LIMIT 1;
