-- name: CreateVoteEvent :one
INSERT INTO vote_events (
    user_id,
    event_type,
    action,
    movie_id,
    movie_name
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetVoteEventsByUser :many
SELECT * FROM vote_events
WHERE user_id = ?
ORDER BY created_at DESC;
