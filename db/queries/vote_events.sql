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

-- name: GetUserMovieDraftCounts :many
SELECT
  movie_id,
  movie_name,
  COALESCE(SUM(CASE WHEN action = 'selected' THEN 1 ELSE -1 END),0) as net_count
FROM vote_events
WHERE user_id = ? AND event_type = 'draft_toggle'
GROUP BY movie_id, movie_name
HAVING net_count > 0
ORDER BY net_count DESC;

-- name: GetUserMovieVoteCounts :many
SELECT
  movie_id,
  movie_name,
  COALESCE(SUM(CASE WHEN action = 'selected' THEN 1 ELSE -1 END),0) as net_count
FROM vote_events
WHERE user_id = ? AND event_type = 'vote_toggle'
GROUP BY movie_id, movie_name
HAVING net_count > 0
ORDER BY net_count DESC;
