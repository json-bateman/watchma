-- name: CreateSession :exec
INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP);

-- name: GetUserIDByToken :one
SELECT user_id FROM refresh_tokens
WHERE token = ?
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM refresh_tokens
WHERE token = ?;
